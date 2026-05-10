package services_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	mocks "github.com/goquizvibe/mocks/services"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/services"
	"go.uber.org/mock/gomock"
)

func TestQuizSessionService_NewQuizSessionService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAttempts := mocks.NewMockAttemptRepository(ctrl)
	mockSessions := mocks.NewMockSessionRepository(ctrl)
	mockQuizzes := mocks.NewMockQuizRepository(ctrl)
	mockQuestions := mocks.NewMockQuestionRepository(ctrl)
	mockImages := mocks.NewMockImageRepository(ctrl)
	mockUsers := mocks.NewMockUserRepository(ctrl)
	mockStats := mocks.NewMockStatsRepository(ctrl)

	gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

	svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
	if svc == nil {
		t.Error("NewQuizSessionService() returned nil")
	}
}

func expectGetQuizWithQuestions(ctx context.Context, mockQuizzes *mocks.MockQuizRepository, mockQuestions *mocks.MockQuestionRepository, mockImages *mocks.MockImageRepository, quiz *models.QuizWithQuestionsAndImages) {
	mockQuizzes.EXPECT().GetQuizByID(ctx, quiz.Quiz.ID).Return(quiz.Quiz, nil).MaxTimes(3)
	mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quiz.Quiz.ID).Return([]db.Question{quiz.Questions[0].Question}, nil).MaxTimes(2)
	mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil).MaxTimes(2)
}

func TestQuizSessionService_CreateSession(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	quizID := uuid.New()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{ID: quizID, TimeLimit: 300, QuestionPoolSize: 0}, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSessions.EXPECT().CreateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		session, err := svc.CreateSession(ctx, userID, quizID)
		if err != nil {
			t.Fatalf("CreateSession() error = %v, want nil", err)
		}
		if session == nil {
			t.Fatal("CreateSession() returned nil session")
		}
	})

	t.Run("create attempt error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("attempt error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CreateSession(ctx, userID, quizID)
		if err == nil {
			t.Fatal("CreateSession() error = nil, want error")
		}
	})

	t.Run("get quiz error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("quiz error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CreateSession(ctx, userID, quizID)
		if err == nil {
			t.Fatal("CreateSession() error = nil, want error")
		}
	})

	t.Run("create session error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{ID: quizID, TimeLimit: 300, QuestionPoolSize: 0}, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSessions.EXPECT().CreateSession(ctx, gomock.Any()).Return(db.QuizSession{}, errors.New("session error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CreateSession(ctx, userID, quizID)
		if err == nil {
			t.Fatal("CreateSession() error = nil, want error")
		}
	})
}

