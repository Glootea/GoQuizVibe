package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
	"github.com/goquizvibe/types"
)

type AdminService struct {
	users          r.UserRepository
	quizzes        r.QuizRepository
	questions      r.QuestionRepository
	images         r.ImageRepository
	attempts       r.AttemptRepository
	stats          r.StatsRepository
	authService    *AuthService
	storageService *StorageService
	cache          *CacheService
}

func NewAdminService(
	users r.UserRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	attempts r.AttemptRepository,
	stats r.StatsRepository,
	auth *AuthService,
	storage *StorageService,
	cache *CacheService,
) *AdminService {
	return &AdminService{
		users:          users,
		quizzes:        quizzes,
		questions:      questions,
		images:         images,
		attempts:       attempts,
		stats:          stats,
		authService:    auth,
		storageService: storage,
		cache:          cache,
	}
}

func (s *AdminService) GetUserFromRequest(r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, errors.Join(errors.New("get cookie"), err)
	}
	claims, err := s.authService.ValidateToken(cookie.Value)
	if err != nil {
		return nil, errors.Join(errors.New("validate token"), err)
	}
	cacheKey := "user:" + claims.UserID.String()
	user, err := GetOrFetch(r.Context(), s.cache, cacheKey, func() (db.User, error) {
		return s.users.GetUserByID(context.Background(), claims.UserID)
	})
	if err != nil {
		return nil, errors.Join(errors.New("can not get user from db"), err)
	}
	return &user, nil
}

