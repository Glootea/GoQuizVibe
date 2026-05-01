package types

import (
	"github.com/goquizvibe/models"
)

type DashboardData struct {
	User        *models.User
	Quizzes     []*models.Quiz
	Stats       *models.UserProgress
	Leaderboard []*models.LeaderboardEntry
}

type QuizPageData struct {
	User      *models.User
	Quiz      *models.Quiz
	Stats     *models.UserProgress
	SessionID string
}

type QuizResultData struct {
	User         *models.User
	Quiz         *models.Quiz
	Stats        *models.UserProgress
	Score        int
	MaxScore     int
	CorrectCount int
	WrongCount   int
	Answers      []AnswerDetail
}

type AnswerDetail struct {
	Question     string
	UserAnswer   string
	CorrectAnswer string
	IsCorrect    bool
}

type ErrorsPageData struct {
	User         *models.User
	QuizErrors   []QuizErrors
}

type QuizErrors struct {
	Quiz        *models.Quiz
	WrongAnswers []models.WrongAnswer
}

type LeaderboardPageData struct {
	User    *models.User
	Entries []*models.LeaderboardEntry
}

type LoginError struct {
	Message string
}

type RegisterError struct {
	Message string
}
