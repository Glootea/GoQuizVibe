package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"
)

type QuizHandler struct {
	sessionService *services.QuizSessionService
	quizService    *services.QuizService
	pool           *db.Queries
	authService    *services.AuthService
}

func NewQuiz(pool *db.Queries, qs *services.QuizService, ss *services.QuizSessionService, auth *services.AuthService) *QuizHandler {
	return &QuizHandler{
		sessionService: ss,
		quizService:    qs,
		pool:           pool,
		authService:    auth,
	}
}

func (h *QuizHandler) QuizStart(w http.ResponseWriter, r *http.Request) error {
	quizIDStr := r.PathValue("id")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}
	http.Redirect(w, r, "/quiz/"+quizID.String()+"/q/0", http.StatusFound)
	return nil
}

func (h *QuizHandler) QuizQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	ctx := r.Context()
	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		session, err := h.sessionService.CreateSession(ctx, userID, quizID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}
		sessionIDStr = session.ID.String()
	} else {
		_, err = uuid.Parse(sessionIDStr)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
		}
	}

	if index >= len(quiz.Questions) {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
		return nil
	}

	isHtmx := r.Header.Get("HX-Request") == "true"
	if !isHtmx {
		user, err := h.pool.GetUserByID(ctx, userID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
		}

		stats, err := h.sessionService.GetUserStats(ctx, userID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}

		data := types.QuizPageData{
			User:      &user,
			Quiz:      quiz,
			Stats:     stats,
			SessionID: sessionIDStr,
			Index:     index,
		}
		t := middleware.GetTranslator(r.Context())
		return pages.QuizPage(data, t).Render(r.Context(), w)
	}

	return pages.QuestionCard(quiz, index, sessionIDStr).Render(r.Context(), w)
}

func (h *QuizHandler) QuizSubmitHTMX(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	sessionIDStr := r.URL.Query().Get("session")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	answer := r.FormValue("answer")
	questionIndexStr := r.FormValue("question_index")
	questionIndex, err := strconv.Atoi(questionIndexStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	ctx := r.Context()
	_, err = h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if questionIndex >= len(quiz.Questions) {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, fmt.Errorf("question index %d out of range", questionIndex)), http.StatusNotFound)
	}

	feedback, err := h.sessionService.SubmitAnswer(ctx, sessionID, quizID, questionIndex, answer)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	feedbackData := &pages.QuestionFeedbackData{
		IsCorrect:     feedback.IsCorrect,
		CorrectAnswer: feedback.CorrectAnswer,
		Explanation:   feedback.Explanation,
		IsLast:        feedback.IsLast,
	}

	if feedback.IsCorrect {
		return pages.QuestionWithFeedback(quiz, questionIndex, sessionIDStr, feedbackData).Render(r.Context(), w)
	}
	return pages.QuestionWithWrongFeedback(quiz, questionIndex, sessionIDStr, feedbackData).Render(r.Context(), w)
}

func (h *QuizHandler) QuizResult(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("session ID missing from query params")), http.StatusBadRequest)
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	ctx := r.Context()
	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.sessionService.GetQuizResultData(ctx, quizID, sessionID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return pages.QuizResultPage(*data, t).Render(r.Context(), w)
}

func (h *QuizHandler) ErrorsPage(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.sessionService.GetErrorsPageData(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return pages.ErrorsPage(*data, t).Render(r.Context(), w)
}

func (h *QuizHandler) LeaderboardPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.sessionService.GetLeaderboardData(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return pages.LeaderboardPage(*data, t).Render(r.Context(), w)
}

func (h *QuizHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return h.sessionService.GetUserIDFromRequest(r, h.authService)
}
