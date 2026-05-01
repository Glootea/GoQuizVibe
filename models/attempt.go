package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizAttempt struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	QuizID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Score       int          `gorm:"default:0" json:"score"`
	MaxScore    int          `gorm:"default:0" json:"max_score"`
	StartedAt   time.Time    `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Answers     []UserAnswer `gorm:"foreignKey:AttemptID" json:"answers,omitempty"`
}

func (QuizAttempt) TableName() string {
	return "quiz_attempts"
}

type UserAnswer struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AttemptID  uuid.UUID `gorm:"type:uuid;not null;index" json:"attempt_id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index" json:"question_id"`
	UserAnswer string    `gorm:"not null" json:"user_answer"`
	IsCorrect  bool      `gorm:"default:false" json:"is_correct"`
}

func (UserAnswer) TableName() string {
	return "user_answers"
}

type QuizSession struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	QuizID        uuid.UUID `gorm:"type:uuid;not null;index" json:"quiz_id"`
	AttemptID     uuid.UUID `gorm:"type:uuid;not null" json:"attempt_id"`
	CurrentIndex  int       `gorm:"default:0" json:"current_index"`
	Answers       []byte    `gorm:"type:jsonb" json:"answers"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (QuizSession) TableName() string {
	return "quiz_sessions"
}

type WrongAnswer struct {
	ID            uuid.UUID `json:"id"`
	QuestionID    uuid.UUID `json:"question_id"`
	QuizID        uuid.UUID `json:"quiz_id"`
	UserAnswer    string    `json:"user_answer"`
	CorrectAnswer string    `json:"correct_answer"`
	Timestamp     time.Time `json:"timestamp"`
}