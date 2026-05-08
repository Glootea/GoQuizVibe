package services_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	mocks "github.com/goquizvibe/mocks/services"
	"github.com/goquizvibe/mocks/servicestest"
	"github.com/goquizvibe/services"
	"github.com/minio/minio-go/v7"
	"go.uber.org/mock/gomock"
)

func TestNewAdminService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)
	authSvc := &services.AuthService{}
	storageSvc := &services.StorageService{}

	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, authSvc, storageSvc)

	if svc == nil {
		t.Fatal("NewAdminService returned nil")
	}
}

func TestAdminService_GetUserFromRequest(t *testing.T) {
	t.Run("no cookie", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		authSvc := &services.AuthService{}
		storageSvc := &services.StorageService{}
		svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, authSvc, storageSvc)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := svc.GetUserFromRequest(req)
		if err == nil {
			t.Fatal("expected error when no cookie")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		authSvc := services.NewAuthService(mockUsers, "test-secret", time.Hour*24)
		storageSvc := &services.StorageService{}
		svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, authSvc, storageSvc)

		userID := uuid.New()
		user := db.User{ID: userID, Name: "Test User", Email: "test@example.com"}

		token, _ := authSvc.GenerateToken(&user)

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: token})

		result, err := svc.GetUserFromRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != userID {
			t.Errorf("result.ID = %v, want %v", result.ID, userID)
		}
	})
}

func TestAdminService_GetDashboardData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get user error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{}, errors.New("db error"))

		_, err := svc.GetDashboardData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get quizzes error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Test"}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return(nil, errors.New("quizzes error"))

		_, err := svc.GetDashboardData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get student count error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Test"}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: uuid.New()}}, nil)
		mockUsers.EXPECT().GetStudentCount(gomock.Any()).Return(int64(0), errors.New("count error"))

		_, err := svc.GetDashboardData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get stats error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Test"}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: uuid.New()}}, nil)
		mockUsers.EXPECT().GetStudentCount(gomock.Any()).Return(int64(10), nil)
		mockStats.EXPECT().GetAdminStatsData(gomock.Any()).Return(db.GetAdminStatsDataRow{}, errors.New("stats error"))

		_, err := svc.GetDashboardData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get recent attempts error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Test"}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: uuid.New()}}, nil)
		mockUsers.EXPECT().GetStudentCount(gomock.Any()).Return(int64(10), nil)
		mockStats.EXPECT().GetAdminStatsData(gomock.Any()).Return(db.GetAdminStatsDataRow{TotalAttempts: 5, AvgScore: 75.5}, nil)
		mockAttempts.EXPECT().GetRecentAttempts(gomock.Any(), int32(10)).Return(nil, errors.New("attempts error"))

		_, err := svc.GetDashboardData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		user := db.User{ID: userID, Name: "Test User"}
		quizID := uuid.New()
		attemptID := uuid.New()

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: quizID, Title: "Test Quiz"}}, nil)
		mockUsers.EXPECT().GetStudentCount(gomock.Any()).Return(int64(25), nil)
		mockStats.EXPECT().GetAdminStatsData(gomock.Any()).Return(db.GetAdminStatsDataRow{TotalAttempts: 100, AvgScore: 85.5}, nil)
		mockAttempts.EXPECT().GetRecentAttempts(gomock.Any(), int32(10)).Return([]db.GetRecentAttemptsRow{
			{ID: attemptID, UserName: "Student 1", QuizTitle: "Quiz 1", Score: 80, MaxScore: 100, CompletedAt: sql.NullTime{Time: time.Now(), Valid: true}},
		}, nil)

		data, err := svc.GetDashboardData(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data.QuizCount != 1 {
			t.Errorf("QuizCount = %d, want 1", data.QuizCount)
		}
		if data.StudentCount != 25 {
			t.Errorf("StudentCount = %d, want 25", data.StudentCount)
		}
		if data.AttemptCount != 100 {
			t.Errorf("AttemptCount = %d, want 100", data.AttemptCount)
		}
	})
}

