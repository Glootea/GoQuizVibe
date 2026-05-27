package services_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	dashboardSvc "github.com/goquizvibe/backend/feature/dashboard/services"
	gamificationSvc "github.com/goquizvibe/backend/feature/gamification/services"
	"github.com/goquizvibe/backend/shared/db"
	mocks "github.com/goquizvibe/backend/shared/mocks/services"
	"github.com/goquizvibe/backend/shared/models"
	"go.uber.org/mock/gomock"
)

func TestDashboardService_GetUserIDFromRequest(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(time.Now()))
		auth := NewMockAuthenticator(ctrl)

		userID := uuid.New()
		auth.EXPECT().ValidateToken("valid-token").Return(&models.AuthClaims{UserID: userID}, nil)

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})

		result, err := svc.GetUserIDFromRequest(req)
		if err != nil {
			t.Fatalf("GetUserIDFromRequest() error = %v, want nil", err)
		}
		if result != userID {
			t.Errorf("GetUserIDFromRequest() = %v, want %v", result, userID)
		}
	})

	t.Run("missing cookie", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(time.Now()))
		auth := NewMockAuthenticator(ctrl)

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, err := svc.GetUserIDFromRequest(req)
		if err == nil {
			t.Fatal("GetUserIDFromRequest() error = nil, want error")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(time.Now()))
		auth := NewMockAuthenticator(ctrl)

		auth.EXPECT().ValidateToken("invalid-token").Return(nil, errors.New("invalid token"))

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})

		_, err := svc.GetUserIDFromRequest(req)
		if err == nil {
			t.Fatal("GetUserIDFromRequest() error = nil, want error")
		}
	})
}

func TestDashboardService_NewDashboardService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(time.Now()))
	auth := NewMockAuthenticator(ctrl)

	svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)
	if svc == nil {
		t.Error("NewDashboardService() returned nil")
	}
}

func TestDashboardService_GetDashboardData(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	baseTime := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)
		auth := NewMockAuthenticator(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(baseTime))

		user := db.User{ID: userID, Email: "test@example.com", Name: "Test User"}
		quizzes := []db.Quiz{{ID: uuid.New(), Title: "Quiz 1"}}
		questions := []db.Question{{ID: uuid.New(), QuizID: quizzes[0].ID, Text: "Q1"}}
		images := []db.QuestionImage{{ID: uuid.New(), QuestionID: questions[0].ID}}

		statsRow := db.GetUserStatsRow{TotalXp: int64(100), CorrectCnt: 10, WrongCnt: 2}
		attempts := []db.QuizAttempt{{ID: uuid.New(), UserID: userID, QuizID: quizzes[0].ID, Score: 100, CompletedAt: sql.NullTime{Time: baseTime, Valid: true}}}
		recentAttempts := []db.GetRecentAttemptsRow{{UserID: userID, UserName: "Test User", Score: 100}}

		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return(quizzes, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizzes[0].ID).Return(questions, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, questions[0].ID).Return(images, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(time.Now(), nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)
		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(recentAttempts, nil)

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		result, err := svc.GetDashboardData(ctx, userID)
		if err != nil {
			t.Fatalf("GetDashboardData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetDashboardData() returned nil")
		}
		if result.User.ID != userID {
			t.Errorf("result.User.ID = %v, want %v", result.User.ID, userID)
		}
		if len(result.Quizzes) != 1 {
			t.Errorf("len(result.Quizzes) = %d, want 1", len(result.Quizzes))
		}
	})

	t.Run("get user error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)
		auth := NewMockAuthenticator(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(baseTime))

		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, errors.New("user not found"))

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		_, err := svc.GetDashboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetDashboardData() error = nil, want error")
		}
	})

	t.Run("get quizzes error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)
		auth := NewMockAuthenticator(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(baseTime))

		user := db.User{ID: userID}
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return(nil, errors.New("quizzes error"))

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		_, err := svc.GetDashboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetDashboardData() error = nil, want error")
		}
	})

	t.Run("get user stats error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)
		auth := NewMockAuthenticator(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(baseTime))

		user := db.User{ID: userID}
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return([]db.Quiz{}, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		_, err := svc.GetDashboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetDashboardData() error = nil, want error")
		}
	})

	t.Run("get leaderboard error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)
		auth := NewMockAuthenticator(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(baseTime))

		user := db.User{ID: userID}
		statsRow := db.GetUserStatsRow{TotalXp: int64(100), CorrectCnt: 10, WrongCnt: 2}
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return([]db.Quiz{}, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(time.Now(), nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(nil, errors.New("leaderboard error"))

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		_, err := svc.GetDashboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetDashboardData() error = nil, want error")
		}
	})

	t.Run("quizzes include questions and images", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)
		auth := NewMockAuthenticator(ctrl)

		gamification := gamificationSvc.NewGamificationService(mockAttempts, mockStats, NewMockTimeProvider(baseTime))

		quizID := uuid.New()
		questionID := uuid.New()
		user := db.User{ID: userID}
		quizzes := []db.Quiz{{ID: quizID, Title: "Test Quiz"}}
		questions := []db.Question{{ID: questionID, QuizID: quizID, Text: "Test Question"}}
		images := []db.QuestionImage{{ID: uuid.New(), QuestionID: questionID}}
		statsRow := db.GetUserStatsRow{TotalXp: int64(100), CorrectCnt: 10, WrongCnt: 2}
		attempts := []db.QuizAttempt{}
		recentAttempts := []db.GetRecentAttemptsRow{}

		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return(quizzes, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, questionID).Return(images, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(time.Now(), nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)
		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(recentAttempts, nil)

		svc := dashboardSvc.NewDashboardService(mockUsers, mockQuizzes, mockQuestions, mockImages, gamification, auth, nil, nil)

		result, err := svc.GetDashboardData(ctx, userID)
		if err != nil {
			t.Fatalf("GetDashboardData() error = %v, want nil", err)
		}
		if len(result.Quizzes) != 1 {
			t.Fatalf("len(result.Quizzes) = %d, want 1", len(result.Quizzes))
		}
		if len(result.Quizzes[0].Questions) != 1 {
			t.Errorf("len(result.Quizzes[0].Questions) = %d, want 1", len(result.Quizzes[0].Questions))
		}
		if len(result.Quizzes[0].Questions[0].Images) != 1 {
			t.Errorf("len(result.Quizzes[0].Questions[0].Images) = %d, want 1", len(result.Quizzes[0].Questions[0].Images))
		}
	})
}
