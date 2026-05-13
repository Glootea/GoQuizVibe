package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"
)

var ErrTimeExpired = errors.New("time expired")

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
	http.Redirect(w, r, "/quiz/"+quizID.String()+"/info", http.StatusFound)
	return nil
}

func (h *QuizHandler) QuizInfo(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	ctx := r.Context()

	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.sessionService.GetQuizInfoData(ctx, userID, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return pages.QuizInfoPage(data, t).Render(r.Context(), w)
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

	sessionID, sessionIDStr, needsCookie, conflict := h.getSessionFromRequest(r, quizID)
	if conflict != nil {
		data := &types.SessionConflictData{
			ExistingSessionID: conflict.SessionID,
			ExistingQuizID:    conflict.QuizID,
			ExistingQuizTitle: conflict.QuizTitle,
			CurrentIndex:      conflict.CurrentIndex,
			RequestedQuizID:   quizID,
		}
		t := middleware.GetTranslator(r.Context())
		return pages.SessionConflictPage(data, t).Render(r.Context(), w)
	}
	if sessionID == uuid.Nil {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/q/0", http.StatusFound)
		return nil
	}

	pageData, err := h.sessionService.GetQuizQuestionData(ctx, sessionID, quizID, index)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if needsCookie {
		quiz, _ := h.quizService.GetQuizByID(ctx, quizID)
		maxAge := 0
		if quiz != nil {
			maxAge = quiz.TimeLimit
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "quiz_session_id",
			Value:    sessionIDStr,
			Path:     "/quiz",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   maxAge,
		})
	}

	t := middleware.GetTranslator(r.Context())

	isHtmx := r.Header.Get("HX-Request") == "true"
	if !isHtmx {
		user, err := h.pool.GetUserByID(ctx, userID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
		}

		data := types.QuizPageData{
			User:             &user,
			Questions:        pageData.Questions,
			SessionID:        pageData.SessionID,
			CurrentIndex:     pageData.CurrentIndex,
			Answers:          pageData.Answers,
			TotalQuestions:   pageData.TotalQuestions,
			RemainingSeconds: pageData.RemainingSeconds,
			TimeLimitMinutes: pageData.TimeLimitMinutes,
			IsLastQuestion:   pageData.IsLastQuestion,
		}

		return pages.QuizPage(&data, t).Render(r.Context(), w)
	}

	return pages.QuestionCard(pageData, t).Render(r.Context(), w)
}

func (h *QuizHandler) QuizNavigate(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	ctx := r.Context()

	_, err = h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	sessionID, sessionIDStr, needsCookie, conflict := h.getSessionFromRequest(r, quizID)
	if conflict != nil {
		data := &types.SessionConflictData{
			ExistingSessionID: conflict.SessionID,
			ExistingQuizID:    conflict.QuizID,
			ExistingQuizTitle: conflict.QuizTitle,
			CurrentIndex:      conflict.CurrentIndex,
			RequestedQuizID:   quizID,
		}
		t := middleware.GetTranslator(r.Context())
		return pages.SessionConflictPage(data, t).Render(r.Context(), w)
	}
	if sessionID == uuid.Nil {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/q/0", http.StatusFound)
		return nil
	}

	currentIndexStr := r.FormValue("current_index")
	currentIndex, err := strconv.Atoi(currentIndexStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	targetIndexStr := r.FormValue("target_index")
	targetIndex, err := strconv.Atoi(targetIndexStr)

	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}
	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	user, err := h.pool.GetUserByID(ctx, userID)

	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	answer := r.FormValue("answer")

	navData, err := h.sessionService.NavigateQuestion(ctx, sessionID, quizID, currentIndex, targetIndex, answer, &user)
	if err != nil {
		if errors.Is(err, services.ErrTimeExpired) {
			http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
			return nil
		}
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if needsCookie {
		quiz, _ := h.quizService.GetQuizByID(ctx, quizID)
		maxAge := 0
		if quiz != nil {
			maxAge = quiz.TimeLimit
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "quiz_session_id",
			Value:    sessionIDStr,
			Path:     "/quiz",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   maxAge,
		})
	}

	w.Header().Set("HX-Push-Url", fmt.Sprintf("/quiz/%s/q/%d", quizID.String(), navData.CurrentIndex))

	t := middleware.GetTranslator(ctx)

	return pages.QuestionCard(navData, t).Render(r.Context(), w)
}

func (h *QuizHandler) QuizFinish(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("session ID missing")), http.StatusBadRequest)
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	currentIndexStr := r.FormValue("current_index")
	currentIndex, err := strconv.Atoi(currentIndexStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	answer := r.FormValue("answer")

	_, err = h.sessionService.SaveAnswer(r.Context(), sessionID, quizID, currentIndex, answer)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/quiz/"+quizID.String()+"/result?session="+sessionIDStr)
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
	return nil
}

func (h *QuizHandler) getSessionFromRequest(r *http.Request, quizID uuid.UUID) (uuid.UUID, string, bool, *types.ActiveSessionInfo) {
	ctx := r.Context()
	cookie, err := r.Cookie("quiz_session_id")
	if err == nil && cookie.Value != "" {
		sessionID, parseErr := uuid.Parse(cookie.Value)
		if parseErr == nil {
			exists, _ := h.sessionService.SessionExists(ctx, sessionID)
			if exists {
				session, err := h.pool.GetSession(ctx, sessionID)
				if err == nil && session.QuizID != quizID {
					existingQuiz, _ := h.quizService.GetQuizByID(ctx, session.QuizID)
					conflict := &types.ActiveSessionInfo{
						SessionID:    session.ID,
						QuizID:       session.QuizID,
						QuizTitle:    existingQuiz.Title,
						CurrentIndex: session.CurrentIndex,
					}
					return uuid.Nil, "", false, conflict
				}
				return sessionID, cookie.Value, false, nil
			}
		}
	}

	sessionIDStr := r.FormValue("session")
	if sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err == nil {
			return sessionID, sessionIDStr, false, nil
		}
	}

	userID, _ := h.sessionService.GetUserIDFromRequest(r, h.authService)
	session, err := h.sessionService.CreateSession(r.Context(), userID, quizID)
	if err != nil {
		return uuid.Nil, "", true, nil
	}

	return session.ID, session.ID.String(), true, nil
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

	http.SetCookie(w, &http.Cookie{
		Name:     "quiz_session_id",
		Value:    "",
		Path:     "/quiz",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

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

func (h *QuizHandler) SyncTime(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	sessionIDStr := r.URL.Query().Get("session")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	ctx := r.Context()

	session, err := h.pool.GetSession(ctx, sessionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	attempt, err := h.pool.GetAttemptByID(ctx, session.AttemptID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	remainingSeconds := int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit) * time.Second)).Seconds())
	if remainingSeconds < 0 {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
		return nil
	}

	if r.Header.Get("HX-Request") == "true" {
		return pages.TimerWidget(quiz.ID, sessionIDStr, remainingSeconds).Render(ctx, w)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"remaining_seconds": %d}`, remainingSeconds)
	return nil
}

func (h *QuizHandler) CancelSession(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	sessionIDStr := r.FormValue("session_id")
	if sessionIDStr == "" {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("session_id missing")), http.StatusBadRequest)
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

	session, err := h.pool.GetSession(ctx, sessionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if session.UserID != userID {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, errors.New("not session owner")), http.StatusUnauthorized)
	}

	_, err = h.sessionService.CompleteSession(ctx, sessionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "quiz_session_id",
		Value:    "",
		Path:     "/quiz",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/quiz/"+quizID.String()+"/q/0", http.StatusFound)
	return nil
}