func TestAdminService_GetQuizzesListData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get user error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{}, errors.New("db error"))

		_, err := svc.GetQuizzesListData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get quizzes error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return(nil, errors.New("quizzes error"))

		_, err := svc.GetQuizzesListData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get quiz stats error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: uuid.New()}}, nil)
		mockStats.EXPECT().GetQuizStats(gomock.Any()).Return(nil, errors.New("stats error"))

		_, err := svc.GetQuizzesListData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		quizID := uuid.New()

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Admin"}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: quizID, Title: "Test Quiz", Subject: "Math"}}, nil)
		mockStats.EXPECT().GetQuizStats(gomock.Any()).Return([]db.GetQuizStatsRow{
			{QuizID: quizID, Title: "Test Quiz", Subject: "Math", AttemptCount: 50, AvgScore: 88.5, PassRate: 1},
		}, nil)

		data, err := svc.GetQuizzesListData(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data.Quizzes) != 1 {
			t.Errorf("len(Quizzes) = %d, want 1", len(data.Quizzes))
		}
		if data.Quizzes[0].AttemptCount != 50 {
			t.Errorf("AttemptCount = %d, want 50", data.Quizzes[0].AttemptCount)
		}
	})
}

func TestAdminService_CreateQuiz(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("create quiz error", func(t *testing.T) {
		userID := uuid.New()
		mockQuizzes.EXPECT().CreateQuiz(gomock.Any(), gomock.Any()).Return(db.Quiz{}, errors.New("create error"))

		_, err := svc.CreateQuiz(context.Background(), userID, "Test Quiz", "Description", "Math", 5, 30)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		expectedQuiz := db.Quiz{ID: uuid.New(), Title: "Test Quiz"}

		mockQuizzes.EXPECT().CreateQuiz(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, params db.CreateQuizParams) (db.Quiz, error) {
				if params.Title != "Test Quiz" {
					t.Errorf("title = %q, want %q", params.Title, "Test Quiz")
				}
				if params.Subject != "Math" {
					t.Errorf("subject = %q, want %q", params.Subject, "Math")
				}
				if params.Grade != 5 {
					t.Errorf("grade = %d, want %d", params.Grade, 5)
				}
				if params.TimeLimit != 30 {
					t.Errorf("timeLimit = %d, want %d", params.TimeLimit, 30)
				}
				return expectedQuiz, nil
			})

		quizID, err := svc.CreateQuiz(context.Background(), userID, "Test Quiz", "Description", "Math", 5, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if quizID == uuid.Nil {
			t.Error("quizID is nil")
		}
	})
}

