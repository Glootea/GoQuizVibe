package store

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/types"
	"gorm.io/gorm"
)

var (
	ErrEmailExists  = errors.New("email already registered")
	ErrUserNotFound = errors.New("user not found")
	ErrQuizNotFound = errors.New("quiz not found")
	ErrNotFound     = errors.New("record not found")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(u *models.User) error {
	var existing User
	err := r.db.Where("email = ?", u.Email).First(&existing).Error
	if err == nil {
		return ErrEmailExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(u).Error
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetQuizzes() ([]*models.Quiz, error) {
	var quizzes []*models.Quiz
	err := r.db.Find(&quizzes).Error
	return quizzes, err
}

func (r *Repository) GetQuizByID(id uuid.UUID) (*models.Quiz, error) {
	var quiz models.Quiz
	err := r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC")
	}).First(&quiz, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrQuizNotFound
	}
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

func (r *Repository) GetQuizzesForUser(userID uuid.UUID) ([]*models.Quiz, error) {
	var quizzes []*models.Quiz
	err := r.db.Where("status = ?", models.QuizStatusAvailable).
		Or("created_by = ?", userID).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Find(&quizzes).Error
	return quizzes, err
}

func (r *Repository) SaveAttempt(attempt *models.QuizAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *Repository) GetAttemptsByUser(userID uuid.UUID) ([]*models.QuizAttempt, error) {
	var attempts []*models.QuizAttempt
	err := r.db.Where("user_id = ?", userID).Find(&attempts).Error
	return attempts, err
}

func (r *Repository) GetAttemptByID(id uuid.UUID) (*models.QuizAttempt, error) {
	var attempt models.QuizAttempt
	err := r.db.Preload("Answers").First(&attempt, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *Repository) UpdateAttempt(attempt *models.QuizAttempt) error {
	return r.db.Save(attempt).Error
}

func (r *Repository) GetUserStats(userID uuid.UUID) (*models.UserStats, error) {
	var result struct {
		TotalXP    int
		CorrectCnt int64
		WrongCnt   int64
	}

	r.db.Model(&models.UserAnswer{}).
		Select("COALESCE(SUM(q.points), 0) as total_xp, COUNT(CASE WHEN ua.is_correct = true THEN 1 END) as correct_cnt, COUNT(CASE WHEN ua.is_correct = false THEN 1 END) as wrong_cnt").
		Joins("JOIN questions q ON q.id = user_answers.question_id").
		Joins("JOIN quiz_attempts a ON a.id = user_answers.attempt_id").
		Where("a.user_id = ?", userID).
		Scan(&result)

	var lastActive time.Time
	r.db.Model(&models.QuizAttempt{}).
		Select("MAX(completed_at)").
		Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Scan(&lastActive)

	var completedCount int64
	r.db.Model(&models.QuizAttempt{}).
		Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Count(&completedCount)

	streak := r.calculateStreak(userID)

	return &models.UserStats{
		UserID:           userID.String(),
		XP:               result.TotalXP,
		Streak:           streak,
		LastActiveDate:   lastActive.Format("2006-01-02"),
		CompletedQuizzes: nil,
		CorrectCount:     int(result.CorrectCnt),
		WrongCount:       int(result.WrongCnt),
	}, nil
}

func (r *Repository) calculateStreak(userID uuid.UUID) int {
	var attempts []models.QuizAttempt
	r.db.Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Order("completed_at DESC").
		Find(&attempts)

	if len(attempts) == 0 {
		return 0
	}

	streak := 1
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for i := 0; i < len(attempts)-1; i++ {
		if attempts[i].CompletedAt == nil {
			continue
		}
		current := time.Date(attempts[i].CompletedAt.Year(), attempts[i].CompletedAt.Month(), attempts[i].CompletedAt.Day(), 0, 0, 0, 0, now.Location())

		var nextDay time.Time
		if i+1 < len(attempts) && attempts[i+1].CompletedAt != nil {
			nextDay = time.Date(attempts[i+1].CompletedAt.Year(), attempts[i+1].CompletedAt.Month(), attempts[i+1].CompletedAt.Day(), 0, 0, 0, 0, now.Location())
		} else {
			nextDay = current.AddDate(0, 0, -1)
		}

		diff := current.Sub(nextDay).Hours()
		if diff >= 23 && diff <= 25 {
			streak++
		} else {
			break
		}
	}

	if len(attempts) > 0 && attempts[0].CompletedAt != nil {
		lastActive := time.Date(attempts[0].CompletedAt.Year(), attempts[0].CompletedAt.Month(), attempts[0].CompletedAt.Day(), 0, 0, 0, 0, now.Location())
		if lastActive.Before(today.AddDate(0, 0, -1)) {
			return 0
		}
	}

	return streak
}

func (r *Repository) GetLeaderboard(limit int) ([]*models.LeaderboardEntry, error) {
	var results []struct {
		UserID   uuid.UUID
		UserName string
		TotalXP  int
	}

	r.db.Model(&models.UserAnswer{}).
		Select("u.id as user_id, u.name as user_name, COALESCE(SUM(q.points), 0) as total_xp").
		Joins("JOIN quiz_attempts a ON a.id = user_answers.attempt_id").
		Joins("JOIN users u ON u.id = a.user_id").
		Joins("JOIN questions q ON q.id = user_answers.question_id").
		Where("user_answers.is_correct = true").
		Group("u.id, u.name").
		Order("total_xp DESC").
		Limit(limit).
		Scan(&results)

	entries := make([]*models.LeaderboardEntry, len(results))
	for i, row := range results {
		entries[i] = &models.LeaderboardEntry{
			UserID:   row.UserID.String(),
			UserName: row.UserName,
			XP:       row.TotalXP,
			Streak:   r.calculateStreak(row.UserID),
			Rank:     i + 1,
		}
	}

	return entries, nil
}

func (r *Repository) CreateSession(session *models.QuizSession) error {
	return r.db.Create(session).Error
}

func (r *Repository) GetSession(sessionID uuid.UUID) (*models.QuizSession, error) {
	var session models.QuizSession
	err := r.db.First(&session, "id = ?", sessionID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) GetSessionByAttemptID(attemptID uuid.UUID) (*models.QuizSession, error) {
	var session models.QuizSession
	err := r.db.First(&session, "attempt_id = ?", attemptID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) UpdateSession(session *models.QuizSession) error {
	return r.db.Save(session).Error
}

func (r *Repository) DeleteSession(sessionID uuid.UUID) error {
	return r.db.Delete(&models.QuizSession{}, "id = ?", sessionID).Error
}

func (r *Repository) SaveUserAnswer(answer *models.UserAnswer) error {
	return r.db.Create(answer).Error
}

func (r *Repository) GetAnswersByAttempt(attemptID uuid.UUID) ([]*models.UserAnswer, error) {
	var answers []*models.UserAnswer
	err := r.db.Where("attempt_id = ?", attemptID).Find(&answers).Error
	return answers, err
}

func (r *Repository) GetQuizWithQuestions(quizID uuid.UUID) (*models.Quiz, error) {
	var quiz models.Quiz
	err := r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC")
	}).First(&quiz, "id = ?", quizID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrQuizNotFound
	}
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

func (r *Repository) GetCompletedAttemptBySessionID(sessionID uuid.UUID) (*models.QuizAttempt, error) {
	var session models.QuizSession
	err := r.db.First(&session, "id = ?", sessionID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var attempt models.QuizAttempt
	err = r.db.Preload("Answers").First(&attempt, "id = ?", session.AttemptID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if attempt.CompletedAt == nil {
		return nil, ErrNotFound
	}

	return &attempt, nil
}

func (r *Repository) GetQuizErrors(userID uuid.UUID) ([]*models.QuizAttempt, error) {
	var attempts []*models.QuizAttempt
	err := r.db.Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Preload("Answers", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_correct = false")
		}).
		Find(&attempts).Error
	return attempts, err
}

type User struct {
	ID       uuid.UUID
	Email    string
	Name     string
	Password string
}

func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (r *Repository) GetQuestionByID(id uuid.UUID) (*models.Question, error) {
	var question models.Question
	err := r.db.First(&question, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *Repository) CreateQuiz(quiz *models.Quiz) error {
	return r.db.Create(quiz).Error
}

func (r *Repository) UpdateQuiz(quiz *models.Quiz) error {
	return r.db.Save(quiz).Error
}

func (r *Repository) UpdateQuizStatus(id uuid.UUID, status models.QuizStatus) error {
	return r.db.Model(&models.Quiz{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) DeleteQuiz(id uuid.UUID) error {
	return r.db.Model(&models.Quiz{}).Where("id = ?", id).Update("status", models.QuizStatusArchived).Error
}

func (r *Repository) GetNonArchivedQuizzes() ([]*models.Quiz, error) {
	var quizzes []*models.Quiz
	err := r.db.Where("status != ?", models.QuizStatusArchived).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).Find(&quizzes).Error
	return quizzes, err
}

func (r *Repository) CreateQuestion(q *models.Question) error {
	return r.db.Create(q).Error
}

func (r *Repository) UpdateQuestion(q *models.Question) error {
	return r.db.Save(q).Error
}

func (r *Repository) DeleteQuestion(id uuid.UUID) error {
	return r.db.Delete(&models.Question{}, "id = ?", id).Error
}

func (r *Repository) GetAllQuizzesWithStats() ([]*types.QuizWithStats, error) {
	var quizzes []*models.Quiz
	err := r.db.Preload("Questions").Find(&quizzes).Error
	if err != nil {
		return nil, err
	}

	results := make([]*types.QuizWithStats, len(quizzes))
	for i, quiz := range quizzes {
		var attemptCount int64
		var avgScore float64

		r.db.Model(&models.QuizAttempt{}).Where("quiz_id = ? AND completed_at IS NOT NULL", quiz.ID).Count(&attemptCount)

		var scoreData struct {
			AvgScore   float64
			TotalScore int
			Count      int64
		}
		r.db.Model(&models.QuizAttempt{}).
			Select("COALESCE(AVG(score * 100.0 / NULLIF(max_score, 0)), 0) as avg_score, SUM(score) as total_score, COUNT(*) as count").
			Where("quiz_id = ? AND completed_at IS NOT NULL AND max_score > 0", quiz.ID).
			Scan(&scoreData)

		if scoreData.Count > 0 {
			avgScore = scoreData.AvgScore
		}

		results[i] = &types.QuizWithStats{
			Quiz:         quiz,
			AttemptCount: int(attemptCount),
			AvgScore:     avgScore,
		}
	}

	return results, nil
}

func (r *Repository) GetAllAttempts() ([]*types.AttemptWithUser, error) {
	var attempts []*types.AttemptWithUser
	err := r.db.Model(&models.QuizAttempt{}).
		Select("quiz_attempts.*, users.name as user_name, quizzes.title as quiz_title").
		Joins("JOIN users ON users.id = quiz_attempts.user_id").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.completed_at IS NOT NULL").
		Order("quiz_attempts.completed_at DESC").
		Scan(&attempts).Error
	return attempts, err
}

func (r *Repository) GetAttemptsByQuiz(quizID uuid.UUID) ([]*types.AttemptWithUser, error) {
	var attempts []*types.AttemptWithUser
	err := r.db.Model(&models.QuizAttempt{}).
		Select("quiz_attempts.*, users.name as user_name, quizzes.title as quiz_title").
		Joins("JOIN users ON users.id = quiz_attempts.user_id").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.quiz_id = ? AND quiz_attempts.completed_at IS NOT NULL", quizID).
		Order("quiz_attempts.completed_at DESC").
		Scan(&attempts).Error
	return attempts, err
}

func (r *Repository) GetStudentCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("role = ?", models.RoleStudent).Count(&count).Error
	return count, err
}

func (r *Repository) GetAdminStatistics() (*types.AdminStatisticsData, error) {
	var totalQuizzes int64
	r.db.Model(&models.Quiz{}).Where("status != ?", models.QuizStatusArchived).Count(&totalQuizzes)

	var totalStudents int64
	r.db.Model(&models.User{}).Where("role = ?", models.RoleStudent).Count(&totalStudents)

	var totalAttempts int64
	r.db.Model(&models.QuizAttempt{}).Where("completed_at IS NOT NULL").Count(&totalAttempts)

	var avgScore float64
	r.db.Model(&models.QuizAttempt{}).
		Select("COALESCE(AVG(score * 100.0 / NULLIF(max_score, 0)), 0)").
		Where("completed_at IS NOT NULL AND max_score > 0").
		Scan(&avgScore)

	quizStats, err := r.getQuizStatistics()
	if err != nil {
		return nil, err
	}

	gradeDist := make(map[int]int)
	r.db.Model(&models.User{}).
		Select("grade, COUNT(*) as count").
		Where("role = ?", models.RoleStudent).
		Group("grade").
		Scan(&gradeDist)

	subjDist := make(map[string]int)
	r.db.Model(&models.Quiz{}).
		Select("subject, COUNT(*) as count").
		Where("status != ?", models.QuizStatusArchived).
		Group("subject").
		Scan(&subjDist)

	return &types.AdminStatisticsData{
		TotalQuizzes:        int(totalQuizzes),
		TotalStudents:       int(totalStudents),
		TotalAttempts:       int(totalAttempts),
		AvgScore:            avgScore,
		QuizStats:           quizStats,
		GradeDistribution:   gradeDist,
		SubjectDistribution: subjDist,
	}, nil
}

func (r *Repository) getQuizStatistics() ([]*types.QuizStatistics, error) {
	var quizzes []*models.Quiz
	err := r.db.Where("status != ?", models.QuizStatusArchived).Find(&quizzes).Error
	if err != nil {
		return nil, err
	}

	stats := make([]*types.QuizStatistics, len(quizzes))
	for i, quiz := range quizzes {
		var attemptCount int64
		r.db.Model(&models.QuizAttempt{}).Where("quiz_id = ? AND completed_at IS NOT NULL", quiz.ID).Count(&attemptCount)

		var avgScore float64
		r.db.Model(&models.QuizAttempt{}).
			Select("COALESCE(AVG(score * 100.0 / NULLIF(max_score, 0)), 0)").
			Where("quiz_id = ? AND completed_at IS NOT NULL AND max_score > 0", quiz.ID).
			Scan(&avgScore)

		var passRate float64
		if attemptCount > 0 {
			var passCount int64
			r.db.Model(&models.QuizAttempt{}).
				Where("quiz_id = ? AND completed_at IS NOT NULL AND max_score > 0 AND score * 100.0 / max_score >= 60", quiz.ID).
				Count(&passCount)
			passRate = float64(passCount) / float64(attemptCount) * 100
		}

		stats[i] = &types.QuizStatistics{
			Quiz:         quiz,
			AttemptCount: int(attemptCount),
			AvgScore:     avgScore,
			PassRate:     passRate,
		}
	}

	return stats, nil
}

func (r *Repository) GetRecentAttempts(limit int) ([]*types.RecentAttempt, error) {
	var attempts []struct {
		AttemptID   uuid.UUID
		UserName    string
		QuizTitle   string
		Score       int
		MaxScore    int
		CompletedAt time.Time
	}

	err := r.db.Model(&models.QuizAttempt{}).
		Select("quiz_attempts.id as attempt_id, users.name as user_name, quizzes.title as quiz_title, quiz_attempts.score, quiz_attempts.max_score, quiz_attempts.completed_at").
		Joins("JOIN users ON users.id = quiz_attempts.user_id").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.completed_at IS NOT NULL").
		Order("quiz_attempts.completed_at DESC").
		Limit(limit).
		Scan(&attempts).Error

	if err != nil {
		return nil, err
	}

	results := make([]*types.RecentAttempt, len(attempts))
	for i, a := range attempts {
		results[i] = &types.RecentAttempt{
			AttemptID:   a.AttemptID.String(),
			UserName:    a.UserName,
			QuizTitle:   a.QuizTitle,
			Score:       a.Score,
			MaxScore:    a.MaxScore,
			CompletedAt: a.CompletedAt.Format("2006-01-02 15:04"),
		}
	}

	return results, nil
}
