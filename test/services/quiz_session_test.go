package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	mocks "github.com/goquizvibe/mocks/services"
	"github.com/goquizvibe/services"
	"go.uber.org/mock/gomock"
)

func TestQuizSessionService_NewQuizSessionService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	gamification := services.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(time.Now()))
	svc := services.NewQuizSessionService(mockAttempts, mockQuizzes, mockQuestions, mockImages, mockUsers, *gamification, services.CacheService{})

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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(now))

		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockQuizzes, mockQuestions, mockImages, mockUsers, *gamification, services.CacheService{})
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(now))

		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockQuizzes, mockQuestions, mockImages, mockUsers, *gamification, services.CacheService{})
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(now))

		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return([]db.GetRecentAttemptsRow{
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 100},
		}, nil)
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockQuizzes, mockQuestions, mockImages, mockUsers, *gamification, services.CacheService{})
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(now))

		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(nil, errors.New("leaderboard error"))
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockQuizzes, mockQuestions, mockImages, mockUsers, *gamification, services.CacheService{})
		_, err := svc.GetLeaderboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetLeaderboardData() error = nil, want error")
		}
	})
}