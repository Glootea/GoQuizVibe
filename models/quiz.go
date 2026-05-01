package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QuizStatus string

const (
	QuizStatusAssigned  QuizStatus = "assigned"
	QuizStatusCompleted QuizStatus = "completed"
	QuizStatusAvailable QuizStatus = "available"
)

type Quiz struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	Subject     string     `json:"subject"`
	Grade       int        `json:"grade"`
	Status      QuizStatus `gorm:"type:varchar(20);not null" json:"status"`
	TimeLimit   int        `json:"time_limit"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	Questions   []Question `gorm:"foreignKey:QuizID" json:"questions,omitempty"`
}

func (Quiz) TableName() string {
	return "quizzes"
}

type QuestionType string

const (
	QuestionTypeChoice QuestionType = "choice"
	QuestionTypeOpen   QuestionType = "open"
	QuestionTypeFill   QuestionType = "fill"
)

type Question struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	QuizID        uuid.UUID    `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Text          string       `gorm:"not null" json:"text"`
	Type          QuestionType `gorm:"type:varchar(20);not null" json:"type"`
	Options       []byte       `gorm:"type:jsonb" json:"options"`
	CorrectAnswer string       `gorm:"not null" json:"-"`
	Explanation   string       `json:"explanation"`
	Points        int          `gorm:"default:10" json:"points"`
	OrderIndex    int          `gorm:"default:0" json:"order_index"`
}

func (Question) TableName() string {
	return "questions"
}

func (q *Question) GetOptions() []string {
	if q.Options == nil {
		return []string{}
	}
	var opts []string
	if err := json.Unmarshal(q.Options, &opts); err != nil {
		return []string{}
	}
	return opts
}