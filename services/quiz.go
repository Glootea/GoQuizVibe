package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/store"
)

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

type QuizService struct {
	repo store.RepositoryInterface
}

func NewQuizService(r store.RepositoryInterface) *QuizService {
	return &QuizService{repo: r}
}

func (s *QuizService) GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*models.Quiz, error) {
	quizzes, err := s.repo.GetQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Quiz, len(quizzes))
	for i, q := range quizzes {
		result[i] = &models.Quiz{
			Quiz:      *q.Quiz,
			Questions: q.Questions,
		}
	}
	return result, err
}

func (s *QuizService) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	quiz, err := s.repo.GetQuizByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.Quiz{
		Quiz:      *quiz.Quiz,
		Questions: quiz.Questions,
	}, nil
}

func (s *QuizService) SubmitQuizAttempt(ctx context.Context, userID, quizID uuid.UUID, answers map[uuid.UUID]string) (*db.QuizAttempt, error) {
	quiz, err := s.repo.GetQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, err
	}

	attempt := &db.QuizAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: time.Now(),
	}

	var score, maxScore int
	var userAnswers []db.UserAnswer

	for _, q := range quiz.Questions {
		maxScore += int(q.Points)
		userAnswer := answers[q.ID]
		isCorrect := NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		userAnswers = append(userAnswers, db.UserAnswer{
			ID:         uuid.New(),
			AttemptID:  attempt.ID,
			QuestionID: q.ID,
			UserAnswer: userAnswer,
			IsCorrect:  isCorrect,
		})

		if isCorrect {
			score += int(q.Points)
		}
	}

	attempt.Score = score
	attempt.MaxScore = maxScore
	attempt.CompletedAt = time.Now()

	if err := s.repo.SaveAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	for _, a := range userAnswers {
		if err := s.repo.SaveUserAnswer(ctx, &a); err != nil {
			return nil, err
		}
	}

	return attempt, nil
}

type GamificationService struct {
	repo store.RepositoryInterface
}

func NewGamificationService(r store.RepositoryInterface) *GamificationService {
	return &GamificationService{repo: r}
}

func (s *GamificationService) UpdateStreak(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *GamificationService) AwardXP(ctx context.Context, userID uuid.UUID, amount int) error {
	return nil
}

func (s *GamificationService) GetLeaderboard(ctx context.Context) []*models.LeaderboardEntry {
	entries, _ := s.repo.GetLeaderboard(ctx, 100)
	return entries
}

func (s *GamificationService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	return s.repo.GetUserStats(ctx, userID)
}
