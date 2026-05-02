package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"
)

type DashboardHandler struct {
	pool         *db.Queries
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewDashboard(pool *db.Queries, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *DashboardHandler {
	return &DashboardHandler{
		pool:         pool,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	user, err := h.pool.GetUserByID(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}
	quizzes, _ := h.quizService.GetQuizzesForUser(ctx, userID)
	stats, _ := h.gamification.GetUserStats(ctx, userID)
	leaderboard, _ := h.gamification.GetLeaderboard(ctx, 100)

	data := types.DashboardData{
		User:        &user,
		Quizzes:     quizzes,
		Stats:       stats,
		Leaderboard: leaderboard,
	}

	return pages.DashboardPage(data).Render(r.Context(), w)
}

func (h *DashboardHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, errors.Join(errors.New("get cookie"), err)
	}
	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, errors.Join(errors.New("validate token"), err)
	}
	return claims.UserID, nil
}
