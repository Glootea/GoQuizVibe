package store

import (
	"context"

	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	Close()
	CreateUser(ctx context.Context, u *db.User) error
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*db.User, error)
	GetQuizzes(ctx context.Context) ([]*QuizWithQuestions, error)
	GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*QuizWithQuestions, error)
	GetQuizByID(ctx context.Context, id uuid.UUID) (*QuizWithQuestions, error)
	GetQuizWithQuestions(ctx context.Context, id uuid.UUID) (*QuizWithQuestions, error)
	SaveAttempt(ctx context.Context, attempt *db.QuizAttempt) error
	GetAttemptsByUser(ctx context.Context, userID uuid.UUID) ([]*db.QuizAttempt, error)
	SaveUserAnswer(ctx context.Context, answer *db.UserAnswer) error
	GetUserAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*db.UserAnswer, error)
	GetLeaderboard(ctx context.Context, limit int32) ([]*models.LeaderboardEntry, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error)
}