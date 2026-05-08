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
	mockStats := mocks.NewMockStatsRepository(ctrl)

	gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

	svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
	if svc == nil {
		t.Error("NewQuizSessionService() returned nil")
	}
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockSessions.EXPECT().CreateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("attempt error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})

		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockSessions.EXPECT().CreateSession(ctx, gomock.Any()).Return(db.QuizSession{}, errors.New("session error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.CreateSession(ctx, userID, quizID)
		if err == nil {
			t.Fatal("CreateSession() error = nil, want error")
		}
	})
}

func TestQuizSessionService_SubmitAnswer(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	quizID := uuid.New()
	userID := uuid.New()

	questionID := uuid.New()
	now := time.Now()

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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		feedback, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err != nil {
			t.Fatalf("SubmitAnswer() error = %v, want nil", err)
		}
		if feedback == nil {
			t.Fatal("SubmitAnswer() returned nil feedback")
		}
		if !feedback.IsCorrect {
			t.Error("SubmitAnswer() IsCorrect = false, want true")
		}
		if feedback.CorrectAnswer != "4" {
			t.Errorf("SubmitAnswer() CorrectAnswer = %s, want 4", feedback.CorrectAnswer)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		feedback, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "5")
		if err != nil {
			t.Fatalf("SubmitAnswer() error = %v, want nil", err)
		}
		if feedback.IsCorrect {
			t.Error("SubmitAnswer() IsCorrect = true, want false")
		}
	})

	t.Run("question index out of range", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.SubmitAnswer(ctx, sessionID, quizID, 10, "4")
		if err == nil {
			t.Fatal("SubmitAnswer() error = nil, want error")
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(db.QuizSession{}, errors.New("session not found"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err == nil {
			t.Fatal("SubmitAnswer() error = nil, want error")
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		invalidSession := &db.QuizSession{
			ID:           sessionID,
			UserID:       userID,
			QuizID:       quizID,
			AttemptID:    uuid.New(),
			CurrentIndex: 0,
			Answers:      []byte("invalid json"),
			CreatedAt:    now,
		}

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*invalidSession, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err == nil {
			t.Fatal("SubmitAnswer() error = nil, want error")
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("quiz not found"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err == nil {
			t.Fatal("SubmitAnswer() error = nil, want error")
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, errors.New("update error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err == nil {
			t.Fatal("SubmitAnswer() error = nil, want error")
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, errors.New("create answer error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err == nil {
			t.Fatal("SubmitAnswer() error = nil, want error")
		}
	})

	t.Run("is last question", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockSessions.EXPECT().UpdateSession(ctx, gomock.Any()).Return(db.QuizSession{}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		feedback, err := svc.SubmitAnswer(ctx, sessionID, quizID, 0, "4")
		if err != nil {
			t.Fatalf("SubmitAnswer() error = %v, want nil", err)
		}
		if !feedback.IsLast {
			t.Error("SubmitAnswer() IsLast = false, want true")
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: attemptID, Score: 10, MaxScore: 10}, nil)
		mockSessions.EXPECT().DeleteSession(ctx, sessionID).Return(nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(db.QuizSession{}, errors.New("session not found"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("quiz not found"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(nil, errors.New("answers error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("update error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: attemptID, Score: 10, MaxScore: 10}, nil)
		mockSessions.EXPECT().DeleteSession(ctx, sessionID).Return(errors.New("delete error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockSessions.EXPECT().GetSession(ctx, sessionID).Return(*session, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: attemptID, Score: 10, MaxScore: 10}, nil)
		mockSessions.EXPECT().DeleteSession(ctx, sessionID).Return(nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		result, err := svc.GetQuizResultData(ctx, quizID, sessionID, userID)
		if err != nil {
			t.Fatalf("GetQuizResultData() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("GetQuizResultData() returned nil")
		}
	})

	t.Run("get attempts error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(nil, errors.New("attempts error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.GetQuizResultData(ctx, quizID, sessionID, userID)
		if err == nil {
			t.Fatal("GetQuizResultData() error = nil, want error")
		}
	})

	t.Run("get quiz with questions error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(nil, errors.New("questions error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(attempts, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz.Quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return([]db.Question{quiz.Questions[0].Question}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, quiz.Questions[0].ID).Return(nil, nil)
		mockAttempts.EXPECT().GetAnswersByAttempt(ctx, attemptID).Return(answers, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(nil, errors.New("quiz errors error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(attempts, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.GetErrorsPageData(ctx, userID)
		if err == nil {
			t.Fatal("GetErrorsPageData() error = nil, want error")
		}
	})

	t.Run("continue on get quiz error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAttempts := mocks.NewMockAttemptRepository(ctrl)
		mockSessions := mocks.NewMockSessionRepository(ctrl)
		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetQuizErrors(ctx, userID).Return(attempts, nil)
		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("quiz error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		result, err := svc.GetErrorsPageData(ctx, userID)
		if err != nil {
			t.Fatalf("GetErrorsPageData() error = %v, want nil", err)
		}
		if len(result.QuizErrors) != 0 {
			t.Errorf("GetErrorsPageData() QuizErrors len = %d, want 0 (skipped due to quiz error)", len(result.QuizErrors))
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return([]db.GetRecentAttemptsRow{
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 100},
		}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockAttempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(nil, errors.New("leaderboard error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})
		auth := newMockAuthenticator(ctrl)

		userID := uuid.New()
		auth.EXPECT().ValidateToken("valid-token").Return(&models.AuthClaims{UserID: userID}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)

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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})
		auth := newMockAuthenticator(ctrl)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)

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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: time.Now()})
		auth := newMockAuthenticator(ctrl)

		auth.EXPECT().ValidateToken("invalid-token").Return(nil, errors.New("invalid token"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)

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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockStats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		mockStats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)
		mockAttempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
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
		mockStats := mocks.NewMockStatsRepository(ctrl)

		gamification := services.NewGamificationService(mockAttempts, mockStats, &mockTimeProvider{now: now})

		mockStats.EXPECT().GetUserStats(ctx, userID).Return(db.GetUserStatsRow{}, errors.New("stats error"))

		svc := services.NewQuizSessionService(mockAttempts, mockSessions, mockQuizzes, mockQuestions, mockImages, gamification)
		_, err := svc.GetUserStats(ctx, userID)
		if err == nil {
			t.Fatal("GetUserStats() error = nil, want error")
		}
	})
}