func (s *AdminService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*types.AdminDashboardData, error) {
	userCacheKey := "user:" + userID.String()
	user, err := GetOrFetch(ctx, s.cache, userCacheKey, func() (db.User, error) {
		return s.users.GetUserByID(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quizzes, err := s.quizzes.GetNonArchivedQuizzes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get quizzes: %w", err)
	}

	studentCount, err := s.users.GetStudentCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get student count: %w", err)
	}

	stats, err := s.stats.GetAdminStatsData(ctx)
	if err != nil {
		return nil, fmt.Errorf("get admin stats: %w", err)
	}

	recentActivity, err := s.attempts.GetRecentAttempts(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("get recent attempts: %w", err)
	}

	recentAttempts := make([]*types.RecentAttempt, 0, len(recentActivity))
	for _, a := range recentActivity {
		completedAt := ""
		if a.CompletedAt.Valid {
			completedAt = a.CompletedAt.Time.Format("2006-01-02 15:04")
		}
		recentAttempts = append(recentAttempts, &types.RecentAttempt{
			AttemptID:   a.ID.String(),
			UserName:    a.UserName,
			QuizTitle:   a.QuizTitle,
			Score:       int(a.Score),
			MaxScore:    int(a.MaxScore),
			CompletedAt: completedAt,
		})
	}

	avgScore := extractFloat(stats.AvgScore)

	return &types.AdminDashboardData{
		User:           &user,
		QuizCount:      len(quizzes),
		StudentCount:   int(studentCount),
		AttemptCount:   int(stats.TotalAttempts),
		AvgScore:       avgScore,
		RecentActivity: recentAttempts,
	}, nil
}

func (s *AdminService) GetQuizzesListData(ctx context.Context, userID uuid.UUID) (*types.AdminQuizListData, error) {
	userCacheKey := "user:" + userID.String()
	user, err := GetOrFetch(ctx, s.cache, userCacheKey, func() (db.User, error) {
		return s.users.GetUserByID(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quizzes, err := s.quizzes.GetNonArchivedQuizzes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get quizzes: %w", err)
	}

	quizStats, err := s.stats.GetQuizStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get quiz stats: %w", err)
	}

	statsMap := make(map[string]types.QuizStatsResponse)
	for _, qs := range quizStats {
		statsMap[qs.QuizID.String()] = types.QuizStatsResponse{
			QuizID:       qs.QuizID,
			Title:        qs.Title,
			Subject:      qs.Subject,
			AttemptCount: int(qs.AttemptCount),
			AvgScore:     extractFloat(qs.AvgScore),
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

	return &types.AdminQuizListData{
		User:    &user,
		Quizzes: quizWithStats,
	}, nil
}

func (s *AdminService) CreateQuiz(ctx context.Context, userID uuid.UUID, title, description, subject string, grade, timeLimit int) (uuid.UUID, error) {
	newQuizID := uuid.New()
	cacheKey := "quiz:" + newQuizID.String()
	createdQuiz, err := SaveOrUpdate(ctx, s.cache, cacheKey, func() (db.Quiz, error) {
		return s.quizzes.CreateQuiz(ctx, db.CreateQuizParams{
			ID:          newQuizID,
			Title:       title,
			Description: description,
			Subject:     subject,
			Grade:       grade,
			Status:      db.QuizStatusAvailable,
			TimeLimit:   timeLimit,
			CreatedBy:   userID,
			CreatedAt:   time.Now(),
		})
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create quiz: %w", err)
	}

	_ = Delete(ctx, s.cache, "quizzes:user:"+userID.String(), func() error {
		return nil
	})

	return createdQuiz.ID, nil
}

func (s *AdminService) GetQuizEditData(ctx context.Context, userID, quizID uuid.UUID) (*types.AdminQuizEditData, error) {
	userCacheKey := "user:" + userID.String()
	user, err := GetOrFetch(ctx, s.cache, userCacheKey, func() (db.User, error) {
		return s.users.GetUserByID(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quizCacheKey := "quiz:" + quizID.String()
	quiz, err := GetOrFetch(ctx, s.cache, quizCacheKey, func() (db.Quiz, error) {
		return s.quizzes.GetQuizByID(ctx, quizID)
	})
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	questionsCacheKey := "questions:quiz:" + quizID.String()
	questions, err := GetOrFetch(ctx, s.cache, questionsCacheKey, func() ([]db.Question, error) {
		return s.questions.GetQuestionsByQuizID(ctx, quizID)
	})
	if err != nil {
		return nil, fmt.Errorf("get questions: %w", err)
	}

	questionsWithImages := s.attachImagesToQuestions(ctx, questions)

	quizWithQuestions := models.QuizWithQuestionsAndImages{
		Quiz:      quiz,
		Questions: questionsWithImages,
	}

	return &types.AdminQuizEditData{
		User:      &user,
		Quiz:      &quizWithQuestions,
		Questions: questionsWithImages,
	}, nil
}

func (s *AdminService) UpdateQuiz(ctx context.Context, quizID uuid.UUID, title, description, subject string, grade, timeLimit int, status db.QuizStatus) error {
	_, err := s.quizzes.GetQuizByID(ctx, quizID)
	if err != nil {
		return fmt.Errorf("quiz not found: %w", err)
	}

	_, err = s.quizzes.UpdateQuiz(ctx, db.UpdateQuizParams{
		ID:          quizID,
		Title:       title,
		Description: description,
		Subject:     subject,
		Grade:       grade,
		Status:      status,
		TimeLimit:   timeLimit,
	})
	if err != nil {
		return fmt.Errorf("update quiz: %w", err)
	}

	return nil
}

func (s *AdminService) DeleteQuiz(ctx context.Context, quizID uuid.UUID) error {
	cacheKey := "quiz:" + quizID.String()
	err := Delete(ctx, s.cache, cacheKey, func() error {
		return s.quizzes.DeleteQuiz(ctx, quizID)
	})
	if err != nil {
		return fmt.Errorf("delete quiz: %w", err)
	}

	_ = Delete(ctx, s.cache, "questions:quiz:"+quizID.String(), func() error { return nil })

	return nil
}

func (s *AdminService) AddQuestion(ctx context.Context, quizID uuid.UUID, text string, questionType db.QuestionType, options []string, correctAnswer, explanation string, points, orderIndex int, files []*multipart.FileHeader) (uuid.UUID, error) {
	if points == 0 {
		points = 10
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal options: %w", err)
	}

	newQuestionID := uuid.New()
	_, err = s.questions.CreateQuestion(ctx, db.CreateQuestionParams{
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
		return uuid.Nil, fmt.Errorf("create question: %w", err)
	}

	for i, fileHeader := range files {
		if i >= MaxImagesPerQuestion {
			break
		}
		count, err := s.images.GetImageCountByQuestionID(ctx, newQuestionID)
		if err != nil {
			continue
		}
		if count >= MaxImagesPerQuestion {
			continue
		}

		url, err := s.storageService.UploadImage(ctx, fileHeader)
		if err != nil {
			continue
		}

		_, err = s.images.CreateQuestionImage(ctx, db.CreateQuestionImageParams{
			ID:         uuid.New(),
			QuestionID: newQuestionID,
			Url:        url,
			OrderIndex: int(count),
			CreatedAt:  time.Now(),
		})
		if err != nil {
			_ = s.storageService.DeleteImage(ctx, url)
		}
	}

	return newQuestionID, nil
}

func (s *AdminService) UpdateQuestion(ctx context.Context, questionID, quizID uuid.UUID, text string, questionType db.QuestionType, options []byte, correctAnswer, explanation string, points, orderIndex int) error {
	question, err := s.questions.GetQuestionByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("get question: %w", err)
	}

	if question.QuizID != quizID {
		return fmt.Errorf("question does not belong to this quiz")
	}

	_, err = s.questions.UpdateQuestion(ctx, db.UpdateQuestionParams{
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
		return fmt.Errorf("update question: %w", err)
	}

	return nil
}

func (s *AdminService) DeleteQuestion(ctx context.Context, questionID, quizID uuid.UUID) error {
	question, err := s.questions.GetQuestionByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("get question: %w", err)
	}

	if question.QuizID != quizID {
		return fmt.Errorf("question does not belong to this quiz")
	}

	err = s.questions.DeleteQuestion(ctx, questionID)
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}

	return nil
}

func (s *AdminService) UploadQuestionImage(ctx context.Context, quizID, questionID uuid.UUID, file *multipart.FileHeader) error {
	question, err := s.questions.GetQuestionByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("get question: %w", err)
	}

	if question.QuizID != quizID {
		return fmt.Errorf("question does not belong to this quiz")
	}

	count, err := s.images.GetImageCountByQuestionID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("get image count: %w", err)
	}
	if count >= MaxImagesPerQuestion {
		return fmt.Errorf("maximum images reached")
	}

	contentType := file.Header.Get("Content-Type")
	if !strings.Contains(AllowedImageTypes, contentType) {
		return fmt.Errorf("invalid image type")
	}

	if file.Size > MaxImageSize {
		return fmt.Errorf("image too large")
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return fmt.Errorf("invalid file extension")
	}

	url, err := s.storageService.UploadImage(ctx, file)
	if err != nil {
		return fmt.Errorf("upload image: %w", err)
	}

	_, err = s.images.CreateQuestionImage(ctx, db.CreateQuestionImageParams{
		ID:         uuid.New(),
		QuestionID: questionID,
		Url:        url,
		OrderIndex: int(count),
		CreatedAt:  time.Now(),
	})
	if err != nil {
		_ = s.storageService.DeleteImage(ctx, url)
		return fmt.Errorf("create question image: %w", err)
	}

	return nil
}

func (s *AdminService) DeleteQuestionImage(ctx context.Context, imageID, questionID uuid.UUID) error {
	image, err := s.images.GetQuestionImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("get image: %w", err)
	}

	if image.QuestionID != questionID {
		return fmt.Errorf("image does not belong to this question")
	}

	err = s.images.DeleteQuestionImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	objectName := filepath.Base(image.Url)
	if err := s.storageService.DeleteImage(ctx, objectName); err != nil {
		return fmt.Errorf("delete from storage: %w", err)
	}

	return nil
}

func (s *AdminService) GetResultsData(ctx context.Context, userID uuid.UUID) (*types.AdminResultsData, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	attempts, err := s.attempts.GetRecentAttempts(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("get attempts: %w", err)
	}

	quizzes, err := s.quizzes.GetNonArchivedQuizzes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get quizzes: %w", err)
	}

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

	quizList := make([]*models.Quiz, len(quizzes))
	for i, q := range quizzes {
		quizList[i] = &models.Quiz{Quiz: q}
	}

	return &types.AdminResultsData{
		User:     &user,
		Attempts: attemptWithUser,
		Quizzes:  quizList,
	}, nil
}

func (s *AdminService) GetStatisticsData(ctx context.Context, userID uuid.UUID) (*types.AdminStatisticsData, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	stats, err := s.stats.GetAdminStatsData(ctx)
	if err != nil {
		return nil, fmt.Errorf("get admin stats: %w", err)
	}

	return &types.AdminStatisticsData{
		User:                &user,
		TotalQuizzes:        int(stats.TotalQuizzes),
		TotalStudents:       int(stats.TotalStudents),
		TotalAttempts:       int(stats.TotalAttempts),
		AvgScore:            extractFloat(stats.AvgScore),
		QuizStats:           nil,
		GradeDistribution:   nil,
		SubjectDistribution: nil,
	}, nil
}

func (s *AdminService) GetQuizStatsData(ctx context.Context) ([]types.QuizStatsResponse, error) {
	quizStats, err := s.stats.GetQuizStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get quiz stats: %w", err)
	}

	stats := make([]types.QuizStatsResponse, 0, len(quizStats))
	for _, qs := range quizStats {
		stats = append(stats, types.QuizStatsResponse{
			QuizID:       qs.QuizID,
			Title:        qs.Title,
			Subject:      qs.Subject,
			AttemptCount: int(qs.AttemptCount),
			AvgScore:     extractFloat(qs.AvgScore),
			PassRate:     float64(qs.PassRate),
		})
	}

	return stats, nil
}

func (s *AdminService) GetGradeDistributionData(ctx context.Context) (map[string]int, error) {
	data, err := s.stats.GetGradeDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("get grade distribution: %w", err)
	}

	var dist map[string]int
	if data != nil {
		if err := json.Unmarshal(data, &dist); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
	}
	if dist == nil {
		dist = make(map[string]int)
	}

	return dist, nil
}

func (s *AdminService) GetSubjectDistributionData(ctx context.Context) (map[string]int, error) {
	data, err := s.stats.GetSubjectDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("get subject distribution: %w", err)
	}

	var dist map[string]int
	if data != nil {
		if err := json.Unmarshal(data, &dist); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
	}
	if dist == nil {
		dist = make(map[string]int)
	}

	return dist, nil
}

