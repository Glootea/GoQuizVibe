package types

import (
	"time"

	"github.com/goquizvibe/models"
	"github.com/google/uuid"
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
	ID          uuid.UUID     `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID     `gorm:"column:user_id;type:uuid;not null;index" json:"user_id"`
	QuizID      uuid.UUID     `gorm:"column:quiz_id;type:uuid;not null;index" json:"quiz_id"`
	Score       int           `gorm:"column:score;default:0" json:"score"`
	MaxScore    int           `gorm:"column:max_score;default:0" json:"max_score"`
	StartedAt   time.Time     `gorm:"column:started_at;autoCreateTime" json:"started_at"`
	CompletedAt *time.Time    `gorm:"column:completed_at" json:"completed_at,omitempty"`
	UserName    string        `gorm:"column:user_name" json:"user_name"`
	QuizTitle   string        `gorm:"column:quiz_title" json:"quiz_title"`
}

func (AttemptWithUser) TableName() string {
	return "quiz_attempts"
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