func TestQuizSessionService_NavigateQuestion(t *testing.T) {
	ctx := context.Background()
	quizID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	defaultQuiz := &models.QuizWithQuestionsAndImages{
		Quiz: db.Quiz{ID: quizID, TimeLimit: 300, QuestionPoolSize: 0},
		Questions: []models.QuestionWithImages{
			{Question: db.Question{ID: uuid.New(), QuizID: quizID, Text: "What is 2+2?", CorrectAnswer: "4", Points: 10}, Images: nil},
		},
	}

	session := &db.QuizSession{
		ID:           uuid.New(),
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    uuid.New(),
		CurrentIndex: 0,
		Answers:      nil,
		CreatedAt:    now,
	}

	t.Run("success correct answer", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*session, nil)
		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, defaultQuiz)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil)
		mockAttempts.EXPECT().GetAttemptByID(ctx, session.AttemptID).Return(db.QuizAttempt{ID: session.AttemptID, StartedAt: now}, nil)

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		pageData, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err != nil {
			t.Fatalf("NavigateQuestion() error = %v, want nil", err)
		}
		if pageData == nil {
			t.Fatal("NavigateQuestion() returned nil pageData")
		}
	})

	t.Run("success wrong answer", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*session, nil)
		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, defaultQuiz)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil)
		mockAttempts.EXPECT().GetAttemptByID(ctx, session.AttemptID).Return(db.QuizAttempt{ID: session.AttemptID, StartedAt: now}, nil)

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "5", testUser)
		if err != nil {
			t.Fatalf("NavigateQuestion() error = %v, want nil", err)
		}
	})

	t.Run("session not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(db.QuizSession{}, errors.New("session not found"))

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err == nil {
			t.Fatal("NavigateQuestion() error = nil, want error")
		}
	})

	t.Run("unmarshal answers error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		invalidSession := &db.QuizSession{
			ID:           session.ID,
			UserID:       userID,
			QuizID:       quizID,
			AttemptID:    uuid.New(),
			CurrentIndex: 0,
			Answers:      []byte("invalid json"),
			CreatedAt:    now,
		}

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*invalidSession, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{ID: quizID, TimeLimit: 300, QuestionPoolSize: 0}, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{{ID: uuid.New(), QuizID: quizID}}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, gomock.Any()).Return(nil, nil)

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err == nil {
			t.Fatal("NavigateQuestion() error = nil, want error")
		}
	})

	t.Run("get quiz error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("quiz not found"))

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err == nil {
			t.Fatal("NavigateQuestion() error = nil, want error")
		}
	})

	t.Run("update session error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*session, nil)
		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, defaultQuiz)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, errors.New("update error"))

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err == nil {
			t.Fatal("NavigateQuestion() error = nil, want error")
		}
	})

	t.Run("create user answer error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*session, nil)
		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, defaultQuiz)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, errors.New("create answer error"))

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err == nil {
			t.Fatal("NavigateQuestion() error = nil, want error")
		}
	})

	t.Run("last question", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, session.ID).Return(*session, nil)
		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, defaultQuiz)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil)
		mockAttempts.EXPECT().GetAttemptByID(ctx, session.AttemptID).Return(db.QuizAttempt{ID: session.AttemptID, StartedAt: now}, nil)

		testUser := &db.User{ID: userID, Email: "test@example.com"}
		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		pageData, err := svc.NavigateQuestion(ctx, session.ID, quizID, 0, 0, "4", testUser)
		if err != nil {
			t.Fatalf("NavigateQuestion() error = %v, want nil", err)
		}
		if !pageData.IsLastQuestion {
			t.Error("NavigateQuestion() IsLastQuestion = false, want true")
		}
	})
}

func TestQuizSessionService_CompleteSession(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	quizID := uuid.New()
	userID := uuid.New()
	attemptID := uuid.New()
	now := time.Now()

	questionID := uuid.New()

	session := &db.QuizSession{
		ID:           sessionID,
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attemptID,
		CurrentIndex: 1,
		Answers:      nil,
		CreatedAt:    now,
	}

	quiz := &models.QuizWithQuestionsAndImages{
		Quiz: db.Quiz{
			ID:        quizID,
			Title:     "Test Quiz",
			CreatedAt: now,
		},
		Questions: []models.QuestionWithImages{
			{
				Question: db.Question{
					ID:     questionID,
					QuizID: quizID,
					Points: 10,
				},
				Images: nil,
			},
		},
	}

	answers := []db.UserAnswer{
		{ID: uuid.New(), AttemptID: attemptID, QuestionID: questionID, UserAnswer: "4", IsCorrect: true},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, quiz)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: attemptID, Score: 10, MaxScore: 10}, nil)
		mockSessions.EXPECT().DeleteSession(ctx, sessionID).Return(nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		attempt, err := svc.CompleteSession(ctx, sessionID)
		if err != nil {
			t.Fatalf("CompleteSession() error = %v, want nil", err)
		}
		if attempt == nil {
			t.Fatal("CompleteSession() returned nil attempt")
		}
	})

	t.Run("get session error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(db.QuizSession{}, errors.New("session not found"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CompleteSession(ctx, sessionID)
		if err == nil {
			t.Fatal("CompleteSession() error = nil, want error")
		}
	})

	t.Run("get quiz error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("quiz not found"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CompleteSession(ctx, sessionID)
		if err == nil {
			t.Fatal("CompleteSession() error = nil, want error")
		}
	})

	t.Run("get answers error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, quiz)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(nil, errors.New("answers error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CompleteSession(ctx, sessionID)
		if err == nil {
			t.Fatal("CompleteSession() error = nil, want error")
		}
	})

	t.Run("update attempt error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, quiz)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("update error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CompleteSession(ctx, sessionID)
		if err == nil {
			t.Fatal("CompleteSession() error = nil, want error")
		}
	})

	t.Run("delete session error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, quiz)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: attemptID, Score: 10, MaxScore: 10}, nil)
		mockSessions.EXPECT().DeleteSession(ctx, sessionID).Return(errors.New("delete error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.CompleteSession(ctx, sessionID)
		if err == nil {
			t.Fatal("CompleteSession() error = nil, want error")
		}
	})
}

