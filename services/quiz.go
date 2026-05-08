package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
)

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

type QuizService struct {
	quizzes   r.QuizRepository
	questions r.QuestionRepository
	images    r.ImageRepository
	attempts  r.AttemptRepository
}

func NewQuizService(quizzes r.QuizRepository, questions r.QuestionRepository, images r.ImageRepository, attempts r.AttemptRepository) *QuizService {
	return &QuizService{
		quizzes:   quizzes,
		questions: questions,
		images:    images,
		attempts:  attempts,
	}
}

func (s *QuizService) GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*models.QuizWithQuestionsAndImages, error) {
	quizzes, err := s.quizzes.GetQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.QuizWithQuestionsAndImages, len(quizzes))
	for i, q := range quizzes {
		questions, _ := s.questions.GetQuestionsByQuizID(ctx, q.ID)
		questionsWithImages := s.attachImagesToQuestions(ctx, questions)
		result[i] = &models.QuizWithQuestionsAndImages{
			Quiz:      q,
			Questions: questionsWithImages,
		}
	}
	return result, err
}

func (s *QuizService) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.QuizWithQuestionsAndImages, error) {
	quiz, err := s.quizzes.GetQuizByID(ctx, id)
	if err != nil {
		return nil, err
	}
	questions, err := s.questions.GetQuestionsByQuizID(ctx, id)
	if err != nil {
		return nil, err
	}
	questionsWithImages := s.attachImagesToQuestions(ctx, questions)
	return &models.QuizWithQuestionsAndImages{
		Quiz:      quiz,
		Questions: questionsWithImages,
	}, nil
}

func (s *QuizService) attachImagesToQuestions(ctx context.Context, questions []db.Question) []models.QuestionWithImages {
	result := make([]models.QuestionWithImages, len(questions))
	for i, q := range questions {
		images, _ := s.images.GetImagesByQuestionID(ctx, q.ID)
		result[i] = models.QuestionWithImages{
			Question: q,
			Images:   images,
		}
	}
	return result
}

func (s *QuizService) SubmitQuizAttempt(ctx context.Context, userID, quizID uuid.UUID, answers map[uuid.UUID]string) (*db.QuizAttempt, error) {
	_, err := s.quizzes.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	questions, err := s.questions.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	attemptID := uuid.New()
	now := time.Now()

	if _, err := s.attempts.CreateAttempt(ctx, db.CreateAttemptParams{
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

	completedAt := sql.NullTime{Time: time.Now(), Valid: true}
	attempt, err := s.attempts.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          attemptID,
		Score:       score,
		MaxScore:    maxScore,
		CompletedAt: completedAt,
	})
	if err != nil {
		return nil, err
	}

	for _, a := range userAnswers {
		if _, err := s.attempts.CreateUserAnswer(ctx, db.CreateUserAnswerParams{
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