func TestAdminService_GetQuizEditData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get user error", func(t *testing.T) {
		userID, quizID := uuid.New(), uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{}, errors.New("db error"))

		_, err := svc.GetQuizEditData(context.Background(), userID, quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get quiz error", func(t *testing.T) {
		userID, quizID := uuid.New(), uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(db.Quiz{}, errors.New("quiz not found"))

		_, err := svc.GetQuizEditData(context.Background(), userID, quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get questions error", func(t *testing.T) {
		userID, quizID := uuid.New(), uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(db.Quiz{ID: quizID}, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quizID).Return(nil, errors.New("questions error"))

		_, err := svc.GetQuizEditData(context.Background(), userID, quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success with images", func(t *testing.T) {
		userID, quizID, questionID := uuid.New(), uuid.New(), uuid.New()
		user := db.User{ID: userID, Name: "Admin"}
		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		question := db.Question{ID: questionID, QuizID: quizID, Text: "Question 1"}
		images := []db.QuestionImage{{ID: uuid.New(), QuestionID: questionID, Url: "http://img.jpg"}}

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quizID).Return([]db.Question{question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(gomock.Any(), questionID).Return(images, nil)

		data, err := svc.GetQuizEditData(context.Background(), userID, quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data.Questions) != 1 {
			t.Errorf("len(Questions) = %d, want 1", len(data.Questions))
		}
		if len(data.Questions[0].Images) != 1 {
			t.Errorf("len(Images) = %d, want 1", len(data.Questions[0].Images))
		}
	})

	t.Run("get images error continues", func(t *testing.T) {
		userID, quizID, questionID := uuid.New(), uuid.New(), uuid.New()
		user := db.User{ID: userID, Name: "Admin"}
		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		question := db.Question{ID: questionID, QuizID: quizID, Text: "Question 1"}

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quizID).Return([]db.Question{question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(gomock.Any(), questionID).Return(nil, errors.New("images error"))

		data, err := svc.GetQuizEditData(context.Background(), userID, quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data.Questions) != 1 {
			t.Errorf("len(Questions) = %d, want 1", len(data.Questions))
		}
		if data.Questions[0].Images != nil {
			t.Errorf("Images = %v, want nil", data.Questions[0].Images)
		}
	})

	t.Run("success without images", func(t *testing.T) {
		userID, quizID := uuid.New(), uuid.New()
		user := db.User{ID: userID, Name: "Admin"}
		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quizID).Return([]db.Question{}, nil)

		data, err := svc.GetQuizEditData(context.Background(), userID, quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data.User.Name != "Admin" {
			t.Errorf("User.Name = %q, want %q", data.User.Name, "Admin")
		}
	})
}

func TestAdminService_UpdateQuiz(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("quiz not found", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(db.Quiz{}, errors.New("not found"))

		err := svc.UpdateQuiz(context.Background(), quizID, "Title", "Desc", "Math", 5, 30, db.QuizStatusAvailable)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("update error", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(db.Quiz{ID: quizID}, nil)
		mockQuizzes.EXPECT().UpdateQuiz(gomock.Any(), gomock.Any()).Return(db.Quiz{}, errors.New("update error"))

		err := svc.UpdateQuiz(context.Background(), quizID, "Title", "Desc", "Math", 5, 30, db.QuizStatusAvailable)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID := uuid.New()
		updatedQuiz := db.Quiz{ID: quizID, Title: "Updated Title"}

		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(db.Quiz{ID: quizID}, nil)
		mockQuizzes.EXPECT().UpdateQuiz(gomock.Any(), gomock.Any()).Return(updatedQuiz, nil)

		err := svc.UpdateQuiz(context.Background(), quizID, "Updated Title", "Desc", "Math", 5, 30, db.QuizStatusAvailable)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_DeleteQuiz(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("delete error", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().DeleteQuiz(gomock.Any(), quizID).Return(errors.New("delete error"))

		err := svc.DeleteQuiz(context.Background(), quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().DeleteQuiz(gomock.Any(), quizID).Return(nil)

		err := svc.DeleteQuiz(context.Background(), quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_AddQuestion(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)
	mockMinio := servicestest.NewMockMinioClient(ctrl)

	storageSvc := services.NewStorageService(mockMinio, "test-bucket")
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("create question error", func(t *testing.T) {
		quizID := uuid.New()
		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{}, errors.New("create error"))

		_, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success without images", func(t *testing.T) {
		quizID := uuid.New()

		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: uuid.New(), QuizID: quizID}, nil)

		id, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("id is nil")
		}
	})

	t.Run("get image count error continues", func(t *testing.T) {
		quizID := uuid.New()

		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: uuid.New(), QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("count error"))

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("fake"))

		id, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, []*multipart.FileHeader{file})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("id is nil")
		}
	})

	t.Run("success with images", func(t *testing.T) {
		quizID := uuid.New()

		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: uuid.New(), QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), gomock.Any()).Return(int64(0), nil)
		mockMinio.EXPECT().PutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, nil)
		mockMinio.EXPECT().PresignedGetObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any()).Return(&url.URL{}, nil)
		mockImages.EXPECT().CreateQuestionImage(gomock.Any(), gomock.Any()).Return(db.QuestionImage{}, nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("fake image content"))

		id, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, []*multipart.FileHeader{file})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("id is nil")
		}
	})

	t.Run("max images reached", func(t *testing.T) {
		quizID := uuid.New()

		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: uuid.New(), QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), gomock.Any()).Return(int64(3), nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("fake"))

		id, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, []*multipart.FileHeader{file})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("id is nil")
		}
	})

	t.Run("upload image error continues", func(t *testing.T) {
		quizID := uuid.New()

		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: uuid.New(), QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), gomock.Any()).Return(int64(0), nil)
		mockMinio.EXPECT().PutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, errors.New("upload error"))

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("fake"))

		id, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, []*multipart.FileHeader{file})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("id is nil")
		}
	})

	t.Run("create image error rolls back", func(t *testing.T) {
		quizID := uuid.New()

		mockQuestions.EXPECT().CreateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: uuid.New(), QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), gomock.Any()).Return(int64(0), nil)
		mockMinio.EXPECT().PutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, nil)
		mockMinio.EXPECT().PresignedGetObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any()).Return(&url.URL{}, nil)
		mockImages.EXPECT().CreateQuestionImage(gomock.Any(), gomock.Any()).Return(db.QuestionImage{}, errors.New("db error"))
		mockMinio.EXPECT().RemoveObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any()).Return(nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("fake"))

		id, err := svc.AddQuestion(context.Background(), quizID, "Question?", db.QuestionTypeChoice, []string{"A", "B"}, "A", "Explanation", 10, 0, []*multipart.FileHeader{file})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("id is nil")
		}
	})
}

