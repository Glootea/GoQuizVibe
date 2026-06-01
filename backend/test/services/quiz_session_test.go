package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	gamificationSvc "github.com/goquizvibe/backend/feature/gamification/services"
	quizSvc "github.com/goquizvibe/backend/feature/quiz/services"
	"github.com/goquizvibe/backend/shared/db"
	cacheSvc "github.com/goquizvibe/backend/shared/infrastructure/cache"
)

func TestQuizSessionService_NewQuizSessionService(t *testing.T) {
	t.Parallel()
	m := NewQuizSessionServiceMocks(t)

	gamification := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
	cacheService := &cacheSvc.CacheService{}
	svc := quizSvc.NewQuizSessionService(m.Attempts, m.Quizzes, m.Questions, m.Images, m.Users, gamification, cacheService)

	if svc == nil {
		t.Error("NewQuizSessionService() returned nil")
	}
}

func TestQuizSessionService_GetUserStats(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	statsRow := db.GetUserStatsRow{
		TotalXp:    int64(500),
		CorrectCnt: 45,
		WrongCnt:   10,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		m := NewQuizSessionServiceMocks(t)

		gamification := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(now))
		cacheService := &cacheSvc.CacheService{}

		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		svc := quizSvc.NewQuizSessionService(m.Attempts, m.Quizzes, m.Questions, m.Images, m.Users, gamification, cacheService)
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats == nil {
			t.Fatal("GetUserStats() returned nil")
		}
		if stats.XP != 500 {
			t.Errorf("GetUserStats() XP = %d, want 500", stats.XP)
		}
	})

	t.Run("get stats error", func(t *testing.T) {
		t.Parallel()
		m := NewQuizSessionServiceMocks(t)

		gamification := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(now))
		cacheService := &cacheSvc.CacheService{}

		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := quizSvc.NewQuizSessionService(m.Attempts, m.Quizzes, m.Questions, m.Images, m.Users, gamification, cacheService)
		_, err := svc.GetUserStats(ctx, userID)
		if err == nil {
			t.Fatal("GetUserStats() error = nil, want error")
		}
	})
}

func TestQuizSessionService_GetLeaderboardData(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		m := NewQuizSessionServiceMocks(t)

		gamification := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(now))
		cacheService := &cacheSvc.CacheService{}

		m.Attempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return([]db.GetRecentAttemptsRow{
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 100},
		}, nil)
		m.Users.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)

		svc := quizSvc.NewQuizSessionService(m.Attempts, m.Quizzes, m.Questions, m.Images, m.Users, gamification, cacheService)
		result, err := svc.GetLeaderboardData(ctx, userID)
		if err != nil {
			t.Fatalf("GetLeaderboardData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetLeaderboardData() returned nil")
		}
	})

	t.Run("get leaderboard error", func(t *testing.T) {
		t.Parallel()
		m := NewQuizSessionServiceMocks(t)

		gamification := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(now))
		cacheService := &cacheSvc.CacheService{}

		m.Attempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(nil, errors.New("leaderboard error"))
		m.Users.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, nil)

		svc := quizSvc.NewQuizSessionService(m.Attempts, m.Quizzes, m.Questions, m.Images, m.Users, gamification, cacheService)
		_, err := svc.GetLeaderboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetLeaderboardData() error = nil, want error")
		}
	})
}
