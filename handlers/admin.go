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
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/locales"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages/admin"
	"github.com/goquizvibe/services"
)

type AdminHandler struct {
	adminService *services.AdminService
	authService  *services.AuthService
	localeSvc    *locales.Service
}

func NewAdmin(adminSvc *services.AdminService, auth *services.AuthService, svc *locales.Service) *AdminHandler {
	return &AdminHandler{
		adminService: adminSvc,
		authService:  auth,
		localeSvc:    svc,
	}
}

func (h *AdminHandler) getUser(r *http.Request) (*db.User, error) {
	return h.adminService.GetUserFromRequest(r)
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.adminService.GetDashboardData(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.DashboardPage(*data, t).Render(r.Context(), w)
}

func (h *AdminHandler) Quizzes(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.adminService.GetQuizzesListData(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.QuizzesPage(*data, t).Render(r.Context(), w)
}

func (h *AdminHandler) QuizzesNew(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	if r.Method == "POST" {
		title := r.FormValue("title")
		description := r.FormValue("description")
		subject := r.FormValue("subject")
		grade, _ := strconv.Atoi(r.FormValue("grade"))
		timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))

		newQuizID, err := h.adminService.CreateQuiz(r.Context(), user.ID, title, description, subject, grade, timeLimit)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}

		if r.Header.Get("hx-request") == "true" {
			w.Header().Set("HX-Redirect", "/admin/quizzes/"+newQuizID.String())
			w.WriteHeader(http.StatusOK)
			return nil
		}

		http.Redirect(w, r, "/admin/quizzes/"+newQuizID.String(), http.StatusFound)
		return nil
	}

	return h.Quizzes(w, r)
}

func (h *AdminHandler) QuizView(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	data, err := h.adminService.GetQuizEditData(r.Context(), user.ID, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.QuizEditPage(*data, t).Render(r.Context(), w)
}

func (h *AdminHandler) QuizUpdate(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	grade, _ := strconv.Atoi(r.FormValue("grade"))
	timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))
	status := db.QuizStatus(r.FormValue("status"))

	err = h.adminService.UpdateQuiz(r.Context(), quizID, r.FormValue("title"), r.FormValue("description"), r.FormValue("subject"), grade, timeLimit, status)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) QuizDelete(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	err = h.adminService.DeleteQuiz(r.Context(), quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}

func (h *AdminHandler) AddQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	text, questionTypeStr, explanation, correctAnswer, points, orderIndex, options, files, err := services.ParseQuestionForm(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	questionType := db.QuestionType(questionTypeStr)
	if questionTypeStr == "" {
		questionType = db.QuestionTypeChoice
	}

	_, err = h.adminService.AddQuestion(r.Context(), quizID, text, questionType, options, correctAnswer, explanation, points, orderIndex, files)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if r.Header.Get("hx-request") == "true" {
		questions, err := h.adminService.GetQuestionsByQuizID(r.Context(), quizID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}

		quiz, err := h.adminService.GetQuizByID(r.Context(), quizID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}

		quizWithQuestions := models.QuizWithQuestionsAndImages{
			Quiz:      quiz.Quiz,
			Questions: questions,
		}

		t := middleware.GetTranslator(r.Context())
		return admin.QuestionsSection(&quizWithQuestions, questions, t).Render(r.Context(), w)
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	questionID, err := uuid.Parse(r.PathValue("qid"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
		}
	}

	points, _ := strconv.Atoi(r.FormValue("points"))
	orderIndex, _ := strconv.Atoi(r.FormValue("order_index"))
	correctAnswer := r.FormValue("correct_answer")
	explanation := r.FormValue("explanation")
	text := r.FormValue("text")

	var options []byte
	questionTypeStr := r.FormValue("type")
	if questionTypeStr == "" {
		questionTypeStr = string(db.QuestionTypeChoice)
	}
	questionType := db.QuestionType(questionTypeStr)

	if questionType == db.QuestionTypeChoice {
		var opts []string
		for i := range 20 {
			key := fmt.Sprintf("option_%d", i)
			if val, ok := r.Form[key]; ok && val[0] != "" {
				opts = append(opts, val[0])
			}
		}
		options, _ = json.Marshal(opts)
		correctAnswerRaw := r.FormValue("correct_answer")
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx >= 0 && idx < len(opts) {
			correctAnswer = opts[idx]
		}
	}

	err = h.adminService.UpdateQuestion(r.Context(), questionID, quizID, text, questionType, options, correctAnswer, explanation, points, orderIndex)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	questionID, err := uuid.Parse(r.PathValue("qid"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	err = h.adminService.DeleteQuestion(r.Context(), questionID, quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) UploadQuestionImage(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	questionID, err := uuid.Parse(r.PathValue("qid"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	if err := r.ParseMultipartForm(services.MaxImageSize); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}
	defer file.Close()

	err = h.adminService.UploadQuestionImage(r.Context(), quizID, questionID, header)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) DeleteQuestionImage(w http.ResponseWriter, r *http.Request) error {
	questionID, err := uuid.Parse(r.PathValue("qid"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	imageID, err := uuid.Parse(r.PathValue("imgid"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	err = h.adminService.DeleteQuestionImage(r.Context(), imageID, questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	return nil
}

func (h *AdminHandler) Results(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.adminService.GetResultsData(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.ResultsPage(*data, t).Render(r.Context(), w)
}

func (h *AdminHandler) Statistics(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.adminService.GetStatisticsData(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.StatisticsPage(data, t).Render(r.Context(), w)
}

func (h *AdminHandler) QuizStatsData(w http.ResponseWriter, r *http.Request) error {
	stats, err := h.adminService.GetQuizStatsData(r.Context())
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.QuizStatsPartial(stats, t).Render(r.Context(), w)
}

func (h *AdminHandler) GradeDistData(w http.ResponseWriter, r *http.Request) error {
	dist, err := h.adminService.GetGradeDistributionData(r.Context())
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.GradeDistPartial(dist, t).Render(r.Context(), w)
}

func (h *AdminHandler) SubjectDistData(w http.ResponseWriter, r *http.Request) error {
	dist, err := h.adminService.GetSubjectDistributionData(r.Context())
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.SubjectDistPartial(dist, t).Render(r.Context(), w)
}

func (h *AdminHandler) RestoreQuiz(w http.ResponseWriter, r *http.Request) error {
	quizID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	err = h.adminService.RestoreQuiz(r.Context(), quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}
