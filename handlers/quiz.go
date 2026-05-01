package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
)

type QuizHandler struct {
	repo        *store.Repository
	quizService  *services.QuizService
	gamification *services.GamificationService
	authService  *services.AuthService
}

func NewQuiz(r *store.Repository, qs *services.QuizService, gs *services.GamificationService, as *services.AuthService) *QuizHandler {
	return &QuizHandler{
		repo:        r,
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

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	stats, _ := h.gamification.GetUserStats(userID)

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		session := h.createSession(userID, quizID)
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

func (h *QuizHandler) createSession(userID, quizID uuid.UUID) *models.QuizSession {
	attempt := &models.QuizAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: time.Now(),
	}
	h.repo.SaveAttempt(attempt)

	session := &models.QuizSession{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		AttemptID: attempt.ID,
	}
	h.repo.CreateSession(session)
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

	session, err := h.repo.GetSession(sessionID)
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
	h.repo.UpdateSession(session)

	userAnswer := &models.UserAnswer{
		ID:         uuid.New(),
		AttemptID:  session.AttemptID,
		QuestionID: question.ID,
		UserAnswer: answer,
		IsCorrect:  isCorrect,
	}
	h.repo.SaveUserAnswer(userAnswer)

	if isCorrect {
		h.gamification.AwardXP(userID, 10)
	}

	feedback := &pages.QuestionFeedbackData{
		IsCorrect:     isCorrect,
		CorrectAnswer: question.CorrectAnswer,
		Explanation:   question.Explanation,
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

	quiz, err := h.quizService.GetQuizByID(quizID)
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

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	stats, _ := h.gamification.GetUserStats(userID)

	attempt, err := h.completeSession(sessionID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	var correctCount, wrongCount int
	var answers []types.AnswerDetail

	answersMap := make(map[uuid.UUID]string)
	for _, a := range attempt.Answers {
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
		isCorrect := answersMap[q.ID] != "" && NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

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

func (h *QuizHandler) completeSession(sessionID uuid.UUID) (*models.QuizAttempt, error) {
	session, err := h.repo.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	attempt, err := h.repo.GetAttemptByID(session.AttemptID)
	if err != nil {
		return nil, err
	}

	quiz, err := h.quizService.GetQuizByID(session.QuizID)
	if err != nil {
		return nil, err
	}

	answers, _ := h.repo.GetAnswersByAttempt(attempt.ID)

	var score, maxScore int
	for _, q := range quiz.Questions {
		maxScore += q.Points
		for _, a := range answers {
			if a.QuestionID == q.ID && a.IsCorrect {
				score += q.Points
				break
			}
		}
	}

	now := time.Now()
	attempt.Score = score
	attempt.MaxScore = maxScore
	attempt.CompletedAt = &now
	h.repo.UpdateAttempt(attempt)

	h.repo.DeleteSession(sessionID)

	return attempt, nil
}

func (h *QuizHandler) ErrorsPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	stats, _ := h.gamification.GetUserStats(userID)

	attempts, _ := h.repo.GetQuizErrors(userID)

	var quizErrors []types.QuizErrors
	for _, attempt := range attempts {
		quiz, err := h.quizService.GetQuizByID(attempt.QuizID)
		if err == nil {
			var wrongAnswers []models.WrongAnswer
			for _, a := range attempt.Answers {
				wrongAnswers = append(wrongAnswers, models.WrongAnswer{
					ID:            a.ID,
					QuestionID:    a.QuestionID,
					QuizID:        attempt.QuizID,
					UserAnswer:    a.UserAnswer,
					CorrectAnswer: "",
					Timestamp:     time.Time{},
				})
			}
			quizErrors = append(quizErrors, types.QuizErrors{
				Quiz:        quiz,
				WrongAnswers: wrongAnswers,
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

	userID, _ := h.getUserIDFromRequest(r)
	user, _ := h.repo.GetUserByID(userID)
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