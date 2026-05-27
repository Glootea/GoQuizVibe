package models

import (
	"time"

	"github.com/google/uuid"
)

type WrongAnswer struct {
	QuestionID    uuid.UUID `json:"question_id"`
	QuizID        uuid.UUID `json:"quiz_id"`
	UserAnswer    string    `json:"user_answer"`
	CorrectAnswer string    `json:"correct_answer"`
	Explanation   string    `json:"explanation"`
	Timestamp     time.Time `json:"timestamp"`
}