func createTestFileHeader(filename, contentType string, content []byte) *multipart.FileHeader {
	body := "--boundary\r\n" +
		`Content-Disposition: form-data; name="file"; filename="` + filename + `"` + "\r\n" +
		"Content-Type: " + contentType + "\r\n\r\n" +
		string(content) + "\r\n" +
		"--boundary--\r\n"

	reader := multipart.NewReader(strings.NewReader(body), "boundary")
	form, _ := reader.ReadForm(32 << 20)
	if len(form.File["file"]) > 0 {
		return form.File["file"][0]
	}

	return &multipart.FileHeader{Filename: filename}
}

func TestAdminService_UpdateQuestion(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get question error", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{}, errors.New("not found"))

		err := svc.UpdateQuestion(context.Background(), questionID, quizID, "Updated?", db.QuestionTypeChoice, nil, "A", "Expl", 10, 0)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("question belongs to different quiz", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		differentQuizID := uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: differentQuizID}, nil)

		err := svc.UpdateQuestion(context.Background(), questionID, quizID, "Updated?", db.QuestionTypeChoice, nil, "A", "Expl", 10, 0)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("update error", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockQuestions.EXPECT().UpdateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{}, errors.New("update error"))

		err := svc.UpdateQuestion(context.Background(), questionID, quizID, "Updated?", db.QuestionTypeChoice, nil, "A", "Expl", 10, 0)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockQuestions.EXPECT().UpdateQuestion(gomock.Any(), gomock.Any()).Return(db.Question{ID: questionID}, nil)

		options, _ := json.Marshal([]string{"A", "B"})
		err := svc.UpdateQuestion(context.Background(), questionID, quizID, "Updated?", db.QuestionTypeChoice, options, "A", "Expl", 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_DeleteQuestion(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get question error", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{}, errors.New("not found"))

		err := svc.DeleteQuestion(context.Background(), questionID, quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("question belongs to different quiz", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		differentQuizID := uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: differentQuizID}, nil)

		err := svc.DeleteQuestion(context.Background(), questionID, quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("delete error", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockQuestions.EXPECT().DeleteQuestion(gomock.Any(), questionID).Return(errors.New("delete error"))

		err := svc.DeleteQuestion(context.Background(), questionID, quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		questionID, quizID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockQuestions.EXPECT().DeleteQuestion(gomock.Any(), questionID).Return(nil)

		err := svc.DeleteQuestion(context.Background(), questionID, quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_UploadQuestionImage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)
	mockMinio := servicestest.NewMockMinioClient(ctrl)

	storageSvc := services.NewStorageService(mockMinio, "test-bucket")
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get question error", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{}, errors.New("not found"))

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("question belongs to different quiz", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		differentQuizID := uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: differentQuizID}, nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("max images reached", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(3), nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid content type", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(0), nil)

		file := createTestFileHeader("test.gif", "image/gif", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("image too large", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(0), nil)

		content := make([]byte, 6<<20+1)
		file := createTestFileHeader("test.jpg", "image/jpeg", content)
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(0), nil)

		file := createTestFileHeader("test.bmp", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("upload error", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(0), nil)
		mockMinio.EXPECT().PutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, errors.New("upload failed"))

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("create image error", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(0), nil)
		mockMinio.EXPECT().PutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, nil)
		mockMinio.EXPECT().PresignedGetObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any()).Return(&url.URL{}, nil)
		mockImages.EXPECT().CreateQuestionImage(gomock.Any(), gomock.Any()).Return(db.QuestionImage{}, errors.New("db error"))
		mockMinio.EXPECT().RemoveObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any()).Return(nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		mockQuestions.EXPECT().GetQuestionByID(gomock.Any(), questionID).Return(db.Question{ID: questionID, QuizID: quizID}, nil)
		mockImages.EXPECT().GetImageCountByQuestionID(gomock.Any(), questionID).Return(int64(0), nil)
		mockMinio.EXPECT().PutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, nil)
		mockMinio.EXPECT().PresignedGetObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), gomock.Any()).Return(&url.URL{}, nil)
		mockImages.EXPECT().CreateQuestionImage(gomock.Any(), gomock.Any()).Return(db.QuestionImage{ID: uuid.New()}, nil)

		file := createTestFileHeader("test.jpg", "image/jpeg", []byte("content"))
		err := svc.UploadQuestionImage(context.Background(), quizID, questionID, file)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_DeleteQuestionImage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)
	mockMinio := servicestest.NewMockMinioClient(ctrl)

	storageSvc := services.NewStorageService(mockMinio, "test-bucket")
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get image error", func(t *testing.T) {
		imageID, questionID := uuid.New(), uuid.New()
		mockImages.EXPECT().GetQuestionImageByID(gomock.Any(), imageID).Return(db.QuestionImage{}, errors.New("not found"))

		err := svc.DeleteQuestionImage(context.Background(), imageID, questionID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("image belongs to different question", func(t *testing.T) {
		imageID, questionID := uuid.New(), uuid.New()
		differentQuestionID := uuid.New()
		mockImages.EXPECT().GetQuestionImageByID(gomock.Any(), imageID).Return(db.QuestionImage{ID: imageID, QuestionID: differentQuestionID}, nil)

		err := svc.DeleteQuestionImage(context.Background(), imageID, questionID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("delete from db error", func(t *testing.T) {
		imageID, questionID := uuid.New(), uuid.New()
		mockImages.EXPECT().GetQuestionImageByID(gomock.Any(), imageID).Return(db.QuestionImage{ID: imageID, QuestionID: questionID, Url: "http://storage/img.jpg"}, nil)
		mockImages.EXPECT().DeleteQuestionImage(gomock.Any(), imageID).Return(errors.New("delete error"))

		err := svc.DeleteQuestionImage(context.Background(), imageID, questionID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("delete from storage error", func(t *testing.T) {
		imageID, questionID := uuid.New(), uuid.New()
		mockImages.EXPECT().GetQuestionImageByID(gomock.Any(), imageID).Return(db.QuestionImage{ID: imageID, QuestionID: questionID, Url: "http://storage/img.jpg"}, nil)
		mockImages.EXPECT().DeleteQuestionImage(gomock.Any(), imageID).Return(nil)
		mockMinio.EXPECT().RemoveObject(gomock.Any(), "test-bucket", "img.jpg", gomock.Any()).Return(errors.New("storage error"))

		err := svc.DeleteQuestionImage(context.Background(), imageID, questionID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		imageID, questionID := uuid.New(), uuid.New()
		mockImages.EXPECT().GetQuestionImageByID(gomock.Any(), imageID).Return(db.QuestionImage{ID: imageID, QuestionID: questionID, Url: "http://storage/img.jpg"}, nil)
		mockImages.EXPECT().DeleteQuestionImage(gomock.Any(), imageID).Return(nil)
		mockMinio.EXPECT().RemoveObject(gomock.Any(), "test-bucket", "img.jpg", gomock.Any()).Return(nil)

		err := svc.DeleteQuestionImage(context.Background(), imageID, questionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_GetResultsData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get user error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{}, errors.New("not found"))

		_, err := svc.GetResultsData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get attempts error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockAttempts.EXPECT().GetRecentAttempts(gomock.Any(), int32(0)).Return(nil, errors.New("attempts error"))

		_, err := svc.GetResultsData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get quizzes error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockAttempts.EXPECT().GetRecentAttempts(gomock.Any(), int32(0)).Return([]db.GetRecentAttemptsRow{}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return(nil, errors.New("quizzes error"))

		_, err := svc.GetResultsData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		attemptID := uuid.New()
		quizID := uuid.New()

		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Admin"}, nil)
		mockAttempts.EXPECT().GetRecentAttempts(gomock.Any(), int32(0)).Return([]db.GetRecentAttemptsRow{
			{ID: attemptID, UserID: userID, QuizID: quizID, Score: 80, MaxScore: 100, UserName: "Student", QuizTitle: "Quiz 1"},
		}, nil)
		mockQuizzes.EXPECT().GetNonArchivedQuizzes(gomock.Any()).Return([]db.Quiz{{ID: quizID, Title: "Quiz 1"}}, nil)

		data, err := svc.GetResultsData(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data.Attempts) != 1 {
			t.Errorf("len(Attempts) = %d, want 1", len(data.Attempts))
		}
		if len(data.Quizzes) != 1 {
			t.Errorf("len(Quizzes) = %d, want 1", len(data.Quizzes))
		}
	})
}

func TestAdminService_GetStatisticsData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get user error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{}, errors.New("not found"))

		_, err := svc.GetStatisticsData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get stats error", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID}, nil)
		mockStats.EXPECT().GetAdminStatsData(gomock.Any()).Return(db.GetAdminStatsDataRow{}, errors.New("stats error"))

		_, err := svc.GetStatisticsData(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		mockUsers.EXPECT().GetUserByID(gomock.Any(), userID).Return(db.User{ID: userID, Name: "Admin"}, nil)
		mockStats.EXPECT().GetAdminStatsData(gomock.Any()).Return(db.GetAdminStatsDataRow{
			TotalQuizzes:  10,
			TotalStudents: 50,
			TotalAttempts: 200,
			AvgScore:      85.5,
		}, nil)

		data, err := svc.GetStatisticsData(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data.TotalQuizzes != 10 {
			t.Errorf("TotalQuizzes = %d, want 10", data.TotalQuizzes)
		}
		if data.TotalStudents != 50 {
			t.Errorf("TotalStudents = %d, want 50", data.TotalStudents)
		}
		if data.TotalAttempts != 200 {
			t.Errorf("TotalAttempts = %d, want 200", data.TotalAttempts)
		}
		if data.AvgScore != 85.5 {
			t.Errorf("AvgScore = %v, want 85.5", data.AvgScore)
		}
	})
}

func TestAdminService_GetQuizStatsData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get quiz stats error", func(t *testing.T) {
		mockStats.EXPECT().GetQuizStats(gomock.Any()).Return(nil, errors.New("stats error"))

		_, err := svc.GetQuizStatsData(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID := uuid.New()
		mockStats.EXPECT().GetQuizStats(gomock.Any()).Return([]db.GetQuizStatsRow{
			{QuizID: quizID, Title: "Quiz 1", Subject: "Math", AttemptCount: 50, AvgScore: 88.5, PassRate: 1},
		}, nil)

		stats, err := svc.GetQuizStatsData(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(stats) != 1 {
			t.Errorf("len(stats) = %d, want 1", len(stats))
		}
		if stats[0].AvgScore != 88.5 {
			t.Errorf("AvgScore = %v, want 88.5", stats[0].AvgScore)
		}
	})
}

func TestAdminService_GetGradeDistributionData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get error", func(t *testing.T) {
		mockStats.EXPECT().GetGradeDistribution(gomock.Any()).Return(nil, errors.New("stats error"))

		_, err := svc.GetGradeDistributionData(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		mockStats.EXPECT().GetGradeDistribution(gomock.Any()).Return([]byte("invalid json"), nil)

		_, err := svc.GetGradeDistributionData(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success with data", func(t *testing.T) {
		dist := map[string]int{"A": 10, "B": 20, "C": 15}
		data, _ := json.Marshal(dist)
		mockStats.EXPECT().GetGradeDistribution(gomock.Any()).Return(data, nil)

		result, err := svc.GetGradeDistributionData(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["A"] != 10 {
			t.Errorf("A = %d, want 10", result["A"])
		}
	})

	t.Run("success with nil data", func(t *testing.T) {
		mockStats.EXPECT().GetGradeDistribution(gomock.Any()).Return(nil, nil)

		result, err := svc.GetGradeDistributionData(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("result is nil")
		}
	})
}

func TestAdminService_GetSubjectDistributionData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get error", func(t *testing.T) {
		mockStats.EXPECT().GetSubjectDistribution(gomock.Any()).Return(nil, errors.New("stats error"))

		_, err := svc.GetSubjectDistributionData(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		mockStats.EXPECT().GetSubjectDistribution(gomock.Any()).Return([]byte("invalid json"), nil)

		_, err := svc.GetSubjectDistributionData(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		dist := map[string]int{"Math": 15, "Science": 25}
		data, _ := json.Marshal(dist)
		mockStats.EXPECT().GetSubjectDistribution(gomock.Any()).Return(data, nil)

		result, err := svc.GetSubjectDistributionData(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["Math"] != 15 {
			t.Errorf("Math = %d, want 15", result["Math"])
		}
	})
}

func TestAdminService_RestoreQuiz(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("restore error", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().UpdateQuizStatus(gomock.Any(), gomock.Any()).Return(errors.New("restore error"))

		err := svc.RestoreQuiz(context.Background(), quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().UpdateQuizStatus(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, params db.UpdateQuizStatusParams) error {
				if params.Status != db.QuizStatusAvailable {
					t.Errorf("status = %v, want %v", params.Status, db.QuizStatusAvailable)
				}
				return nil
			})

		err := svc.RestoreQuiz(context.Background(), quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAdminService_GetQuestionsByQuizID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get questions error", func(t *testing.T) {
		quizID := uuid.New()
		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quizID).Return(nil, errors.New("not found"))

		_, err := svc.GetQuestionsByQuizID(context.Background(), quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID, questionID := uuid.New(), uuid.New()
		questions := []db.Question{{ID: questionID, QuizID: quizID}}

		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quizID).Return(questions, nil)
		mockImages.EXPECT().GetImagesByQuestionID(gomock.Any(), questionID).Return(nil, nil)

		result, err := svc.GetQuestionsByQuizID(context.Background(), quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("len(result) = %d, want 1", len(result))
		}
	})
}

func TestAdminService_GetQuizByID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	storageSvc := &services.StorageService{}
	svc := services.NewAdminService(mockUsers, mockQuizzes, mockQuestions, mockImages, mockAttempts, mockStats, nil, storageSvc)

	t.Run("get quiz error", func(t *testing.T) {
		quizID := uuid.New()
		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(db.Quiz{}, errors.New("not found"))

		_, err := svc.GetQuizByID(context.Background(), quizID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		quizID := uuid.New()
		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}

		mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quizID).Return(quiz, nil)

		result, err := svc.GetQuizByID(context.Background(), quizID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Quiz.Title != "Test Quiz" {
			t.Errorf("Title = %q, want %q", result.Quiz.Title, "Test Quiz")
		}
	})
}

func TestParseQuestionForm_ChoiceType(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, qType, _, _, _, _, _, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
	if qType != "choice" {
		t.Errorf("qType = %q, want %q", qType, "choice")
	}
}

func TestParseQuestionForm_WithValues(t *testing.T) {
	t.Parallel()
	form := "text=What+is+2%2B2%3F&type=choice&explanation=Basic+math&correct_answer=4&points=10&order_index=0"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	text, qType, explanation, _, points, _, _, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
	if text != "What is 2+2?" {
		t.Errorf("text = %q, want %q", text, "What is 2+2?")
	}
	if qType != "choice" {
		t.Errorf("qType = %q, want %q", qType, "choice")
	}
	if explanation != "Basic math" {
		t.Errorf("explanation = %q, want %q", explanation, "Basic math")
	}
	if points != 10 {
		t.Errorf("points = %d, want %d", points, 10)
	}
}

func TestParseQuestionForm_DefaultPoints(t *testing.T) {
	t.Parallel()
	form := "text=Question+text&points=0"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, _, _, _, points, _, _, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
	if points != 10 {
		t.Errorf("points = %d, want default %d", points, 10)
	}
}

func TestParseQuestionForm_ChoiceWithOptions(t *testing.T) {
	t.Parallel()
	form := "text=Select+answer&type=choice&option_0=Option+A&option_1=Option+B&option_2=Option+C&correct_answer=option_1"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, _, _, correctAnswer, _, _, options, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
	if len(options) != 3 {
		t.Errorf("len(options) = %d, want %d", len(options), 3)
	}
	if options[0] != "Option A" {
		t.Errorf("options[0] = %q, want %q", options[0], "Option A")
	}
	if correctAnswer != "Option B" {
		t.Errorf("correctAnswer = %q, want %q", correctAnswer, "Option B")
	}
}

func TestParseQuestionForm_OpenType(t *testing.T) {
	t.Parallel()
	form := "text=What+is+the+capital+of+France%3F&type=open&correct_answer=Paris"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, _, _, correctAnswer, _, _, options, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
	if correctAnswer != "Paris" {
		t.Errorf("correctAnswer = %q, want %q", correctAnswer, "Paris")
	}
	if len(options) != 0 {
		t.Errorf("len(options) = %d, want 0 for open type", len(options))
	}
}

func TestParseQuestionForm_FillType(t *testing.T) {
	t.Parallel()
	form := "text=Complete+the+sentence&type=fill&correct_answer=hello"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, qType, _, correctAnswer, _, _, _, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
	if qType != "fill" {
		t.Errorf("qType = %q, want %q", qType, "fill")
	}
	if correctAnswer != "hello" {
		t.Errorf("correctAnswer = %q, want %q", correctAnswer, "hello")
	}
}

func TestParseQuestionForm_ParsesMultipartForm(t *testing.T) {
	t.Parallel()
	boundary := "----WebKitFormBoundary"
	body := "--" + boundary + "\r\n" +
		`Content-Disposition: form-data; name="text"` + "\r\n\r\n" +
		"Test question\r\n" +
		"--" + boundary + "\r\n" +
		`Content-Disposition: form-data; name="type"` + "\r\n\r\n" +
		"open\r\n" +
		"--" + boundary + "--\r\n"

	req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(body)))
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	_, _, _, _, _, _, _, _, err := services.ParseQuestionForm(req)
	if err != nil {
		t.Fatalf("ParseQuestionForm() error = %v, want nil", err)
	}
}
