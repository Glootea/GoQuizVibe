package models

import (
	"encoding/json"

	"github.com/goquizvibe/db"
)

type QuizWithQuestions struct {
	db.Quiz
	Questions []db.Question
}

func (q *QuizWithQuestions) GetOptions(index int) []string {
	if index < 0 || index >= len(q.Questions) {
		return nil
	}
	if q.Questions[index].Options == nil {
		return nil
	}
	var opts []string
	if err := json.Unmarshal(q.Questions[index].Options, &opts); err != nil {
		return nil
	}
	return opts
}

type User = db.User
type Quiz = QuizWithQuestions
type Question = db.Question
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