func TestQuizSessionService_GetQuizResultData(t *testing.T) {
	ctx := context.Background()
	quizID := uuid.New()
	sessionID := uuid.New()
	userID := uuid.New()
	attemptID := uuid.New()
	now := time.Now()

	questionID := uuid.New()

	quiz := &models.QuizWithQuestionsAndImages{
		Quiz: db.Quiz{
			ID:        quizID,
			Title:     "Test Quiz",
			CreatedAt: now,
		},
		Questions: []models.QuestionWithImages{
			{
				Question: db.Question{
					ID:            questionID,
					QuizID:        quizID,
					Text:          "What is 2+2?",
					CorrectAnswer: "4",
					Explanation:   "Basic math",
					Points:        10,
				},
				Images: nil,
			},
		},
	}

	session := &db.QuizSession{
		ID:           sessionID,
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attemptID,
		CurrentIndex: 1,
		Answers:      nil,
		CreatedAt:    now,
	}

	answers := []db.UserAnswer{
		{ID: uuid.New(), AttemptID: attemptID, QuestionID: questionID, UserAnswer: "4", IsCorrect: true},
	}

	statsRow := db.GetUserStatsRow{
		TotalXp:    int64(500),
		CorrectCnt: 45,
		WrongCnt:   10,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil).MaxTimes(3)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil).MaxTimes(2)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil).MaxTimes(2)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil).MaxTimes(2)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: attemptID, Score: 10, MaxScore: 10}, nil)
		mockSessions.EXPECT().DeleteSession(ctx, sessionID).Return(nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		result, err := svc.GetQuizResultData(ctx, quizID, sessionID, userID)
		if err != nil {
			t.Fatalf("GetQuizResultData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetQuizResultData() returned nil")
		}
	})

	t.Run("get user error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, errors.New("user error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.GetQuizResultData(ctx, quizID, sessionID, userID)
		if err == nil {
			t.Fatal("GetQuizResultData() error = nil, want error")
		}
	})

	t.Run("get user stats error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.GetQuizResultData(ctx, quizID, sessionID, userID)
		if err == nil {
			t.Fatal("GetQuizResultData() error = nil, want error")
		}
	})
}

