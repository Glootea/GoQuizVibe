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
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages/admin"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"

	ce "github.com/goquizvibe/custom_errors"
)

type AdminHandler struct {
	pool        *db.Queries
	authService *services.AuthService
}

func NewAdmin(pool *db.Queries, a *services.AuthService) *AdminHandler {
	return &AdminHandler{pool: pool, authService: a}
}

func (h *AdminHandler) getUser(r *http.Request) (*db.User, error) {
	ctx := context.Background()
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, errors.Join(errors.New("get cookie"), err)
	}
	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return nil, errors.Join(errors.New("validate token"), err)
	}
	user, err := h.pool.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.Join(errors.New("can not get user from db"), err)
	}

	return &user, nil
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	ctx := context.Background()
	quizzes, _ := h.pool.GetNonArchivedQuizzes(ctx)
	studentCount, _ := h.pool.GetStudentCount(ctx)
	stats, _ := h.pool.GetAdminStatsData(ctx)
	recentActivity, _ := h.pool.GetRecentAttempts(ctx, 10)

	recentAttempts := make([]*types.RecentAttempt, 0, len(recentActivity))
	for _, a := range recentActivity {
		recentAttempts = append(recentAttempts, &types.RecentAttempt{
			AttemptID:   a.ID.String(),
			UserName:    a.UserName,
			QuizTitle:   a.QuizTitle,
			Score:       int(a.Score),
			MaxScore:    int(a.MaxScore),
			CompletedAt: a.CompletedAt.Format("2006-01-02 15:04"),
		})
	}

	var avgScore float64
	if stats.AvgScore != nil {
		if f, ok := stats.AvgScore.(float64); ok {
			avgScore = f
		}
	}

	data := types.AdminDashboardData{
		User:           user,
		QuizCount:      len(quizzes),
		StudentCount:   int(studentCount),
		AttemptCount:   int(stats.TotalAttempts),
		AvgScore:       avgScore,
		RecentActivity: recentAttempts,
	}

	return admin.DashboardPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) Quizzes(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	ctx := context.Background()
	quizzes, _ := h.pool.GetNonArchivedQuizzes(ctx)

	data := types.AdminQuizListData{
		User:    user,
		Quizzes: nil,
	}

	quizWithStats := make([]*types.QuizWithStats, len(quizzes))
	for i, q := range quizzes {
		quizWithStats[i] = &types.QuizWithStats{
			Quiz: &models.Quiz{Quiz: q, Questions: nil},
		}
	}
	data.Quizzes = quizWithStats

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
	grade, _ := strconv.Atoi(r.FormValue("grade"))
	timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))

	_, err = h.pool.CreateQuiz(context.Background(), db.CreateQuizParams{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Subject:     subject,
		Grade:       grade,
		Status:      db.QuizStatusAvailable,
		TimeLimit:   timeLimit,
		CreatedBy:   user.ID,
		CreatedAt:   time.Now(),
	})
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

	quiz, err := h.pool.GetQuizByID(context.Background(), quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	questions, err := h.pool.GetQuestionsByQuizID(context.Background(), quizID)
	if err != nil {
		questions = nil
	}

	data := types.AdminQuizEditData{
		User:      user,
		Quiz:      &models.Quiz{Quiz: quiz, Questions: questions},
		Questions: questions,
	}

	return admin.QuizEditPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) updateQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	_, err := h.pool.GetQuizByID(context.Background(), quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	grade, _ := strconv.Atoi(r.FormValue("grade"))
	timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))

	_, err = h.pool.UpdateQuiz(context.Background(), db.UpdateQuizParams{
		ID:          quizID,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Subject:     r.FormValue("subject"),
		Grade:       grade,
		Status:      db.QuizStatus(r.FormValue("status")),
		TimeLimit:   timeLimit,
	})
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) deleteQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	err := h.pool.DeleteQuiz(context.Background(), quizID)
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

	err = h.pool.UpdateQuizStatus(context.Background(), db.UpdateQuizStatusParams{
		ID:     quizID,
		Status: db.QuizStatusAvailable,
	})
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
	questionType := db.QuestionType(r.FormValue("type"))
	explanation := r.FormValue("explanation")
	points, _ := strconv.Atoi(r.FormValue("points"))
	orderIndex, _ := strconv.Atoi(r.FormValue("order_index"))

	if points == 0 {
		points = 10
	}

	var options []string
	var correctAnswer string

	if questionType == db.QuestionTypeChoice {
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

	_, err = h.pool.CreateQuestion(context.Background(), db.CreateQuestionParams{
		ID:            uuid.New(),
		QuizID:        quizID,
		Text:          text,
		Type:          questionType,
		Options:       optionsJSON,
		CorrectAnswer: correctAnswer,
		Explanation:   explanation,
		Points:        points,
		OrderIndex:    orderIndex,
	})
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

	question, err := h.pool.GetQuestionByID(context.Background(), questionID)
	if err != nil {
		if r.Header.Get("hx-request") == "true" {
			return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
		}
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}
	quizID := question.QuizID

	err = h.pool.DeleteQuestion(context.Background(), questionID)
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

	question, err := h.pool.GetQuestionByID(context.Background(), questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	points, _ := strconv.Atoi(r.FormValue("points"))
	orderIndex, _ := strconv.Atoi(r.FormValue("order_index"))

	var options []byte
	var correctAnswer string
	questionType := db.QuestionType(r.FormValue("type"))

	if questionType == db.QuestionTypeChoice {
		var opts []string
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("option_%d", i)
			if val, ok := r.Form[key]; ok && val[0] != "" {
				opts = append(opts, val[0])
			}
		}
		options, _ = json.Marshal(opts)
		correctAnswerRaw := r.FormValue("correct_answer")
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx < len(opts) {
			correctAnswer = opts[idx]
		}
	} else {
		correctAnswer = r.FormValue("correct_answer")
	}

	_, err = h.pool.UpdateQuestion(context.Background(), db.UpdateQuestionParams{
		ID:            questionID,
		Text:          r.FormValue("text"),
		Type:          questionType,
		Options:       options,
		CorrectAnswer: correctAnswer,
		Explanation:   r.FormValue("explanation"),
		Points:        points,
		OrderIndex:    orderIndex,
	})
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

	attempts, _ := h.pool.GetAllAttempts(context.Background())
	quizzes, _ := h.pool.GetNonArchivedQuizzes(context.Background())

	attemptWithUser := make([]*types.AttemptWithUser, 0, len(attempts))
	for _, a := range attempts {
		attemptWithUser = append(attemptWithUser, &types.AttemptWithUser{
			QuizAttempt: &db.QuizAttempt{
				ID:          a.ID,
				UserID:      a.UserID,
				QuizID:      a.QuizID,
				Score:       a.Score,
				MaxScore:    a.MaxScore,
				StartedAt:   a.StartedAt,
				CompletedAt: a.CompletedAt,
			},
			UserName:  a.UserName,
			QuizTitle: a.QuizTitle,
		})
	}

	data := types.AdminResultsData{
		User:     user,
		Attempts: attemptWithUser,
		Quizzes:  nil,
	}

	data.Quizzes = make([]*models.Quiz, len(quizzes))
	for i, q := range quizzes {
		data.Quizzes[i] = &models.Quiz{Quiz: q}
	}

	return admin.ResultsPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) Statistics(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	stats, _ := h.pool.GetAdminStatsData(context.Background())

	var avgScore float64
	if stats.AvgScore != nil {
		if f, ok := stats.AvgScore.(float64); ok {
			avgScore = f
		}
	}

	data := types.AdminStatisticsData{
		User:                user,
		TotalQuizzes:        int(stats.TotalQuizzes),
		TotalStudents:       int(stats.TotalStudents),
		TotalAttempts:       int(stats.TotalAttempts),
		AvgScore:            avgScore,
		QuizStats:           nil,
		GradeDistribution:   nil,
		SubjectDistribution: nil,
	}

	return admin.StatisticsPage(&data).Render(r.Context(), w)
}

