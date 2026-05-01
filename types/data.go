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
	User  *models.User
	Quiz  *models.Quiz
	Stats *models.UserProgress
}

type ErrorsPageData struct {
	User         *models.User
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
