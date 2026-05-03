package models

import (
	"encoding/json"

	"github.com/goquizvibe/db"
)

type QuizWithQuestionsAndImages struct {
	db.Quiz
	Questions []QuestionWithImages `json:"questions"`
}

type QuestionWithImages struct {
	db.Question
	Images []db.QuestionImage `json:"images"`
}

func (q *QuestionWithImages) GetOptions() []string {
	if q.Options == nil {
		return nil
	}
	var opts []string
	if err := json.Unmarshal(q.Options, &opts); err != nil {
		return nil
	}
	return opts
}

type User = db.User
type Quiz = QuizWithQuestionsAndImages
type Question = QuestionWithImages
type QuestionImage = db.QuestionImage
type QuizAttempt = db.QuizAttempt
type UserAnswer = db.UserAnswer
type QuizSession = db.QuizSession
type Role = db.Role

const (
	RoleTeacher Role = db.RoleTeacher
	RoleStudent Role = db.RoleStudent
)

type QuizStatus = db.QuizStatus
type QuestionType = db.QuestionType

const (
	QuizStatusAvailable  QuizStatus = db.QuizStatusAvailable
	QuizStatusAssigned   QuizStatus = db.QuizStatusAssigned
	QuizStatusCompleted  QuizStatus = db.QuizStatusCompleted
	QuizStatusArchived   QuizStatus = db.QuizStatusArchived
)

func GetQuestionOptions(q Question) []string {
	if q.Options == nil {
		return nil
	}
	var opts []string
	if err := json.Unmarshal(q.Options, &opts); err != nil {
		return nil
	}
	return opts
}

func GetQuestionImages(q Question) []db.QuestionImage {
	return q.Images
}