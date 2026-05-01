package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
)

type DashboardHandler struct {
	store        *store.MemoryStore
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewDashboard(s *store.MemoryStore, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *DashboardHandler {
	return &DashboardHandler{
		store:        s,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, err := h.store.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	quizzes := h.quizService.GetQuizzesForUser(userID)
	stats, _ := h.gamification.GetUserStats(userID)
	leaderboard := h.gamification.GetLeaderboard()

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
