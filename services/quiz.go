package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
)

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

type QuizService struct {
	pool *db.Queries
}

func NewQuizService(pool *db.Queries) *QuizService {
	return &QuizService{pool: pool}
}

func (s *QuizService) GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*models.Quiz, error) {
	quizzes, err := s.pool.GetQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Quiz, len(quizzes))
	for i, q := range quizzes {
		result[i] = &models.Quiz{
			Quiz:      q,
			Questions: nil,
		}
	}
	return result, err
}

func (s *QuizService) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	quiz, err := s.pool.GetQuizByID(ctx, id)
	if err != nil {
		return nil, err
	}
	questions, err := s.pool.GetQuestionsByQuizID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.Quiz{
		Quiz:      quiz,
		Questions: questions,
	}, nil
}

func (s *QuizService) SubmitQuizAttempt(ctx context.Context, userID, quizID uuid.UUID, answers map[uuid.UUID]string) (*db.QuizAttempt, error) {
	_, err := s.pool.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	questions, err := s.pool.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	attemptID := uuid.New()
	now := time.Now()

	if _, err := s.pool.CreateAttempt(ctx, db.CreateAttemptParams{
		ID:        attemptID,
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: now,
	}); err != nil {
		return nil, err
	}

	var score, maxScore int
	var userAnswers []db.UserAnswer

	for _, q := range questions {
		maxScore += int(q.Points)
		userAnswer := answers[q.ID]
		isCorrect := NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		userAnswers = append(userAnswers, db.UserAnswer{
			ID:         uuid.New(),
			AttemptID:  attemptID,
			QuestionID: q.ID,
			UserAnswer: userAnswer,
			IsCorrect:  isCorrect,
		})

		if isCorrect {
			score += int(q.Points)
		}
	}

	completedAt := time.Now()
	attempt, err := s.pool.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          attemptID,
		Score:       score,
		MaxScore:    maxScore,
		CompletedAt: completedAt,
	})
	if err != nil {
		return nil, err
	}

	for _, a := range userAnswers {
		if _, err := s.pool.CreateUserAnswer(ctx, db.CreateUserAnswerParams{
			ID:         a.ID,
			AttemptID:  a.AttemptID,
			QuestionID: a.QuestionID,
			UserAnswer: a.UserAnswer,
			IsCorrect:  a.IsCorrect,
		}); err != nil {
			return nil, err
		}
	}

	return &attempt, nil
}