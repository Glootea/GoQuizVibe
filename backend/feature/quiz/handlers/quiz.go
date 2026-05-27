package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	authSvc "github.com/goquizvibe/backend/feature/auth/services"
	quizSvc "github.com/goquizvibe/backend/feature/quiz/services"
	quizUI "github.com/goquizvibe/backend/feature/quiz/ui"
	ce "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/locales"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/types"
)

var ErrTimeExpired = errors.New("time expired")

type QuizHandler struct {
	sessionService *quizSvc.QuizSessionService
	quizService    *quizSvc.QuizService
	pool           *db.Queries
	authService    *authSvc.AuthService
}

func NewQuiz(pool *db.Queries, qs *quizSvc.QuizService, ss *quizSvc.QuizSessionService, auth *authSvc.AuthService) *QuizHandler {
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

	t := locales.GetTranslator(r.Context())
	return quizUI.QuizInfoPage(data, t).Render(r.Context(), w)
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

	session, sessionIDStr, needsCookie, conflict := h.getSessionFromRequest(r, quizID)
	if conflict != nil {
		data := &types.SessionConflictData{
			ExistingSessionID: conflict.SessionID,
			ExistingQuizID:    conflict.QuizID,
			ExistingQuizTitle: conflict.QuizTitle,
			CurrentIndex:      conflict.CurrentIndex,
			RequestedQuizID:   quizID,
		}
		t := locales.GetTranslator(r.Context())
		return quizUI.SessionConflictPage(data, t).Render(r.Context(), w)
	}
	if session == nil {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/q/0", http.StatusFound)
		return nil
	}

	pageData, err := h.sessionService.GetQuizQuestionData(ctx, userID, quizID, index)
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

	t := locales.GetTranslator(r.Context())

	isHtmx := middleware.IsHTMXRequest(r)
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

		return quizUI.QuizPage(&data, t).Render(r.Context(), w)
	}

	return quizUI.QuestionCard(pageData, t).Render(r.Context(), w)
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

	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	session, sessionIDStr, needsCookie, conflict := h.getSessionFromRequest(r, quizID)
	if conflict != nil {
		data := &types.SessionConflictData{
			ExistingSessionID: conflict.SessionID,
			ExistingQuizID:    conflict.QuizID,
			ExistingQuizTitle: conflict.QuizTitle,
			CurrentIndex:      conflict.CurrentIndex,
			RequestedQuizID:   quizID,
		}
		t := locales.GetTranslator(r.Context())
		return quizUI.SessionConflictPage(data, t).Render(r.Context(), w)
	}
	if session == nil {
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

	usr, err := h.pool.GetUserByID(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	answer := r.FormValue("answer")

	gapAnswers := ParseGapAnswers(r)
	if len(gapAnswers) > 0 {
		answer = joinGapAnswers(gapAnswers)
	}

	navData, err := h.sessionService.NavigateQuestion(ctx, userID, quizID, currentIndex, targetIndex, answer, &usr)
	if err != nil {
		if errors.Is(err, quizSvc.ErrTimeExpired) {
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

	t := locales.GetTranslator(ctx)

	return quizUI.QuestionCard(navData, t).Render(r.Context(), w)
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

	gapAnswers := ParseGapAnswers(r)
	if len(gapAnswers) > 0 {
		answer = joinGapAnswers(gapAnswers)
	}

	userID, err := h.sessionService.GetUserIDFromRequest(r, h.authService)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	session := h.sessionService.GetSession(r.Context(), sessionID)
	if session == nil {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/q/0", http.StatusFound)
		return nil
	}

	_, err = h.sessionService.SaveAnswer(r.Context(), userID, quizID, currentIndex, answer)
	if err != nil {
		if errors.Is(err, quizSvc.ErrTimeExpired) {
			http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
			return nil
		}
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/quiz/"+quizID.String()+"/result?session="+sessionIDStr)
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
	return nil
}

func (h *QuizHandler) getSessionFromRequest(r *http.Request, quizID uuid.UUID) (*quizSvc.QuizSession, string, bool, *types.ActiveSessionInfo) {
	ctx := r.Context()
	userID, _ := h.sessionService.GetUserIDFromRequest(r, h.authService)

	existing := h.sessionService.GetActiveSessionForUser(ctx, userID)
	if existing != nil && existing.QuizID != quizID {
		return nil, "", false, existing
	}

	if existing != nil {
		return &quizSvc.QuizSession{
			UserID:       userID,
			QuizID:       existing.QuizID,
			AttemptID:    existing.SessionID,
			CurrentIndex: existing.CurrentIndex,
		}, userID.String(), false, nil
	}

	cookie, err := r.Cookie("quiz_session_id")
	if err == nil && cookie.Value != "" {
		cookieUserID, parseErr := uuid.Parse(cookie.Value)
		if parseErr == nil {
			session := h.sessionService.GetSession(ctx, cookieUserID)
			if session != nil && session.QuizID != quizID {
				conflict := &types.ActiveSessionInfo{
					SessionID:    session.UserID,
					QuizID:       session.QuizID,
					CurrentIndex: session.CurrentIndex,
				}
				return nil, "", false, conflict
			}
			if session != nil {
				return session, cookie.Value, false, nil
			}
		}
	}

	sessionIDStr := r.FormValue("session")
	if sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err == nil {
			session := h.sessionService.GetSession(ctx, sessionID)
			if session != nil {
				return session, sessionIDStr, false, nil
			}
		}
	}

	session, err := h.sessionService.CreateSession(r.Context(), userID, quizID)
	if err != nil {
		return nil, "", true, nil
	}

	return session, session.UserID.String(), true, nil
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

	session := h.sessionService.GetSession(ctx, sessionID)
	if session == nil {
		attempt, err := h.pool.GetAttemptByID(ctx, sessionID)
		if err == nil && attempt.CompletedAt.Valid {
			data, err := h.sessionService.GetQuizResultData(ctx, quizID, userID)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     "quiz_session_id",
					Value:    "",
					Path:     "/quiz",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   -1,
				})
				t := locales.GetTranslator(r.Context())
				return quizUI.QuizResultPage(*data, t).Render(r.Context(), w)
			}
		}
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/info", http.StatusFound)
		return nil
	}

	data, err := h.sessionService.GetQuizResultData(ctx, quizID, userID)
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

	t := locales.GetTranslator(r.Context())
	return quizUI.QuizResultPage(*data, t).Render(r.Context(), w)
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

	t := locales.GetTranslator(r.Context())
	return quizUI.ErrorsPage(*data, t).Render(r.Context(), w)
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

	t := locales.GetTranslator(r.Context())
	return quizUI.LeaderboardPage(*data, t).Render(r.Context(), w)
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

	session := h.sessionService.GetSession(ctx, sessionID)
	if session == nil {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/info", http.StatusFound)
		return nil
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

	if middleware.IsHTMXRequest(r) {
		return quizUI.TimerWidget(quiz.ID, sessionIDStr, remainingSeconds).Render(ctx, w)
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

	session := h.sessionService.GetSession(ctx, sessionID)
	if session == nil {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/info", http.StatusFound)
		return nil
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

func ParseGapAnswers(r *http.Request) []string {
	var answers []string
	for i := 0; ; i++ {
		ans := r.FormValue(fmt.Sprintf("gap_%d", i))
		if ans == "" {
			break
		}
		answers = append(answers, ans)
	}
	return answers
}

func joinGapAnswers(answers []string) string {
	var result strings.Builder
	for i, ans := range answers {
		if i > 0 {
			result.WriteString("|")
		}
		result.WriteString(ans)
	}
	return result.String()
}
