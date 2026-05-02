package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"
)

type QuizHandler struct {
	pool         *db.Queries
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewQuiz(pool *db.Queries, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *QuizHandler {
	return &QuizHandler{
		pool:         pool,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (h *QuizHandler) QuizPage(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	ctx := r.Context()
	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	user, err := h.pool.GetUserByID(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}
	stats, _ := h.gamification.GetUserStats(ctx, userID)

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		session := h.createSession(ctx, userID, quizID)
		sessionID = session.ID.String()
	}

	data := types.QuizPageData{
		User:      &user,
		Quiz:      quiz,
		Stats:     stats,
		SessionID: sessionID,
	}

	return pages.QuizPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) createSession(ctx context.Context, userID, quizID uuid.UUID) *db.QuizSession {
	attemptID := uuid.New()
	now := time.Now()
	h.pool.CreateAttempt(ctx, db.CreateAttemptParams{
		ID:        attemptID,
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: now,
	})

	session := &db.QuizSession{
		ID:           uuid.New(),
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attemptID,
		CurrentIndex: 0,
		Answers:      nil,
		CreatedAt:    now,
	}
	_, _ = h.pool.CreateSession(ctx, db.CreateSessionParams{
		ID:           session.ID,
		UserID:       session.UserID,
		QuizID:       session.QuizID,
		AttemptID:    session.AttemptID,
		CurrentIndex: session.CurrentIndex,
		Answers:      session.Answers,
		CreatedAt:    session.CreatedAt,
	})

	return session
}

func (h *QuizHandler) QuizSubmitHTMX(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	quizID, err := extractQuizIDFromPath(r.URL.Path)
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
	questionIndex, _ := strconv.Atoi(questionIndexStr)

	ctx := r.Context()
	userID, err := h.getUserIDFromRequest(r)
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

	session, err := h.pool.GetSession(ctx, sessionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	question := quiz.Questions[questionIndex]
	isCorrect := NormalizeAnswer(answer) == NormalizeAnswer(question.CorrectAnswer)

	var answers map[int]string
	if session.Answers != nil {
		json.Unmarshal(session.Answers, &answers)
	} else {
		answers = make(map[int]string)
	}
	answers[questionIndex] = answer

	answersJSON, _ := json.Marshal(answers)
	h.pool.UpdateSession(ctx, db.UpdateSessionParams{
		ID:           session.ID,
		CurrentIndex: questionIndex + 1,
		Answers:      answersJSON,
	})

	h.pool.CreateUserAnswer(ctx, db.CreateUserAnswerParams{
		ID:         uuid.New(),
		AttemptID:  session.AttemptID,
		QuestionID: question.ID,
		UserAnswer: answer,
		IsCorrect:  isCorrect,
	})

	if isCorrect {
		h.gamification.AwardXP(ctx, userID, 10)
	}

	feedback := &pages.QuestionFeedbackData{
		IsCorrect:     isCorrect,
		CorrectAnswer: question.CorrectAnswer,
		Explanation:   getExplanation(question),
		IsLast:        questionIndex >= len(quiz.Questions)-1,
	}

	if isCorrect {
		return pages.QuestionWithFeedback(quiz, questionIndex, sessionID.String(), feedback).Render(r.Context(), w)
	}
	return pages.QuestionWithWrongFeedback(quiz, questionIndex, sessionID.String(), feedback).Render(r.Context(), w)
}

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

func (h *QuizHandler) QuizNextHTMX(w http.ResponseWriter, r *http.Request) error {
	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("session ID missing from query params")), http.StatusBadRequest)
	}

	indexStr := r.URL.Query().Get("index")
	index, _ := strconv.Atoi(indexStr)

	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	_, err = h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	ctx := r.Context()
	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if index >= len(quiz.Questions) {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
		return nil
	}

	return pages.QuestionCard(quiz, index, sessionIDStr).Render(r.Context(), w)
}

func (h *QuizHandler) QuizResult(w http.ResponseWriter, r *http.Request) error {
	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("session ID missing from query params")), http.StatusBadRequest)
	}
	sessionID, _ := uuid.Parse(sessionIDStr)

	ctx := r.Context()
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	user, err := h.pool.GetUserByID(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	stats, _ := h.gamification.GetUserStats(ctx, userID)

	attempt, err := h.completeSession(ctx, sessionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, fmt.Errorf("complete session: %w", err)), http.StatusInternalServerError)
	}

	var correctCount, wrongCount int
	var answerDetails []types.AnswerDetail

	answers, _ := h.pool.GetAnswersByAttempt(ctx, attempt.ID)
	answersMap := make(map[uuid.UUID]string)
	for _, a := range answers {
		answersMap[a.QuestionID] = a.UserAnswer
		if a.IsCorrect {
			correctCount++
		} else {
			wrongCount++
		}
	}

	for _, q := range quiz.Questions {
		userAnswer := answersMap[q.ID]
		if userAnswer == "" {
			userAnswer = "Нет ответа"
		}
		isCorrect := userAnswer != "" && NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		answerDetails = append(answerDetails, types.AnswerDetail{
			Question:      q.Text,
			UserAnswer:    userAnswer,
			CorrectAnswer: q.CorrectAnswer,
			IsCorrect:     isCorrect,
		})
	}

	data := types.QuizResultData{
		User:         &user,
		Quiz:         quiz,
		Stats:        stats,
		Score:        int(attempt.Score),
		MaxScore:     int(attempt.MaxScore),
		CorrectCount: correctCount,
		WrongCount:   wrongCount,
		Answers:      answerDetails,
	}

	return pages.QuizResultPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) completeSession(ctx context.Context, sessionID uuid.UUID) (*db.QuizAttempt, error) {
	session, err := h.pool.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	quiz, err := h.quizService.GetQuizByID(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	answers, _ := h.pool.GetAnswersByAttempt(ctx, session.AttemptID)

	var score, maxScore int
	for _, q := range quiz.Questions {
		maxScore += int(q.Points)
		for _, a := range answers {
			if a.QuestionID == q.ID && a.IsCorrect {
				score += int(q.Points)
				break
			}
		}
	}

	now := time.Now()
	updatedAttempt, err := h.pool.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          session.AttemptID,
		Score:       score,
		MaxScore:    maxScore,
		CompletedAt: now,
	})
	if err != nil {
		return nil, err
	}

	h.pool.DeleteSession(ctx, sessionID)

	return &updatedAttempt, nil
}