func (h *AdminHandler) QuizStatsData(w http.ResponseWriter, r *http.Request) error {
	quizStats, err := h.pool.GetQuizStats(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	stats := make([]types.QuizStatsResponse, 0, len(quizStats))
	for _, qs := range quizStats {
		var avgScore float64
		if qs.AvgScore != nil {
			if f, ok := qs.AvgScore.(float64); ok {
				avgScore = f
			}
		}
		stats = append(stats, types.QuizStatsResponse{
			QuizID:       qs.QuizID,
			Title:        qs.Title,
			Subject:      qs.Subject,
			AttemptCount: int(qs.AttemptCount),
			AvgScore:     avgScore,
			PassRate:     float64(qs.PassRate),
		})
	}

	return admin.QuizStatsPartial(stats).Render(r.Context(), w)
}

func (h *AdminHandler) GradeDistData(w http.ResponseWriter, r *http.Request) error {
	data, err := h.pool.GetGradeDistribution(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	var dist map[string]int
	if data != nil {
		json.Unmarshal(data, &dist)
	}
	if dist == nil {
		dist = make(map[string]int)
	}

	return admin.GradeDistPartial(dist).Render(r.Context(), w)
}

func (h *AdminHandler) SubjectDistData(w http.ResponseWriter, r *http.Request) error {
	data, err := h.pool.GetSubjectDistribution(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	var dist map[string]int
	if data != nil {
		json.Unmarshal(data, &dist)
	}
	if dist == nil {
		dist = make(map[string]int)
	}

	return admin.SubjectDistPartial(dist).Render(r.Context(), w)
}