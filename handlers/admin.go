package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages/admin"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
	"github.com/google/uuid"
)

type AdminHandler struct {
	repo        *store.Repository
	authService *services.AuthService
}

func NewAdmin(r *store.Repository, a *services.AuthService) *AdminHandler {
	return &AdminHandler{repo: r, authService: a}
}

func (h *AdminHandler) getUser(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}
	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return nil, err
	}
	return h.repo.GetUserByID(claims.UserID)
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	quizzes, _ := h.repo.GetAllQuizzesWithStats()
	studentCount, _ := h.repo.GetStudentCount()
	stats, _ := h.repo.GetAdminStatistics()
	recentActivity, _ := h.repo.GetRecentAttempts(10)

	data := types.AdminDashboardData{
		User:           user,
		QuizCount:      len(quizzes),
		StudentCount:   int(studentCount),
		AttemptCount:   stats.TotalAttempts,
		AvgScore:       stats.AvgScore,
		RecentActivity: recentActivity,
	}

	admin.DashboardPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) Quizzes(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	quizzes, _ := h.repo.GetAllQuizzesWithStats()

	data := types.AdminQuizListData{
		User:    user,
		Quizzes: quizzes,
	}

	admin.QuizzesPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) QuizzesCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.createQuiz(w, r)
		return
	}
	h.Quizzes(w, r)
}

func (h *AdminHandler) QuizOp(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/question/delete") {
		h.deleteQuestion(w, r)
		return
	}

	if strings.HasSuffix(path, "/question") && r.Method == "POST" {
		h.addQuestion(w, r)
		return
	}

	if strings.HasSuffix(path, "/delete") && r.Method == "POST" {
		idStr := strings.TrimPrefix(path, "/quizzes/")
		idStr = strings.TrimSuffix(idStr, "/delete")
		quizID, err := uuid.Parse(idStr)
		if err != nil {
			http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
			return
		}
		h.deleteQuiz(w, r, quizID)
		return
	}

	idStr := strings.TrimPrefix(path, "/quizzes/")
	quizID, err := uuid.Parse(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	if r.Method == "POST" {
		h.updateQuiz(w, r, quizID)
		return
	}

	h.editQuiz(w, r, quizID)
}

func (h *AdminHandler) createQuiz(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	subject := r.FormValue("subject")
	gradeStr := r.FormValue("grade")
	timeLimitStr := r.FormValue("time_limit")

	grade, _ := strconv.Atoi(gradeStr)
	timeLimit, _ := strconv.Atoi(timeLimitStr)

	quiz := &models.Quiz{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Subject:     subject,
		Grade:       grade,
		Status:      models.QuizStatusAvailable,
		TimeLimit:   timeLimit,
		CreatedBy:   user.ID,
	}

	h.repo.CreateQuiz(quiz)

	http.Redirect(w, r, "/admin/quizzes/"+quiz.ID.String(), http.StatusFound)
}

func (h *AdminHandler) editQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) {
	user, err := h.getUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	quiz, err := h.repo.GetQuizByID(quizID)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	data := types.AdminQuizEditData{
		User:      user,
		Quiz:      quiz,
		Questions: quiz.Questions,
	}

	admin.QuizEditPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) updateQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) {
	quiz, err := h.repo.GetQuizByID(quizID)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	quiz.Title = r.FormValue("title")
	quiz.Description = r.FormValue("description")
	quiz.Subject = r.FormValue("subject")
	quiz.Grade, _ = strconv.Atoi(r.FormValue("grade"))
	quiz.TimeLimit, _ = strconv.Atoi(r.FormValue("time_limit"))
	quiz.Status = models.QuizStatus(r.FormValue("status"))

	h.repo.UpdateQuiz(quiz)

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
}

func (h *AdminHandler) deleteQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) {
	h.repo.DeleteQuiz(quizID)
	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
}

func (h *AdminHandler) addQuestion(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/quizzes/")
	idStr = strings.TrimSuffix(idStr, "/question")
	quizID, err := uuid.Parse(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	text := r.FormValue("text")
	questionType := models.QuestionType(r.FormValue("type"))
	explanation := r.FormValue("explanation")
	pointsStr := r.FormValue("points")
	orderIndexStr := r.FormValue("order_index")

	points, _ := strconv.Atoi(pointsStr)
	orderIndex, _ := strconv.Atoi(orderIndexStr)

	if points == 0 {
		points = 10
	}

	var options []string
	if questionType == models.QuestionTypeChoice {
		opts := r.FormValue("options")
		if opts != "" {
			options = []string{opts}
		}
	}

	optionsJSON, _ := json.Marshal(options)

	correctAnswer := r.FormValue("correct_answer")

	question := &models.Question{
		ID:            uuid.New(),
		QuizID:        quizID,
		Text:          text,
		Type:          questionType,
		Options:       optionsJSON,
		CorrectAnswer: correctAnswer,
		Explanation:   explanation,
		Points:        points,
		OrderIndex:    orderIndex,
	}

	h.repo.CreateQuestion(question)

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
}

func (h *AdminHandler) updateQuestion(w http.ResponseWriter, r *http.Request) {
	questionIDStr := r.FormValue("question_id")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	question, err := h.repo.GetQuestionByID(questionID)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	question.Text = r.FormValue("text")
	question.Type = models.QuestionType(r.FormValue("type"))
	question.Explanation = r.FormValue("explanation")
	question.CorrectAnswer = r.FormValue("correct_answer")
	question.Points, _ = strconv.Atoi(r.FormValue("points"))

	if question.Type == models.QuestionTypeChoice {
		opts := r.FormValue("options")
		if opts != "" {
			options := []string{opts}
			question.Options, _ = json.Marshal(options)
		}
	}

	h.repo.UpdateQuestion(question)

	http.Redirect(w, r, "/admin/quizzes/"+question.QuizID.String(), http.StatusFound)
}

func (h *AdminHandler) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionIDStr := r.FormValue("question_id")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}

	question, err := h.repo.GetQuestionByID(questionID)
	if err != nil {
		http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
		return
	}
	quizID := question.QuizID

	h.repo.DeleteQuestion(questionID)

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
}

func (h *AdminHandler) Results(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	attempts, _ := h.repo.GetAllAttempts()
	quizzes, _ := h.repo.GetNonArchivedQuizzes()

	data := types.AdminResultsData{
		User:     user,
		Attempts: attempts,
		Quizzes:  quizzes,
	}

	admin.ResultsPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	stats, _ := h.repo.GetAdminStatistics()
	stats.User = user

	admin.StatisticsPage(stats).Render(r.Context(), w)
}
