package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/jackc/pgx/v5"
)

var (
	ErrEmailExists  = errors.New("email already registered")
	ErrUserNotFound = errors.New("user not found")
	ErrQuizNotFound = errors.New("quiz not found")
	ErrNotFound     = errors.New("record not found")
)

type Repository struct {
	pool *db.Queries
}

func NewRepository(pool *db.Queries) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Close() {}

func (r *Repository) CreateUser(ctx context.Context, u *db.User) error {
	exists, err := r.pool.EmailExists(ctx, u.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailExists
	}
	_, err = r.pool.CreateUser(ctx, db.CreateUserParams{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
	})
	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.pool.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := r.pool.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

type QuizWithQuestions struct {
	*db.Quiz
	Questions []db.Question
}

func (q *QuizWithQuestions) GetOptions(index int) []string {
	if index < 0 || index >= len(q.Questions) {
		return nil
	}
	return GetOptions(q.Questions[index])
}

func (r *Repository) GetQuizzes(ctx context.Context) ([]*QuizWithQuestions, error) {
	quizzes, err := r.pool.GetAvailableQuizzes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*QuizWithQuestions, len(quizzes))
	for i := range quizzes {
		q := &quizzes[i]
		questions, err := r.pool.GetQuestionsByQuizID(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		result[i] = &QuizWithQuestions{Quiz: q, Questions: questions}
	}
	return result, nil
}

func (r *Repository) GetQuizByID(ctx context.Context, id uuid.UUID) (*QuizWithQuestions, error) {
	quiz, err := r.pool.GetQuizByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQuizNotFound
		}
		return nil, err
	}
	questions, err := r.pool.GetQuestionsByQuizID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &QuizWithQuestions{Quiz: &quiz, Questions: questions}, nil
}

func (r *Repository) GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*QuizWithQuestions, error) {
	quizzes, err := r.pool.GetQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*QuizWithQuestions, len(quizzes))
	for i := range quizzes {
		q := &quizzes[i]
		questions, err := r.pool.GetQuestionsByQuizID(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		result[i] = &QuizWithQuestions{Quiz: q, Questions: questions}
	}
	return result, nil
}

func (r *Repository) SaveAttempt(ctx context.Context, attempt *db.QuizAttempt) error {
	_, err := r.pool.CreateAttempt(ctx, db.CreateAttemptParams{
		ID:        attempt.ID,
		UserID:    attempt.UserID,
		QuizID:    attempt.QuizID,
		StartedAt: attempt.StartedAt,
	})
	return err
}

func (r *Repository) GetAttemptsByUser(ctx context.Context, userID uuid.UUID) ([]*db.QuizAttempt, error) {
	attempts, err := r.pool.GetAttemptsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*db.QuizAttempt, len(attempts))
	for i := range attempts {
		result[i] = &attempts[i]
	}
	return result, nil
}

func (r *Repository) GetAttemptByID(ctx context.Context, id uuid.UUID) (*db.QuizAttempt, error) {
	attempt, err := r.pool.GetAttemptByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &attempt, nil
}

func (r *Repository) UpdateAttempt(ctx context.Context, attempt *db.QuizAttempt) error {
	_, err := r.pool.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          attempt.ID,
		Score:       attempt.Score,
		MaxScore:    attempt.MaxScore,
		CompletedAt: attempt.CompletedAt,
	})
	return err
}

func (r *Repository) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	stats, err := r.pool.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	lastActive, _ := r.pool.GetLastActiveDate(ctx, userID)
	_, _ = r.pool.GetCompletedAttemptsCount(ctx, userID)

	streak := r.calculateStreak(ctx, userID)

	var lastActiveStr string
	if lastActive != nil {
		if t, ok := lastActive.(time.Time); ok && !t.IsZero() {
			lastActiveStr = t.Format("2006-01-02")
		}
	}

	var totalXP int
	if stats.TotalXp != nil {
		if v, ok := stats.TotalXp.(int64); ok {
			totalXP = int(v)
		}
	}

	return &models.UserStats{
		UserID:           userID.String(),
		XP:               totalXP,
		Streak:           streak,
		LastActiveDate:   lastActiveStr,
		CompletedQuizzes: nil,
		CorrectCount:     int(stats.CorrectCnt),
		WrongCount:       int(stats.WrongCnt),
	}, nil
}

