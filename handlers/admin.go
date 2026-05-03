package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
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

const (
	MaxImagesPerQuestion = 3
	MaxImageSize         = 5 << 20
	AllowedImageTypes    = "image/jpeg,image/png,image/webp"
)

type AdminHandler struct {
	pool           *db.Queries
	authService    *services.AuthService
	storageService *services.StorageService
}

func NewAdmin(pool *db.Queries, a *services.AuthService, storage *services.StorageService) *AdminHandler {
	return &AdminHandler{pool: pool, authService: a, storageService: storage}
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

func (h *AdminHandler) actualMethod(r *http.Request) string {
	if m := r.Header.Get("Hx-Http-Method"); m != "" {
		return m
	}
	return r.Method
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

	avgScore := 0.0
	if stats.AvgScore != nil {
		switch v := stats.AvgScore.(type) {
		case float64:
			avgScore = v
		case int64:
			avgScore = float64(v)
		default:
			if num, ok := v.(interface{ Float64() (float64, error) }); ok {
				f, _ := num.Float64()
				avgScore = f
			}
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

	quizStats, _ := h.pool.GetQuizStats(ctx)
	statsMap := make(map[string]types.QuizStatsResponse)
	for _, qs := range quizStats {
		avgScore := 0.0
		if qs.AvgScore != nil {
			switch v := qs.AvgScore.(type) {
			case float64:
				avgScore = v
			case int64:
				avgScore = float64(v)
			default:
				if num, ok := v.(interface{ Float64() (float64, error) }); ok {
					f, _ := num.Float64()
					avgScore = f
				}
			}
		}
		statsMap[qs.QuizID.String()] = types.QuizStatsResponse{
			QuizID:       qs.QuizID,
			Title:        qs.Title,
			Subject:      qs.Subject,
			AttemptCount: int(qs.AttemptCount),
			AvgScore:     avgScore,
			PassRate:     float64(qs.PassRate),
		}
	}

	quizWithStats := make([]*types.QuizWithStats, len(quizzes))
	for i, q := range quizzes {
		stats := statsMap[q.ID.String()]
		quizWithStats[i] = &types.QuizWithStats{
			Quiz:         &models.Quiz{Quiz: q, Questions: nil},
			AttemptCount: stats.AttemptCount,
			AvgScore:     stats.AvgScore,
		}
	}
	data.Quizzes = quizWithStats

	return admin.QuizzesPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) QuizzesNew(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	if h.actualMethod(r) == "POST" {
		title := r.FormValue("title")
		description := r.FormValue("description")
		subject := r.FormValue("subject")
		grade, _ := strconv.Atoi(r.FormValue("grade"))
		timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))

		newQuizID := uuid.New()
		_, err = h.pool.CreateQuiz(context.Background(), db.CreateQuizParams{
			ID:          newQuizID,
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
	quizID, _ := uuid.Parse(r.PathValue("id"))
	return h.editQuiz(w, r, quizID)
}

func (h *AdminHandler) QuizUpdate(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	return h.updateQuiz(w, r, quizID)
}

func (h *AdminHandler) QuizDelete(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	return h.deleteQuiz(w, r, quizID)
}

func (h *AdminHandler) AddQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	return h.addQuestion(w, r, quizID)
}

func (h *AdminHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	questionID, _ := uuid.Parse(r.PathValue("qid"))
	return h.updateQuestion(w, r, quizID, questionID)
}

func (h *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	questionID, _ := uuid.Parse(r.PathValue("qid"))
	return h.deleteQuestion(w, r, quizID, questionID)
}

func (h *AdminHandler) UploadQuestionImage(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	questionID, _ := uuid.Parse(r.PathValue("qid"))
	return h.uploadQuestionImage(w, r, quizID, questionID)
}

func (h *AdminHandler) DeleteQuestionImage(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))
	questionID, _ := uuid.Parse(r.PathValue("qid"))
	imageID, _ := uuid.Parse(r.PathValue("imgid"))
	return h.deleteQuestionImage(w, r, quizID, questionID, imageID)
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

	questionsWithImages := h.attachImagesToQuestions(context.Background(), questions)

	quizWithQuestions := models.QuizWithQuestionsAndImages{
		Quiz:      quiz,
		Questions: questionsWithImages,
	}

	data := types.AdminQuizEditData{
		User:      user,
		Quiz:      &quizWithQuestions,
		Questions: questionsWithImages,
	}

	return admin.QuizEditPage(data).Render(r.Context(), w)
}

func (h *AdminHandler) attachImagesToQuestions(ctx context.Context, questions []db.Question) []models.Question {
	result := make([]models.Question, len(questions))
	for i, q := range questions {
		images, _ := h.pool.GetImagesByQuestionID(ctx, q.ID)
		result[i] = models.Question{
			Question: q,
			Images:   images,
		}
	}
	return result
}

func (h *AdminHandler) updateQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	_, err := h.pool.GetQuizByID(context.Background(), quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	grade, _ := strconv.Atoi(r.FormValue("grade"))
	timeLimit, _ := strconv.Atoi(r.FormValue("time_limit"))

	inserted, err := h.pool.UpdateQuiz(context.Background(), db.UpdateQuizParams{
		ID:          quizID,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Subject:     r.FormValue("subject"),
		Grade:       grade,
		Status:      db.QuizStatus(r.FormValue("status")),
		TimeLimit:   timeLimit,
	})
	insertedJson, _ := json.Marshal(inserted)
	fmt.Printf("Inserted: %s", string(insertedJson))
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

func (h *AdminHandler) deleteQuiz(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	err := h.pool.DeleteQuiz(context.Background(), quizID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}

func (h *AdminHandler) addQuestion(w http.ResponseWriter, r *http.Request, quizID uuid.UUID) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
		}
	}

	text := r.FormValue("text")
	questionTypeStr := r.FormValue("type")
	if questionTypeStr == "" {
		questionTypeStr = "choice"
	}
	questionType := db.QuestionType(questionTypeStr)
	explanation := r.FormValue("explanation")
	correctAnswer := r.FormValue("correct_answer")
	points, _ := strconv.Atoi(r.FormValue("points"))
	orderIndex, _ := strconv.Atoi(r.FormValue("order_index"))

	if points == 0 {
		points = 10
	}

	var options []string

	if questionType == db.QuestionTypeChoice {
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("option_%d", i)
			if val, ok := r.Form[key]; ok && val[0] != "" {
				options = append(options, val[0])
			}
		}
		correctAnswerRaw := r.FormValue("correct_answer")
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx >= 0 && idx < len(options) {
			correctAnswer = options[idx]
		}
	}

	optionsJSON, _ := json.Marshal(options)

	newQuestionID := uuid.New()
	_, err := h.pool.CreateQuestion(context.Background(), db.CreateQuestionParams{
		ID:            newQuestionID,
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

	if mpForm := r.MultipartForm; mpForm != nil {
		files := mpForm.File["images"]
		for i, fileHeader := range files {
			if i >= MaxImagesPerQuestion {
				break
			}
			count, _ := h.pool.GetImageCountByQuestionID(context.Background(), newQuestionID)
			if count >= MaxImagesPerQuestion {
				continue
			}

			url, err := h.storageService.UploadImage(context.Background(), fileHeader)
			if err != nil {
				continue
			}

			h.pool.CreateQuestionImage(context.Background(), db.CreateQuestionImageParams{
				ID:         uuid.New(),
				QuestionID: newQuestionID,
				URL:        url,
				OrderIndex: int(count),
				CreatedAt:  time.Now(),
			})
		}
	}

	if r.Header.Get("hx-request") == "true" {
		questions, _ := h.pool.GetQuestionsByQuizID(context.Background(), quizID)
		questionsWithImages := h.attachImagesToQuestions(context.Background(), questions)

		quiz, _ := h.pool.GetQuizByID(context.Background(), quizID)

		quizWithQuestions := models.QuizWithQuestionsAndImages{
			Quiz:      quiz,
			Questions: questionsWithImages,
		}

		return admin.QuestionsSection(&quizWithQuestions, questionsWithImages).Render(r.Context(), w)
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) updateQuestion(w http.ResponseWriter, r *http.Request, quizID, questionID uuid.UUID) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
		}
	}

	question, err := h.pool.GetQuestionByID(context.Background(), questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if question.QuizID != quizID {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, errors.New("question does not belong to this quiz")), http.StatusNotFound)
	}

	points, _ := strconv.Atoi(r.FormValue("points"))
	orderIndex, _ := strconv.Atoi(r.FormValue("order_index"))
	correctAnswer := r.FormValue("correct_answer")
	explanation := r.FormValue("explanation")
	text := r.FormValue("text")

	var options []byte
	questionTypeStr := r.FormValue("type")
	if questionTypeStr == "" {
		questionTypeStr = string(question.Type)
	}
	questionType := db.QuestionType(questionTypeStr)

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
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx >= 0 && idx < len(opts) {
			correctAnswer = opts[idx]
		}
	}

	_, err = h.pool.UpdateQuestion(context.Background(), db.UpdateQuestionParams{
		ID:            questionID,
		Text:          text,
		Type:          questionType,
		Options:       options,
		CorrectAnswer: correctAnswer,
		Explanation:   explanation,
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

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) deleteQuestion(w http.ResponseWriter, r *http.Request, quizID, questionID uuid.UUID) error {
	question, err := h.pool.GetQuestionByID(context.Background(), questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if question.QuizID != quizID {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, errors.New("question does not belong to this quiz")), http.StatusNotFound)
	}

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

func (h *AdminHandler) uploadQuestionImage(w http.ResponseWriter, r *http.Request, quizID, questionID uuid.UUID) error {
	if err := r.ParseMultipartForm(MaxImageSize); err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	question, err := h.pool.GetQuestionByID(context.Background(), questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if question.QuizID != quizID {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, errors.New("question does not belong to this quiz")), http.StatusNotFound)
	}

	count, err := h.pool.GetImageCountByQuestionID(context.Background(), questionID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}
	if count >= MaxImagesPerQuestion {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("maximum images reached")), http.StatusBadRequest)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.Contains(AllowedImageTypes, contentType) {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("invalid image type")), http.StatusBadRequest)
	}

	if header.Size > MaxImageSize {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("image too large")), http.StatusBadRequest)
	}

	ext := filepath.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, errors.New("invalid file extension")), http.StatusBadRequest)
	}

	url, err := h.storageService.UploadImage(context.Background(), header)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	_, err = h.pool.CreateQuestionImage(context.Background(), db.CreateQuestionImageParams{
		ID:         uuid.New(),
		QuestionID: questionID,
		URL:        url,
		OrderIndex: int(count),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		_ = h.storageService.DeleteImage(context.Background(), url)
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if r.Header.Get("hx-request") == "true" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/quizzes/"+quizID.String(), http.StatusFound)
	return nil
}

func (h *AdminHandler) deleteQuestionImage(w http.ResponseWriter, r *http.Request, quizID, questionID, imageID uuid.UUID) error {
	ctx := context.Background()

	image, err := h.pool.GetQuestionImageByID(ctx, imageID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	if image.QuestionID != questionID {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, errors.New("image does not belong to this question")), http.StatusNotFound)
	}

	err = h.pool.DeleteQuestionImage(ctx, imageID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	objectName := filepath.Base(image.URL)
	_ = h.storageService.DeleteImage(ctx, objectName)

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

	avgScore := 0.0
	if stats.AvgScore != nil {
		switch v := stats.AvgScore.(type) {
		case float64:
			avgScore = v
		case int64:
			avgScore = float64(v)
		default:
			if num, ok := v.(interface{ Float64() (float64, error) }); ok {
				f, _ := num.Float64()
				avgScore = f
			}
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
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	stats := make([]types.QuizStatsResponse, 0, len(quizStats))
	for _, qs := range quizStats {
		avgScore := 0.0
		if qs.AvgScore != nil {
			switch v := qs.AvgScore.(type) {
			case float64:
				avgScore = v
			case int64:
				avgScore = float64(v)
			default:
				if num, ok := v.(interface{ Float64() (float64, error) }); ok {
					f, _ := num.Float64()
					avgScore = f
				}
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
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
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
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
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

func (h *AdminHandler) RestoreQuiz(w http.ResponseWriter, r *http.Request) error {
	quizID, _ := uuid.Parse(r.PathValue("id"))

	err := h.pool.UpdateQuizStatus(context.Background(), db.UpdateQuizStatusParams{
		ID:     quizID,
		Status: db.QuizStatusAvailable,
	})
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/admin/quizzes", http.StatusFound)
	return nil
}