func (h *QuizHandler) ErrorsPage(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	user, err := h.pool.GetUserByID(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}
	stats, _ := h.gamification.GetUserStats(ctx, userID)

	attempts, _ := h.pool.GetQuizErrors(ctx, userID)

	var quizErrors []types.QuizErrors
	for _, attempt := range attempts {
		quiz, err := h.quizService.GetQuizByID(ctx, attempt.QuizID)
		if err == nil {
			answers, _ := h.pool.GetAnswersByAttempt(ctx, attempt.ID)
			questionMap := make(map[uuid.UUID]db.Question)
			for _, q := range quiz.Questions {
				questionMap[q.ID] = q
			}
			var wrongAnswers []models.WrongAnswer
			for _, a := range answers {
				if !a.IsCorrect {
					q := questionMap[a.QuestionID]
					wrongAnswers = append(wrongAnswers, models.WrongAnswer{
						ID:            a.ID,
						QuestionID:    a.QuestionID,
						QuizID:        attempt.QuizID,
						UserAnswer:    a.UserAnswer,
						CorrectAnswer: q.CorrectAnswer,
						Explanation:   q.Explanation,
						Timestamp:     attempt.StartedAt,
					})
				}
			}
			quizErrors = append(quizErrors, types.QuizErrors{
				Quiz:         quiz,
				WrongAnswers: wrongAnswers,
			})
		}
	}

	data := types.ErrorsPageData{
		User:       &user,
		QuizErrors: quizErrors,
		Stats:      stats,
	}

	return pages.ErrorsPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) LeaderboardPage(w http.ResponseWriter, r *http.Request) error {
	_, err := h.getUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	ctx := r.Context()
	userID, _ := h.getUserIDFromRequest(r)
	user, _ := h.pool.GetUserByID(ctx, userID)
	entries, _ := h.gamification.GetLeaderboard(ctx, 100)

	data := types.LeaderboardPageData{
		User:    &user,
		Entries: entries,
	}

	return pages.LeaderboardPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
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

func extractQuizIDFromPath(path string) (uuid.UUID, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return uuid.Nil, fmt.Errorf("invalid path format: %s", path)
	}
	return uuid.Parse(parts[1])
}

func getExplanation(q models.Question) string {
	return q.Explanation
}
