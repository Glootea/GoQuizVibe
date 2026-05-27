package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/models"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	r "github.com/goquizvibe/backend/shared/repositories"
	"github.com/goquizvibe/backend/shared/types"
	gamification "github.com/goquizvibe/backend/feature/gamification/services"
	auth "github.com/goquizvibe/backend/feature/auth/services"
	quiz "github.com/goquizvibe/backend/feature/quiz/services"
	cache "github.com/goquizvibe/backend/shared/infrastructure/cache"
	imgHelpers "github.com/goquizvibe/backend/shared/infrastructure/cache"
)

type DashboardService struct {
	users          r.UserRepository
	quizzes        r.QuizRepository
	questions      r.QuestionRepository
	images         r.ImageRepository
	gamification   *gamification.GamificationService
	auth           interfaces.Authenticator
	sessionService *quiz.QuizSessionService
	cache          *cache.CacheService
}

func NewDashboardService(
	users r.UserRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	gamification *gamification.GamificationService,
	auth interfaces.Authenticator,
	sessionService *quiz.QuizSessionService,
	cache *cache.CacheService,
) *DashboardService {
	return &DashboardService{
		users:          users,
		quizzes:        quizzes,
		questions:      questions,
		images:         images,
		gamification:   gamification,
		auth:           auth,
		sessionService: sessionService,
		cache:          cache,
	}
}

func (s *DashboardService) GetUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return auth.GetUserIDFromRequest(r, s.auth)
}

func (s *DashboardService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*types.DashboardData, error) {
	userCacheKey := "user:" + userID.String()
	user, err := cache.GetOrFetch(ctx, s.cache, userCacheKey, func() (db.User, error) {
		return s.users.GetUserByID(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quizzes, err := s.getQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get quizzes: %w", err)
	}

	stats, err := s.gamification.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	leaderboard, err := s.gamification.GetLeaderboard(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}

	var activeSession *types.ActiveSessionInfo
	if s.sessionService != nil {
		activeSession = s.sessionService.GetActiveSessionForUser(ctx, userID)
	}

	return &types.DashboardData{
		User:          &user,
		Quizzes:       quizzes,
		Stats:         stats,
		Leaderboard:   leaderboard,
		ActiveSession: activeSession,
	}, nil
}

func (s *DashboardService) getQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*models.QuizWithQuestionsAndImages, error) {
	cacheKey := "quizzes:user:" + userID.String()
	quizzes, err := cache.GetOrFetch(ctx, s.cache, cacheKey, func() ([]db.Quiz, error) {
		return s.quizzes.GetQuizzesForUser(ctx, userID)
	})
	if err != nil {
		return nil, err
	}
	result := make([]*models.QuizWithQuestionsAndImages, len(quizzes))
	for i, q := range quizzes {
		questions, _ := s.questions.GetQuestionsByQuizID(ctx, q.ID)
		questionsWithImages := imgHelpers.AttachImagesToQuestions(ctx, questions, s.images)
		result[i] = &models.QuizWithQuestionsAndImages{
			Quiz:      q,
			Questions: questionsWithImages,
		}
	}
	return result, err
}