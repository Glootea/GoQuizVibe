package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/models"
)

type DashboardData struct {
	User        *models.User
	Quizzes     []*models.Quiz
	Stats       *models.UserStats
	Leaderboard []*models.LeaderboardEntry
}

type QuizPageData struct {
	User      *models.User
	Quiz      *models.Quiz
	Stats     *models.UserStats
	SessionID string
}

type QuizResultData struct {
	User         *models.User
	Quiz         *models.Quiz
	Stats        *models.UserStats
	Score        int
	MaxScore     int
	CorrectCount int
	WrongCount   int
	Answers      []AnswerDetail
}

type AnswerDetail struct {
	Question      string
	UserAnswer    string
	CorrectAnswer string
	IsCorrect     bool
}

type ErrorsPageData struct {
	User       *models.User
	QuizErrors []QuizErrors
	Stats      *models.UserStats
}

type QuizErrors struct {
	Quiz         *models.Quiz
	WrongAnswers []models.WrongAnswer
}

type LeaderboardPageData struct {
	User    *models.User
	Entries []*models.LeaderboardEntry
}

type ErrorData struct {
	Message    string
	RedirectTo string
}

type LoginError struct {
	Message string
}

type RegisterError struct {
	Message string
}
