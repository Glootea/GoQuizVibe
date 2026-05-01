package models

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type QuizStatus string

const (
	QuizStatusAssigned QuizStatus = "assigned"
	QuizStatusCompleted QuizStatus = "completed"
	QuizStatusAvailable QuizStatus = "available"
)

type Quiz struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Subject     string     `json:"subject"`
	Grade       int        `json:"grade"`
	Questions   []Question `json:"questions"`
	AssignedTo  []uuid.UUID `json:"assigned_to"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	Status      QuizStatus `json:"status"`
	TimeLimit   int        `json:"time_limit"`
	CreatedAt   time.Time  `json:"created_at"`
}

type QuestionType string

const (
	QuestionTypeChoice QuestionType = "choice"
	QuestionTypeOpen   QuestionType = "open"
	QuestionTypeFill   QuestionType = "fill"
)

type Question struct {
	ID            uuid.UUID      `json:"id"`
	Text          string         `json:"text"`
	Type          QuestionType   `json:"type"`
	Options       []string       `json:"options"`
	CorrectAnswer string         `json:"-"`
	Explanation   string         `json:"explanation"`
	Points        int            `json:"points"`
}

type WrongAnswer struct {
	ID           uuid.UUID `json:"id"`
	QuestionID   uuid.UUID `json:"question_id"`
	QuizID       uuid.UUID `json:"quiz_id"`
	UserAnswer   string    `json:"user_answer"`
	CorrectAnswer string   `json:"correct_answer"`
	Timestamp    time.Time `json:"timestamp"`
}

type UserProgress struct {
	UserID          uuid.UUID     `json:"user_id"`
	XP              int           `json:"xp"`
	Streak          int           `json:"streak"`
	LastActiveDate  time.Time     `json:"last_active_date"`
	CompletedQuizzes []uuid.UUID  `json:"completed_quizzes"`
	WrongAnswers    []WrongAnswer `json:"wrong_answers"`
}

type QuizAttempt struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	QuizID      uuid.UUID     `json:"quiz_id"`
	Score       int           `json:"score"`
	MaxScore    int           `json:"max_score"`
	Answers     []UserAnswer  `json:"answers"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
}

type UserAnswer struct {
	QuestionID   uuid.UUID `json:"question_id"`
	UserAnswer   string    `json:"user_answer"`
	IsCorrect    bool      `json:"is_correct"`
}

type LeaderboardEntry struct {
	UserID   uuid.UUID `json:"user_id"`
	UserName string    `json:"user_name"`
	XP       int       `json:"xp"`
	Streak   int       `json:"streak"`
	Rank     int       `json:"rank"`
}

type AuthClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   Role      `json:"role"`
}
