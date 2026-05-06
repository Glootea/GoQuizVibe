package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/mocks/services"
	"github.com/goquizvibe/services"
	"go.uber.org/mock/gomock"
)

func TestNormalizeAnswer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim spaces and lowercase",
			input:    "  Hello  ",
			expected: "hello",
		},
		{
			name:     "trim trailing dot",
			input:    "Привет.",
			expected: "привет",
		},
		{
			name:     "only spaces becomes empty",
			input:    "  ",
			expected: "",
		},
		{
			name:     "multiple trailing dots",
			input:    "A.B.C.",
			expected: "a.b.c",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already lowercase",
			input:    "ALREADY_LOWER",
			expected: "already_lower",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := services.NormalizeAnswer(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAnswer(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQuizService_GetQuizzesForUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful retrieval", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz1 := db.Quiz{ID: uuid.New(), Title: "Quiz 1"}
		quiz2 := db.Quiz{ID: uuid.New(), Title: "Quiz 2"}
		quizzes := []db.Quiz{quiz1, quiz2}

		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return(quizzes, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quiz1.ID).Return([]db.Question{}, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quiz2.ID).Return([]db.Question{}, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, gomock.Any()).Return([]db.QuestionImage{}, nil).AnyTimes()

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		result, err := svc.GetQuizzesForUser(ctx, userID)
		if err != nil {
			t.Fatalf("GetQuizzesForUser() error = %v, want nil", err)
		}
		if len(result) != 2 {
			t.Errorf("len(result) = %d, want 2", len(result))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		mockQuizzes.EXPECT().GetQuizzesForUser(ctx, userID).Return(nil, errors.New("db error"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.GetQuizzesForUser(ctx, userID)
		if err == nil {
			t.Fatal("GetQuizzesForUser() error = nil, want error")
		}
	})
}

func TestQuizService_GetQuizByID(t *testing.T) {
	ctx := context.Background()
	quizID := uuid.New()

	t.Run("successful retrieval", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Question 1", Points: 10},
		}

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		mockImages.EXPECT().GetImagesByQuestionID(ctx, questions[0].ID).Return([]db.QuestionImage{}, nil)

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		result, err := svc.GetQuizByID(ctx, quizID)
		if err != nil {
			t.Fatalf("GetQuizByID() error = %v, want nil", err)
		}
		if result.Quiz.Title != "Test Quiz" {
			t.Errorf("result.Quiz.Title = %v, want Test Quiz", result.Quiz.Title)
		}
		if len(result.Questions) != 1 {
			t.Errorf("len(result.Questions) = %d, want 1", len(result.Questions))
		}
	})

	t.Run("quiz not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("sql: no rows"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.GetQuizByID(ctx, quizID)
		if err == nil {
			t.Fatal("GetQuizByID() error = nil, want error")
		}
	})

	t.Run("questions retrieval error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(nil, errors.New("db error"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.GetQuizByID(ctx, quizID)
		if err == nil {
			t.Fatal("GetQuizByID() error = nil, want error")
		}
	})
}

func TestQuizService_SubmitQuizAttempt(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	quizID := uuid.New()

	t.Run("successful submission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
			{ID: uuid.New(), QuizID: quizID, Text: "Q2", CorrectAnswer: "b", Points: 20},
		}
		answers := map[uuid.UUID]string{
			questions[0].ID: "a",
			questions[1].ID: "b",
		}

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{
			ID:       uuid.New(),
			UserID:   userID,
			QuizID:   quizID,
			Score:    30,
			MaxScore: 30,
		}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil).Times(2)

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		result, err := svc.SubmitQuizAttempt(ctx, userID, quizID, answers)
		if err != nil {
			t.Fatalf("SubmitQuizAttempt() error = %v, want nil", err)
		}
		if result.Score != 30 {
			t.Errorf("result.Score = %d, want 30", result.Score)
		}
	})

	t.Run("quiz not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("sql: no rows"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, nil)
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})

	t.Run("create attempt error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
		}

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("create failed"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, map[uuid.UUID]string{})
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})

	t.Run("update attempt error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
		}

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("update failed"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, map[uuid.UUID]string{})
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})

	t.Run("create user answers error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockQuizzes := mocks.NewMockQuizRepository(ctrl)
		mockQuestions := mocks.NewMockQuestionRepository(ctrl)
		mockImages := mocks.NewMockImageRepository(ctrl)
		mockAttempts := mocks.NewMockAttemptRepository(ctrl)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
		}

		mockQuizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		mockQuestions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		mockAttempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		mockAttempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{ID: uuid.New()}, nil)
		mockAttempts.EXPECT().CreateUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, errors.New("create answer failed"))

		svc := services.NewQuizService(mockQuizzes, mockQuestions, mockImages, mockAttempts)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, map[uuid.UUID]string{})
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})
}
