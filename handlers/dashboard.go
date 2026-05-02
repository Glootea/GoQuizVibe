package handlers

import (
	"net/http"

	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	repo        *store.Repository
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewDashboard(r *store.Repository, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *DashboardHandler {
	return &DashboardHandler{
		repo:        r,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, err := h.repo.GetUserByID(ctx, userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	quizzes, _ := h.quizService.GetQuizzesForUser(ctx, userID)
	stats, _ := h.gamification.GetUserStats(ctx, userID)
	leaderboard := h.gamification.GetLeaderboard(ctx)

	data := types.DashboardData{
		User:        user,
		Quizzes:     quizzes,
		Stats:       stats,
		Leaderboard: leaderboard,
	}

	pages.DashboardPage(data).Render(r.Context(), w)
}

func (h *DashboardHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}