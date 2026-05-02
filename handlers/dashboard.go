package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
)

type DashboardHandler struct {
	repo         *store.Repository
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewDashboard(r *store.Repository, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *DashboardHandler {
	return &DashboardHandler{
		repo:         r,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}
	quizzes, _ := h.quizService.GetQuizzesForUser(userID)
	stats, _ := h.gamification.GetUserStats(userID)
	leaderboard := h.gamification.GetLeaderboard()

	data := types.DashboardData{
		User:        user,
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
