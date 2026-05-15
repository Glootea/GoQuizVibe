package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
	"github.com/goquizvibe/types"
)

type DashboardService struct {
	users          r.UserRepository
	quizzes        r.QuizRepository
	questions      r.QuestionRepository
	images         r.ImageRepository
	gamification   *GamificationService
	auth           Authenticator
	sessionService *QuizSessionService
	cache          *CacheService
}

func NewDashboardService(
	users r.UserRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	gamification *GamificationService,
	auth Authenticator,
	sessionService *QuizSessionService,
	cache *CacheService,
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
	return GetUserIDFromRequest(r, s.auth)
}

func (s *DashboardService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*types.DashboardData, error) {
	userCacheKey := "user:" + userID.String()
	user, err := GetOrFetch(ctx, s.cache, userCacheKey, func() (db.User, error) {
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
	quizzes, err := GetOrFetch(ctx, s.cache, cacheKey, func() ([]db.Quiz, error) {
		return s.quizzes.GetQuizzesForUser(ctx, userID)
	})
	if err != nil {
		return nil, err
	}
	result := make([]*models.QuizWithQuestionsAndImages, len(quizzes))
	for i, q := range quizzes {
		questions, _ := s.questions.GetQuestionsByQuizID(ctx, q.ID)
		questionsWithImages := AttachImagesToQuestions(ctx, questions, s.images)
		result[i] = &models.QuizWithQuestionsAndImages{
			Quiz:      q,
			Questions: questionsWithImages,
		}
	}
	return result, err
}