func (r *Repository) calculateStreak(ctx context.Context, userID uuid.UUID) int {
	attempts, err := r.GetAttemptsByUser(ctx, userID)
	if err != nil || len(attempts) == 0 {
		return 0
	}

	streak := 1
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var completed []*db.QuizAttempt
	for _, a := range attempts {
		if !a.CompletedAt.IsZero() {
			completed = append(completed, a)
		}
	}

	if len(completed) == 0 {
		return 0
	}

	for i := 0; i < len(completed)-1; i++ {
		if completed[i].CompletedAt.IsZero() {
			continue
		}
		current := time.Date(completed[i].CompletedAt.Year(), completed[i].CompletedAt.Month(), completed[i].CompletedAt.Day(), 0, 0, 0, 0, now.Location())

		var nextDay time.Time
		if i+1 < len(completed) && !completed[i+1].CompletedAt.IsZero() {
			nextDay = time.Date(completed[i+1].CompletedAt.Year(), completed[i+1].CompletedAt.Month(), completed[i+1].CompletedAt.Day(), 0, 0, 0, 0, now.Location())
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

	if len(completed) > 0 && !completed[0].CompletedAt.IsZero() {
		lastActive := time.Date(completed[0].CompletedAt.Year(), completed[0].CompletedAt.Month(), completed[0].CompletedAt.Day(), 0, 0, 0, 0, now.Location())
		if lastActive.Before(today.AddDate(0, 0, -1)) {
			return 0
		}
	}

	return streak
}

func (r *Repository) GetLeaderboard(ctx context.Context, limit int32) ([]*models.LeaderboardEntry, error) {
	quizzes, err := r.pool.GetQuizzes(ctx)
	if err != nil {
		return nil, err
	}

	type xpEntry struct {
		UserID   uuid.UUID
		UserName string
		TotalXP  int
	}

	xpMap := make(map[uuid.UUID]*xpEntry)

	for _, quiz := range quizzes {
		attempts, err := r.GetAttemptsByUser(ctx, quiz.CreatedBy)
		if err != nil {
			continue
		}

		for _, a := range attempts {
			answers, err := r.pool.GetAnswersByAttempt(ctx, a.ID)
			if err != nil {
				continue
			}

			for _, ans := range answers {
				if ans.IsCorrect {
					questions, err := r.pool.GetQuestionsByQuizID(ctx, a.QuizID)
					if err != nil {
						continue
					}
					for _, q := range questions {
						if q.ID == ans.QuestionID {
							pts := q.Points
							if xpMap[a.UserID] == nil {
								user, _ := r.GetUserByID(ctx, a.UserID)
								name := ""
								if user != nil {
									name = user.Name
								}
								xpMap[a.UserID] = &xpEntry{
									UserID:   a.UserID,
									UserName: name,
								}
							}
							xpMap[a.UserID].TotalXP += int(pts)
						}
					}
				}
			}
		}
	}

	var entries []*models.LeaderboardEntry
	for _, e := range xpMap {
		entries = append(entries, &models.LeaderboardEntry{
			UserID:   e.UserID.String(),
			UserName: e.UserName,
			XP:       e.TotalXP,
			Streak:   r.calculateStreak(ctx, e.UserID),
		})
	}

	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

func (r *Repository) CreateSession(ctx context.Context, session *db.QuizSession) error {
	_, err := r.pool.CreateSession(ctx, db.CreateSessionParams{
		ID:           session.ID,
		UserID:       session.UserID,
		QuizID:       session.QuizID,
		AttemptID:    session.AttemptID,
		CurrentIndex: session.CurrentIndex,
		Answers:      session.Answers,
		CreatedAt:    session.CreatedAt,
	})
	return err
}

func (r *Repository) GetSession(ctx context.Context, sessionID uuid.UUID) (*db.QuizSession, error) {
	session, err := r.pool.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (r *Repository) GetSessionByAttemptID(ctx context.Context, attemptID uuid.UUID) (*db.QuizSession, error) {
	session, err := r.pool.GetSessionByAttemptID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (r *Repository) UpdateSession(ctx context.Context, session *db.QuizSession) error {
	_, err := r.pool.UpdateSession(ctx, db.UpdateSessionParams{
		ID:           session.ID,
		CurrentIndex: session.CurrentIndex,
		Answers:      session.Answers,
	})
	return err
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return r.pool.DeleteSession(ctx, sessionID)
}

func (r *Repository) SaveUserAnswer(ctx context.Context, answer *db.UserAnswer) error {
	_, err := r.pool.CreateUserAnswer(ctx, db.CreateUserAnswerParams{
		ID:         answer.ID,
		AttemptID:  answer.AttemptID,
		QuestionID: answer.QuestionID,
		UserAnswer: answer.UserAnswer,
		IsCorrect:  answer.IsCorrect,
	})
	return err
}

func (r *Repository) GetAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*db.UserAnswer, error) {
	answers, err := r.pool.GetAnswersByAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	result := make([]*db.UserAnswer, len(answers))
	for i := range answers {
		result[i] = &answers[i]
	}
	return result, nil
}

func (r *Repository) GetUserAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*db.UserAnswer, error) {
	return r.GetAnswersByAttempt(ctx, attemptID)
}

func (r *Repository) GetQuizWithQuestions(ctx context.Context, quizID uuid.UUID) (*QuizWithQuestions, error) {
	return r.GetQuizByID(ctx, quizID)
}

func (r *Repository) GetCompletedAttemptBySessionID(ctx context.Context, sessionID uuid.UUID) (*db.QuizAttempt, error) {
	attempt, err := r.pool.GetCompletedAttemptBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &attempt, nil
}

func (r *Repository) GetQuizErrors(ctx context.Context, userID uuid.UUID) ([]*db.QuizAttempt, error) {
	attempts, err := r.pool.GetQuizErrors(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*db.QuizAttempt, len(attempts))
	for i := range attempts {
		result[i] = &attempts[i]
	}
	return result, nil
}

func (r *Repository) GetWrongAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]*db.UserAnswer, error) {
	answers, err := r.pool.GetWrongAnswersByAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	result := make([]*db.UserAnswer, len(answers))
	for i := range answers {
		result[i] = &answers[i]
	}
	return result, nil
}

func GetOptions(q db.Question) []string {
	if q.Options == nil {
		return []string{}
	}
	var opts []string
	if err := json.Unmarshal(q.Options, &opts); err != nil {
		return []string{}
	}
	return opts
}
