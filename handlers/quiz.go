package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	user, err := h.store.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	stats, _ := h.gamification.GetUserStats(userID)

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		session := h.store.CreateSession("", userID, quizID)
		sessionID = session.AttemptID.String()
	}

	data := types.QuizPageData{
		User:      user,
		Quiz:      quiz,
		Stats:     stats,
		SessionID: sessionID,
	}

	pages.QuizPage(data).Render(r.Context(), w)
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

	sessionID := r.URL.Query().Get("session")
	answer := r.FormValue("answer")
	questionIndexStr := r.FormValue("question_index")
	questionIndex, _ := strconv.Atoi(questionIndexStr)

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	quiz, err := h.quizService.GetQuizByID(quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	if questionIndex >= len(quiz.Questions) {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	if sessionID == "" {
		session := h.store.CreateSession(sessionID, userID, quizID)
		sessionID = session.AttemptID.String()
	}

	h.store.UpdateSessionAnswer(sessionID, questionIndex, answer)

	question := quiz.Questions[questionIndex]
	isCorrect := NormalizeAnswer(answer) == NormalizeAnswer(question.CorrectAnswer)
	isLast := questionIndex >= len(quiz.Questions)-1

	if isCorrect {
		h.gamification.AwardXP(userID, 10)
	}

	feedback := &pages.QuestionFeedbackData{
		IsCorrect:     isCorrect,
		CorrectAnswer: question.CorrectAnswer,
		Explanation:   question.Explanation,
		IsLast:        isLast,
	}

	if isCorrect {
		pages.QuestionWithFeedback(quiz, questionIndex, sessionID, feedback).Render(r.Context(), w)
	} else {
		pages.QuestionWithWrongFeedback(quiz, questionIndex, sessionID, feedback).Render(r.Context(), w)
	}
}

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

func (h *QuizHandler) QuizNextHTMX(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	indexStr := r.URL.Query().Get("index")
	index, _ := strconv.Atoi(indexStr)

	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}
	_ = userID

	quiz, err := h.quizService.GetQuizByID(quizID)
	if err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	if index >= len(quiz.Questions) {
		http.Redirect(w, r, "/quiz/"+quizID.String()+"/result?session="+sessionID, http.StatusFound)
		return
	}

	pages.QuestionCard(quiz, index, sessionID).Render(r.Context(), w)
}

func (h *QuizHandler) QuizResult(w http.ResponseWriter, r *http.Request) {
	quizID, err := extractQuizIDFromPath(r.URL.Path)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
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

	user, err := h.store.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	stats, _ := h.gamification.GetUserStats(userID)

	attempt, err := h.store.CompleteSession(sessionID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	var correctCount, wrongCount int
	var answers []types.AnswerDetail
	for i, q := range quiz.Questions {
		var userAnswer string
		var isCorrect bool
		if i < len(attempt.Answers) {
			ua := attempt.Answers[i]
			userAnswer = ua.UserAnswer
			isCorrect = ua.IsCorrect
		} else {
			userAnswer = "Нет ответа"
			isCorrect = false
		}

		if isCorrect {
			correctCount++
		} else {
			wrongCount++
		}

		answers = append(answers, types.AnswerDetail{
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
		Score:        attempt.Score,
		MaxScore:     attempt.MaxScore,
		CorrectCount: correctCount,
		WrongCount:   wrongCount,
		Answers:      answers,
	}

	pages.QuizResultPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) ErrorsPage(w http.ResponseWriter, r *http.Request) {
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
	progress, _ := h.gamification.GetUserStats(userID)

	data := types.ErrorsPageData{
		User:         user,
		WrongAnswers: progress.WrongAnswers,
	}

	pages.ErrorsPage(data).Render(r.Context(), w)
}

func (h *QuizHandler) LeaderboardPage(w http.ResponseWriter, r *http.Request) {
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
	entries := h.gamification.GetLeaderboard()

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