func TestQuizSessionService_GetErrorsPageData(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	quizID := uuid.New()
	attemptID := uuid.New()
	questionID := uuid.New()

	attempts := []db.QuizAttempt{
		{ID: attemptID, UserID: userID, QuizID: quizID, Score: 5, MaxScore: 10, StartedAt: now, CompletedAt: sql.NullTime{Time: now, Valid: true}},
	}

	quiz := &models.QuizWithQuestionsAndImages{
		Quiz: db.Quiz{
			ID:        quizID,
			Title:     "Test Quiz",
			CreatedAt: now,
		},
		Questions: []models.QuestionWithImages{
			{
				Question: db.Question{
					ID:            questionID,
					QuizID:        quizID,
					Text:          "What is 2+2?",
					CorrectAnswer: "4",
					Explanation:   "Basic math",
					Points:        10,
				},
				Images: nil,
			},
		},
	}

	answers := []db.UserAnswer{
		{ID: uuid.New(), AttemptID: attemptID, QuestionID: questionID, UserAnswer: "5", IsCorrect: false},
	}

	statsRow := db.GetUserStatsRow{
		TotalXp:    int64(500),
		CorrectCnt: 45,
		WrongCnt:   10,
	}

	t.Run("success with errors", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(attempts, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)
		expectGetQuizWithQuestions(ctx, mockQuizzes, mockQuestions, mockImages, quiz)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		result, err := svc.GetErrorsPageData(ctx, userID)
		if err != nil {
			t.Fatalf("GetErrorsPageData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetErrorsPageData() returned nil")
		}
		if len(result.QuizErrors) != 1 {
			t.Errorf("GetErrorsPageData() QuizErrors len = %d, want 1", len(result.QuizErrors))
		}
	})

	t.Run("success no errors", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		result, err := svc.GetErrorsPageData(ctx, userID)
		if err != nil {
			t.Fatalf("GetErrorsPageData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetErrorsPageData() returned nil")
		}
		if len(result.QuizErrors) != 0 {
			t.Errorf("GetErrorsPageData() QuizErrors len = %d, want 0", len(result.QuizErrors))
		}
	})

	t.Run("get quiz errors error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(nil, errors.New("quiz errors error"))
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.GetErrorsPageData(ctx, userID)
		if err == nil {
			t.Fatal("GetErrorsPageData() error = nil, want error")
		}
	})

	t.Run("get user stats error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(attempts, nil)
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.GetErrorsPageData(ctx, userID)
		if err == nil {
			t.Fatal("GetErrorsPageData() error = nil, want error")
		}
	})
}

func TestQuizSessionService_GetLeaderboardData(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return([]db.GetRecentAttemptsRow{
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 100},
		}, nil)
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{ID: userID, Name: "Test User", Email: "test@example.com"}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		result, err := svc.GetLeaderboardData(ctx, userID)
		if err != nil {
			t.Fatalf("GetLeaderboardData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetLeaderboardData() returned nil")
		}
	})

	t.Run("get leaderboard error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(nil, errors.New("leaderboard error"))
		mockUsers.EXPECT().GetUserByID(ctx, userID).Return(db.User{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.GetLeaderboardData(ctx, userID)
		if err == nil {
			t.Fatal("GetLeaderboardData() error = nil, want error")
		}
	})
}

func TestQuizSessionService_GetUserIDFromRequest(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})
		auth := newMockAuthenticator(ctrl)

		userID := uuid.New()
		auth.EXPECT().ValidateToken("valid-token").Return(&models.AuthClaims{UserID: userID}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})

		result, err := svc.GetUserIDFromRequest(req, auth)
		if err != nil {
			t.Fatalf("GetUserIDFromRequest() error = %v, want nil", err)
		}
		if result != userID {
			t.Errorf("GetUserIDFromRequest() = %v, want %v", result, userID)
		}
	})

	t.Run("missing cookie", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})
		auth := newMockAuthenticator(ctrl)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, err := svc.GetUserIDFromRequest(req, auth)
		if err == nil {
			t.Fatal("GetUserIDFromRequest() error = nil, want error")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})
		auth := newMockAuthenticator(ctrl)

		auth.EXPECT().ValidateToken("invalid-token").Return(nil, errors.New("invalid token"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})

		_, err := svc.GetUserIDFromRequest(req, auth)
		if err == nil {
			t.Fatal("GetUserIDFromRequest() error = nil, want error")
		}
	})
}

func TestQuizSessionService_GetUserStats(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	statsRow := db.GetUserStatsRow{
		TotalXp:    int64(500),
		CorrectCnt: 45,
		WrongCnt:   10,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats == nil {
			t.Fatal("GetUserStats() returned nil")
		}
		if stats.XP != 500 {
			t.Errorf("GetUserStats() XP = %d, want 500", stats.XP)
		}
	})

	t.Run("get stats error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, mockUsers, gamification, nil)
		_, err := svc.GetUserStats(ctx, userID)
		if err == nil {
			t.Fatal("GetUserStats() error = nil, want error")
		}
	})
}
