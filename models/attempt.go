package models

import (
	"time"

	"github.com/google/uuid"
)

type WrongAnswer struct {
	ID            uuid.UUID `json:"id"`
	QuestionID    uuid.UUID `json:"question_id"`
	QuizID        uuid.UUID `json:"quiz_id"`
	UserAnswer    string    `json:"user_answer"`
	CorrectAnswer string    `json:"correct_answer"`
	Timestamp     time.Time `json:"timestamp"`
}