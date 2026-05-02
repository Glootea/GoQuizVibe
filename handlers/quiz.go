package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
)

type QuizHandler struct {
	repo         *store.Repository
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewQuiz(r *store.Repository, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *QuizHandler {
	return &QuizHandler{
		repo:         r,
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

	ctx := r.Context()
	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	user, err := h.repo.GetUserByID(ctx, userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	stats, _ := h.gamification.GetUserStats(ctx, userID)

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		session := h.createSession(ctx, userID, quizID)
		sessionID = session.ID.String()
	}

	data := types.QuizPageData{
		User:      user,
		Quiz:      quiz,
		Stats:     stats,
		SessionID: sessionID,
	}

	pages.QuizPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) createSession(ctx context.Context, userID, quizID uuid.UUID) *db.QuizSession {
	attempt := &db.QuizAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: time.Now(),
	}
	h.repo.SaveAttempt(ctx, attempt)

	session := &db.QuizSession{
		ID:           uuid.New(),
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attempt.ID,
		CurrentIndex: 0,
		Answers:      nil,
		CreatedAt:    time.Now(),
	}
	h.repo.CreateSession(ctx, session)
	return session
}

func (h *QuizHandler) QuizSubmitHTMX(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	sessionIDStr := r.URL.Query().Get("session")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	answer := r.FormValue("answer")
	questionIndexStr := r.FormValue("question_index")
	questionIndex, _ := strconv.Atoi(questionIndexStr)

	ctx := r.Context()
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	if questionIndex >= len(quiz.Questions) {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	session, err := h.repo.GetSession(ctx, sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
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
	session.Answers = answersJSON
	session.CurrentIndex = questionIndex + 1
	h.repo.UpdateSession(ctx, session)

	userAnswer := &db.UserAnswer{
		ID:         uuid.New(),
		AttemptID:  session.AttemptID,
		QuestionID: question.ID,
		UserAnswer: answer,
		IsCorrect:  isCorrect,
	}
	h.repo.SaveUserAnswer(ctx, userAnswer)

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
		pages.QuestionWithFeedback(quiz, questionIndex, sessionID.String(), feedback).Render(r.Context(), w)
	} else {
		pages.QuestionWithWrongFeedback(quiz, questionIndex, sessionID.String(), feedback).Render(r.Context(), w)
	}
}

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

func getExplanation(q db.Question) string {
	return q.Explanation
}

func (h *QuizHandler) QuizNextHTMX(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, _ := strconv.Atoi(indexStr)

	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	_, err = h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	if index >= len(quiz.Questions) {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionIDStr, http.StatusFound)
		return
	}

	pages.QuestionCard(quiz, index, sessionIDStr).Render(r.Context(), w)
}

func (h *QuizHandler) QuizResult(w http.ResponseWriter, r *http.Request) {
	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	sessionIDStr := r.URL.Query().Get("session")
	if sessionIDStr == "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	sessionID, _ := uuid.Parse(sessionIDStr)

	ctx := r.Context()
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	quiz, err := h.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	user, err := h.repo.GetUserByID(ctx, userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	stats, _ := h.gamification.GetUserStats(ctx, userID)

	attempt, err := h.completeSession(ctx, sessionID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	var correctCount, wrongCount int
	var answerDetails []types.AnswerDetail

	answers, _ := h.repo.GetAnswersByAttempt(ctx, attempt.ID)
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
		User:         user,
		Quiz:         quiz,
		Stats:        stats,
		Score:        int(attempt.Score),
		MaxScore:     int(attempt.MaxScore),
		CorrectCount: correctCount,
		WrongCount:   wrongCount,
		Answers:      answerDetails,
	}

	pages.QuizResultPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) completeSession(ctx context.Context, sessionID uuid.UUID) (*db.QuizAttempt, error) {
	session, err := h.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	attempt, err := h.repo.GetAttemptByID(ctx, session.AttemptID)
	if err != nil {
		return nil, err
	}

	quiz, err := h.quizService.GetQuizByID(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	answers, _ := h.repo.GetAnswersByAttempt(ctx, attempt.ID)

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
	attempt.Score = score
	attempt.MaxScore = maxScore
	attempt.CompletedAt = now
	h.repo.UpdateAttempt(ctx, attempt)

	h.repo.DeleteSession(ctx, sessionID)

	return attempt, nil
}

func (h *QuizHandler) ErrorsPage(w http.ResponseWriter, r *http.Request) {
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
	stats, _ := h.gamification.GetUserStats(ctx, userID)

	attempts, _ := h.repo.GetQuizErrors(ctx, userID)

	var quizErrors []types.QuizErrors
	for _, attempt := range attempts {
		quiz, err := h.quizService.GetQuizByID(ctx, attempt.QuizID)
		if err == nil {
			wrongAnswers, _ := h.repo.GetWrongAnswersByAttempt(ctx, attempt.ID)
			var wrong []models.WrongAnswer
			for _, a := range wrongAnswers {
				wrong = append(wrong, models.WrongAnswer{
					ID:            a.ID,
					QuestionID:    a.QuestionID,
					QuizID:        attempt.QuizID,
					UserAnswer:    a.UserAnswer,
					CorrectAnswer: "",
					Timestamp:     attempt.StartedAt,
				})
			}
			quizErrors = append(quizErrors, types.QuizErrors{
				Quiz:         quiz,
				WrongAnswers: wrong,
			})
		}
	}

	data := types.ErrorsPageData{
		User:       user,
		QuizErrors: quizErrors,
		Stats:      stats,
	}

	pages.ErrorsPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) LeaderboardPage(w http.ResponseWriter, r *http.Request) {
	_, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ctx := r.Context()
	userID, _ := h.getUserIDFromRequest(r)
	user, _ := h.repo.GetUserByID(ctx, userID)
	entries := h.gamification.GetLeaderboard(ctx)

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

func extractQuizIDFromPath(path string) (uuid.UUID, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return uuid.Nil, fmt.Errorf("invalid path")
	}
	return uuid.Parse(parts[1])
}
