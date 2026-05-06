package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/types"
)

type DashboardService struct {
	pool         *db.Queries
	quizService  *QuizService
	gamification *GamificationService
	authService  *AuthService
}

func NewDashboardService(pool *db.Queries, qs *QuizService, gs *GamificationService, as *AuthService) *DashboardService {
	return &DashboardService{
		pool:         pool,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (s *DashboardService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*types.DashboardData, error) {
	user, err := s.pool.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quizzes, err := s.quizService.GetQuizzesForUser(ctx, userID)
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

	return &types.DashboardData{
		User:        &user,
		Quizzes:     quizzes,
		Stats:       stats,
		Leaderboard: leaderboard,
	}, nil
}

func (s *DashboardService) GetUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, fmt.Errorf("get cookie: %w", err)
	}
	claims, err := s.authService.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate token: %w", err)
	}
	return claims.UserID, nil
}
