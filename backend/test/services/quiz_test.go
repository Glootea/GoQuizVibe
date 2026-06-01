package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	quizSvc "github.com/goquizvibe/backend/feature/quiz/services"
	"github.com/goquizvibe/backend/shared/db"
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
			result := quizSvc.NormalizeAnswer(tt.input)
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
		m := NewQuizServiceMocks(t)

		quiz1 := db.Quiz{ID: uuid.New(), Title: "Quiz 1"}
		quiz2 := db.Quiz{ID: uuid.New(), Title: "Quiz 2"}
		quizzes := []db.Quiz{quiz1, quiz2}

		m.Groups.EXPECT().GetUserGroupsForStudent(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Quizzes.EXPECT().GetQuizzesForStudent(ctx, gomock.Any()).Return(quizzes, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quiz1.ID).Return([]db.Question{}, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quiz2.ID).Return([]db.Question{}, nil)
		m.Images.EXPECT().GetImagesByQuestionID(ctx, gomock.Any()).Return([]db.QuestionImage{}, nil).AnyTimes()

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
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
		m := NewQuizServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsForStudent(ctx, userID).Return(nil, errors.New("groups error"))

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
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
		m := NewQuizServiceMocks(t)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Question 1", Points: 10},
		}

		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		m.Images.EXPECT().GetImagesByQuestionID(ctx, questions[0].ID).Return([]db.QuestionImage{}, nil)

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
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
		m := NewQuizServiceMocks(t)

		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("sql: no rows"))

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
		_, err := svc.GetQuizByID(ctx, quizID)
		if err == nil {
			t.Fatal("GetQuizByID() error = nil, want error")
		}
	})

	t.Run("questions retrieval error", func(t *testing.T) {
		t.Parallel()
		m := NewQuizServiceMocks(t)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(nil, errors.New("db error"))

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
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
		m := NewQuizServiceMocks(t)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
			{ID: uuid.New(), QuizID: quizID, Text: "Q2", CorrectAnswer: "b", Points: 20},
		}

		answers := map[uuid.UUID]string{
			questions[0].ID: "a",
			questions[1].ID: "b",
		}

		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		m.Attempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		m.Attempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{
			ID:       uuid.New(),
			UserID:   userID,
			QuizID:   quizID,
			Score:    30,
			MaxScore: 30,
		}, nil)
		m.Attempts.EXPECT().UpsertUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil).Times(2)

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
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
		m := NewQuizServiceMocks(t)

		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(db.Quiz{}, errors.New("sql: no rows"))

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, map[uuid.UUID]string{})
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})

	t.Run("questions retrieval error", func(t *testing.T) {
		t.Parallel()
		m := NewQuizServiceMocks(t)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(nil, errors.New("db error"))

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, map[uuid.UUID]string{})
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})

	t.Run("create attempt error", func(t *testing.T) {
		t.Parallel()
		m := NewQuizServiceMocks(t)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		questions := []db.Question{
			{ID: uuid.New(), QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
		}

		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		m.Attempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, errors.New("create failed"))

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
		_, err := svc.SubmitQuizAttempt(ctx, userID, quizID, map[uuid.UUID]string{})
		if err == nil {
			t.Fatal("SubmitQuizAttempt() error = nil, want error")
		}
	})

	t.Run("all correct answers", func(t *testing.T) {
		t.Parallel()
		m := NewQuizServiceMocks(t)

		quiz := db.Quiz{ID: quizID, Title: "Test Quiz"}
		q1ID := uuid.New()
		q2ID := uuid.New()
		questions := []db.Question{
			{ID: q1ID, QuizID: quizID, Text: "Q1", CorrectAnswer: "a", Points: 10},
			{ID: q2ID, QuizID: quizID, Text: "Q2", CorrectAnswer: "b", Points: 10},
		}

		answers := map[uuid.UUID]string{
			q1ID: "a",
			q2ID: "b",
		}

		m.Quizzes.EXPECT().GetQuizByID(ctx, quizID).Return(quiz, nil)
		m.Questions.EXPECT().GetQuestionsByQuizID(ctx, quizID).Return(questions, nil)
		m.Attempts.EXPECT().CreateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{}, nil)
		m.Attempts.EXPECT().UpdateAttempt(ctx, gomock.Any()).Return(db.QuizAttempt{
			ID:       uuid.New(),
			UserID:   userID,
			QuizID:   quizID,
			Score:    20,
			MaxScore: 20,
		}, nil)
		m.Attempts.EXPECT().UpsertUserAnswer(ctx, gomock.Any()).Return(db.UserAnswer{}, nil).Times(2)

		svc := quizSvc.NewQuizService(m.Quizzes, m.Questions, m.Images, m.Attempts, m.Groups, nil)
		result, err := svc.SubmitQuizAttempt(ctx, userID, quizID, answers)
		if err != nil {
			t.Fatalf("SubmitQuizAttempt() error = %v, want nil", err)
		}
		if result.Score != 20 {
			t.Errorf("result.Score = %d, want 20", result.Score)
		}
	})
}