func (s *AdminService) RestoreQuiz(ctx context.Context, quizID uuid.UUID) error {
	err := s.quizzes.UpdateQuizStatus(ctx, db.UpdateQuizStatusParams{
		ID:     quizID,
		Status: db.QuizStatusAvailable,
	})
	if err != nil {
		return fmt.Errorf("restore quiz: %w", err)
	}
	return nil
}

func (s *AdminService) attachImagesToQuestions(ctx context.Context, questions []db.Question) []models.Question {
	return AttachImagesToQuestions(ctx, questions, s.images)
}

func (s *AdminService) GetQuestionsByQuizID(ctx context.Context, quizID uuid.UUID) ([]models.Question, error) {
	cacheKey := "questions:quiz:" + quizID.String()
	questions, err := GetOrFetch(ctx, s.cache, cacheKey, func() ([]db.Question, error) {
		return s.questions.GetQuestionsByQuizID(ctx, quizID)
	})
	if err != nil {
		return nil, fmt.Errorf("get questions: %w", err)
	}
	return s.attachImagesToQuestions(ctx, questions), nil
}

func (s *AdminService) GetQuizByID(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error) {
	cacheKey := "quiz:" + quizID.String()
	quiz, err := GetOrFetch(ctx, s.cache, cacheKey, func() (db.Quiz, error) {
		return s.quizzes.GetQuizByID(ctx, quizID)
	})
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}
	return &models.Quiz{Quiz: quiz}, nil
}

