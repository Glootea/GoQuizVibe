package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages/admin"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
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
		return nil, errors.Join(errors.New("get cookie"), err)
	}
	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return nil, errors.Join(errors.New("validate token"), err)
	}
	return h.repo.GetUserByID(claims.UserID)
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
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

	return admin.DashboardPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) Quizzes(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	quizzes, _ := h.repo.GetAllQuizzesWithStats()

	data := types.AdminQuizListData{
		User:    user,
		Quizzes: quizzes,
	}

	return admin.QuizzesPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) QuizzesCreate(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		return h.createQuiz(w, r)
	}
	return h.Quizzes(w, r)
}

func (h *AdminHandler) QuizOp(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path

	if strings.HasSuffix(path, "/question/delete") && r.Method == "POST" {
		return h.deleteQuestion(w, r)
	}

	if strings.HasSuffix(path, "/question") && r.Method == "POST" {
		if r.FormValue("question_id") != "" {
			return h.updateQuestion(w, r)
		}
		return h.addQuestion(w, r)
	}

	idStr := strings.TrimPrefix(path, "/admin/quizzes/")
	quizID, err := uuid.Parse(idStr)
	if err != nil {
		if strings.HasSuffix(path, "/restore") && r.Method == "POST" {
			idStr2 := strings.TrimSuffix(strings.TrimPrefix(path, "/admin/quizzes/"), "/restore")
			quizID, err = uuid.Parse(idStr2)
			if err == nil {
				return h.RestoreQuiz(w, r)
			}
		}
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, fmt.Errorf("invalid quiz ID: %s", idStr)), http.StatusNotFound)
	}

	if strings.HasSuffix(path, "/restore") && r.Method == "POST" {
		return h.RestoreQuiz(w, r)
	}

	if strings.HasSuffix(path, "/delete") && r.Method == "POST" {
		return h.deleteQuiz(w, r, quizID)
	}

	if r.Method == "POST" {
		return h.updateQuiz(w, r, quizID)
	}

	return h.editQuiz(w, r, quizID)
}

func (h *AdminHandler) createQuiz(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
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

	err = h.repo.CreateQuiz(quiz)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}

func (h *AdminHandler) editQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	quiz, err := h.repo.GetQuizByID(quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	data := types.AdminQuizEditData{
		User:      user,
		Quiz:      quiz,
		Questions: quiz.Questions,
	}

	return admin.QuizEditPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) updateQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	quiz, err := h.repo.GetQuizByID(quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	quiz.Title = r.FormValue("title")
	quiz.Description = r.FormValue("description")
	quiz.Subject = r.FormValue("subject")
	quiz.Grade, _ = strconv.Atoi(r.FormValue("grade"))
	quiz.TimeLimit, _ = strconv.Atoi(r.FormValue("time_limit"))
	quiz.Status = models.QuizStatus(r.FormValue("status"))

	err = h.repo.UpdateQuiz(quiz)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) deleteQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	err := h.repo.DeleteQuiz(quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}

func (h *AdminHandler) RestoreQuiz(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/admin/quizzes/")
	idStr = strings.TrimSuffix(idStr, "/restore")
	quizID, err := uuid.Parse(idStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, fmt.Errorf("invalid quiz ID: %s", idStr)), http.StatusNotFound)
	}

	err = h.repo.UpdateQuizStatus(quizID, models.QuizStatusAvailable)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}

func (h *AdminHandler) addQuestion(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/quizzes/")
	idStr = strings.TrimSuffix(idStr, "/question")
	quizID, err := uuid.Parse(idStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, fmt.Errorf("invalid quiz ID in path: %s", r.URL.Path)), http.StatusBadRequest)
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
	var correctAnswer string

	if questionType == models.QuestionTypeChoice {
		r.ParseForm()
		for key, values := range r.Form {
			if strings.HasPrefix(key, "option_") {
				val := values[0]
				if val != "" {
					options = append(options, val)
				}
			}
		}
		correctAnswerRaw := r.FormValue("correct_answer")
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx < len(options) {
			correctAnswer = options[idx]
		}
	} else {
		correctAnswer = r.FormValue("correct_answer")
	}

	optionsJSON, _ := json.Marshal(options)

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

	err = h.repo.CreateQuestion(question)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) deleteQuestion(w http.ResponseWriter, r *http.Request) error {
	questionIDStr := r.FormValue("question_id")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		if r.Header.Get("hx-request") == "true" {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("invalid question ID")), http.StatusBadRequest)
		}
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, errors.New("question not found")), http.StatusNotFound)
	}

	question, err := h.repo.GetQuestionByID(questionID)
	if err != nil {
		if r.Header.Get("hx-request") == "true" {
			return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
		}
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}
	quizID := question.QuizID

	err = h.repo.DeleteQuestion(questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) updateQuestion(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	questionIDStr := r.FormValue("question_id")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("invalid question ID")), http.StatusBadRequest)
	}

	question, err := h.repo.GetQuestionByID(questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	question.Text = r.FormValue("text")
	question.Type = models.QuestionType(r.FormValue("type"))
	question.Explanation = r.FormValue("explanation")
	question.Points, _ = strconv.Atoi(r.FormValue("points"))

	if question.Type == models.QuestionTypeChoice {
		var options []string
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("option_%d", i)
			if val, ok := r.Form[key]; ok && val[0] != "" {
				options = append(options, val[0])
			}
		}
		question.Options, _ = json.Marshal(options)
		correctAnswerRaw := r.FormValue("correct_answer")
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx < len(options) {
			question.CorrectAnswer = options[idx]
		}
	} else {
		question.CorrectAnswer = r.FormValue("correct_answer")
	}

	err = h.repo.UpdateQuestion(question)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+question.QuizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) Results(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	attempts, _ := h.repo.GetAllAttempts()
	quizzes, _ := h.repo.GetNonArchivedQuizzes()

	data := types.AdminResultsData{
		User:     user,
		Attempts: attempts,
		Quizzes:  quizzes,
	}

	return admin.ResultsPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) Statistics(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	stats, _ := h.repo.GetAdminStatistics()
	stats.User = user

	return admin.StatisticsPage(stats).Render(r.Context(), w)
}
