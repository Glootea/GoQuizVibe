package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
)

type QuizHandler struct {
	store        *store.MemoryStore
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewQuiz(s *store.MemoryStore, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *QuizHandler {
	return &QuizHandler{
		store:        s,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (h *QuizHandler) QuizPage(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	quiz, err := h.quizService.GetQuizByID(quizID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	user, _ := h.store.GetUserByID(userID)
	stats, _ := h.gamification.GetUserStats(userID)

	data := types.QuizPageData{
		User:  user,
		Quiz:  quiz,
		Stats: stats,
	}

	pages.QuizPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) QuizSubmitHTMX(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuestionIndex int    `json:"question_index"`
		Answer        string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_correct": true,
		"is_last":    true,
	})
}

func (h *QuizHandler) QuizResult(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	quiz, err := h.quizService.GetQuizByID(quizID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	user, _ := h.store.GetUserByID(userID)
	stats, _ := h.gamification.GetUserStats(userID)

	data := types.QuizResultData{
		User:  user,
		Quiz:  quiz,
		Stats: stats,
	}

	pages.QuizResultPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) ErrorsPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	progress, _ := h.gamification.GetUserStats(userID)
	user, _ := h.store.GetUserByID(userID)

	data := types.ErrorsPageData{
		User:         user,
		WrongAnswers: progress.WrongAnswers,
	}

	pages.ErrorsPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) LeaderboardPage(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.getUserIDFromRequest(r)
	entries := h.gamification.GetLeaderboard()
	user, _ := h.store.GetUserByID(userID)

	data := types.LeaderboardPageData{
		User:    user,
		Entries: entries,
	}

	pages.LeaderboardPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
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