func extractFloat(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	default:
		if num, ok := val.(interface{ Float64() (float64, error) }); ok {
			f, _ := num.Float64()
			return f
		}
	}
	return 0.0
}

const (
	MaxImagesPerQuestion = 3
	MaxImageSize         = 5 << 20
	AllowedImageTypes    = "image/jpeg,image/png,image/webp"
)

func ParseQuestionForm(r *http.Request) (text, questionTypeStr, explanation, correctAnswer string, points, orderIndex int, options []string, files []*multipart.FileHeader, err error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			return "", "", "", "", 0, 0, nil, nil, fmt.Errorf("parse form: %w", err)
		}
	}

	text = r.FormValue("text")
	questionTypeStr = r.FormValue("type")
	if questionTypeStr == "" {
		questionTypeStr = "choice"
	}
	explanation = r.FormValue("explanation")
	correctAnswer = r.FormValue("correct_answer")
	points, _ = strconv.Atoi(r.FormValue("points"))
	orderIndex, _ = strconv.Atoi(r.FormValue("order_index"))

	if points == 0 {
		points = 10
	}

	var opts []string
	if questionTypeStr == "choice" || questionTypeStr == string(db.QuestionTypeChoice) {
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("option_%d", i)
			if val, ok := r.Form[key]; ok && val[0] != "" {
				opts = append(opts, val[0])
			}
		}
		correctAnswerRaw := r.FormValue("correct_answer")
		if idx, err := strconv.Atoi(strings.TrimPrefix(correctAnswerRaw, "option_")); err == nil && idx >= 0 && idx < len(opts) {
			correctAnswer = opts[idx]
		}
	}

	options = opts
	files = nil
	if mpForm := r.MultipartForm; mpForm != nil {
		files = mpForm.File["images"]
	}

	return text, questionTypeStr, explanation, correctAnswer, points, orderIndex, options, files, nil
}
