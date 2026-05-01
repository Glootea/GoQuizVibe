package types

import (
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
	Question     string
	UserAnswer   string
	CorrectAnswer string
	IsCorrect    bool
}

type ErrorsPageData struct {
	User       *models.User
	QuizErrors []QuizErrors
	Stats      *models.UserStats
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

type AdminDashboardData struct {
	User           *models.User
	QuizCount      int
	StudentCount   int
	AttemptCount   int
	AvgScore       float64
	RecentActivity []*RecentAttempt
}

type RecentAttempt struct {
	AttemptID   string
	UserName    string
	QuizTitle   string
	Score       int
	MaxScore    int
	CompletedAt string
}

type AdminQuizListData struct {
	User    *models.User
	Quizzes []*QuizWithStats
}

type QuizWithStats struct {
	*models.Quiz
	AttemptCount int
	AvgScore     float64
}

type AdminResultsData struct {
	User    *models.User
	Attempts []*AttemptWithUser
	Quizzes  []*models.Quiz
}

type AttemptWithUser struct {
	*models.QuizAttempt
	UserName  string `gorm:"column:user_name"`
	QuizTitle string `gorm:"column:quiz_title"`
}

type AdminStatisticsData struct {
	User                *models.User
	TotalQuizzes        int
	TotalStudents       int
	TotalAttempts       int
	AvgScore            float64
	QuizStats           []*QuizStatistics
	GradeDistribution   map[int]int
	SubjectDistribution map[string]int
}

type QuizStatistics struct {
	Quiz         *models.Quiz
	AttemptCount int
	AvgScore     float64
	PassRate     float64
}

type AdminQuizEditData struct {
	User    *models.User
	Quiz    *models.Quiz
	Questions []models.Question
}