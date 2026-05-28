package types

import (
	"encoding/json"

	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/models"

	"github.com/google/uuid"
)

type DashboardData struct {
	User          *models.User
	Quizzes       []*models.Quiz
	Stats         *models.UserStats
	Leaderboard   []*models.LeaderboardEntry
	ActiveSession *ActiveSessionInfo
}

type ActiveSessionInfo struct {
	SessionID        uuid.UUID
	QuizID           uuid.UUID
	QuizTitle        string
	CurrentIndex     int
	RemainingSeconds int
}

type SessionConflictData struct {
	ExistingSessionID uuid.UUID
	ExistingQuizID    uuid.UUID
	ExistingQuizTitle string
	CurrentIndex      int
	RequestedQuizID   uuid.UUID
}

type AnswerState struct {
	Text     string
	Answered bool
}

type QuizPageData struct {
	User             *models.User
	Questions        []models.QuestionWithImages
	QuestionOrder    []int
	AnswerOptions    map[int][]string
	SessionID        string
	CurrentIndex     int
	Answers          map[int]AnswerState
	TotalQuestions   int
	RemainingSeconds int
	TimeLimitMinutes int
	IsLastQuestion   bool
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
	QuestionType  string
	UserAnswer    string
	CorrectAnswer string
	IsCorrect     bool
	Explanation   string
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

type QuizInfoData struct {
	User          *models.User
	Quiz          *models.Quiz
	QuestionCount int
	TimeLimitMin  int
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

type AdminDashboardData struct {
	User            *models.User
	QuizCount       int
	StudentCount    int
	AttemptCount    int
	AvgScore        float64
	MaterialCount   int
	RecentActivity  []*RecentAttempt
	RecentMaterials []MaterialWithURL
}

type RecentAttempt struct {
	AttemptID   string
	UserName    string
	QuizTitle   string
	Score       int
	MaxScore    int
	CompletedAt string
}

type MaterialWithURL struct {
	Material   db.LearningMaterial
	PublicURL  string
	Type       string
	Permission db.PermissionType
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
	User     *models.User
	Attempts []*AttemptWithUser
	Quizzes  []*models.Quiz
}

type AttemptWithUser struct {
	*models.QuizAttempt
	UserName  string `json:"user_name"`
	QuizTitle string `json:"quiz_title"`
}

type AdminStatisticsData struct {
	User                *models.User
	TotalQuizzes        int
	TotalStudents       int
	TotalAttempts       int
	AvgScore            float64
	QuizStats           []*QuizStatistics
	GradeDistribution   map[string]int
	SubjectDistribution map[string]int
}

type QuizStatistics struct {
	QuizID       uuid.UUID `json:"quiz_id"`
	Title        string    `json:"title"`
	Subject      string    `json:"subject"`
	AttemptCount int       `json:"attempt_count"`
	AvgScore     float64   `json:"avg_score"`
	PassRate     float64   `json:"pass_rate"`
}

type AdminQuizEditData struct {
	User      *models.User
	Quiz      *models.Quiz
	Questions []models.Question
}

type QuizStatsResponse struct {
	QuizID       uuid.UUID `json:"quiz_id"`
	Title        string    `json:"title"`
	Subject      string    `json:"subject"`
	AttemptCount int       `json:"attempt_count"`
	AvgScore     float64   `json:"avg_score"`
	PassRate     float64   `json:"pass_rate"`
}

type GradeDistResponse struct {
	Distribution map[string]int `json:"distribution"`
}

type SubjectDistResponse struct {
	Distribution map[string]int `json:"distribution"`
}

type HTMXQuizStatsRow struct {
	QuizID       uuid.UUID `json:"quiz_id"`
	Title        string    `json:"title"`
	Subject      string    `json:"subject"`
	AttemptCount int64     `json:"attempt_count"`
	AvgScore     float64   `json:"avg_score"`
	PassRate     float64   `json:"pass_rate"`
}

type HTMXGradeDistRow struct {
	GradeDist json.RawMessage `json:"grade_dist"`
}

type HTMXSubjectDistRow struct {
	SubjectDist json.RawMessage `json:"subject_dist"`
}
