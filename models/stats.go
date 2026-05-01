package models

type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	XP       int    `json:"xp"`
	Streak   int    `json:"streak"`
	Rank     int    `json:"rank"`
}

type UserStats struct {
	UserID          string   `json:"user_id"`
	XP              int      `json:"xp"`
	Streak          int      `json:"streak"`
	LastActiveDate  string   `json:"last_active_date"`
	CompletedQuizzes []string `json:"completed_quizzes"`
	CorrectCount    int      `json:"correct_count"`
	WrongCount      int      `json:"wrong_count"`
}