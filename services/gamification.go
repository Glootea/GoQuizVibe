package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
)

type GamificationService struct {
	attempts     r.AttemptRepository
	stats        r.StatsRepository
	timeProvider TimeProvider
}

func NewGamificationService(attempts r.AttemptRepository, stats r.StatsRepository, tp TimeProvider) *GamificationService {
	return &GamificationService{
		attempts:     attempts,
		stats:        stats,
		timeProvider: tp,
	}
}

func (s *GamificationService) CalculateStreak(ctx context.Context, userID uuid.UUID) (int, error) {
	attempts, err := s.attempts.GetAttemptsByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	if len(attempts) == 0 {
		return 0, nil
	}

	dailyAttempts := make(map[string]int)
	for _, a := range attempts {
		day := a.CompletedAt.Format("2006-01-02")
		dailyAttempts[day]++
	}

	streak := 0
	currentDate := s.timeProvider.Now().Format("2006-01-02")
	for {
		if count, ok := dailyAttempts[currentDate]; ok && count > 0 {
			streak++
			t, _ := time.Parse("2006-01-02", currentDate)
			currentDate = t.AddDate(0, 0, -1).Format("2006-01-02")
		} else {
			break
		}
	}

	return streak, nil
}

func (s *GamificationService) GetLeaderboard(ctx context.Context, limit int) ([]*models.LeaderboardEntry, error) {
	attempts, err := s.attempts.GetRecentAttempts(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	leaderboardMap := make(map[string]*models.LeaderboardEntry)
	for _, attempt := range attempts {
		key := attempt.UserID.String()
		if entry, ok := leaderboardMap[key]; ok {
			entry.XP += int(attempt.Score)
		} else {
			leaderboardMap[key] = &models.LeaderboardEntry{
				UserID:   attempt.UserID.String(),
				UserName: attempt.UserName,
				XP:       int(attempt.Score),
			}
		}
	}

	entries := make([]*models.LeaderboardEntry, 0, len(leaderboardMap))
	for _, entry := range leaderboardMap {
		entries = append(entries, entry)
	}

	SortByXP(entries)

	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

func SortByXP(entries []*models.LeaderboardEntry) {
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].XP > entries[i].XP {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

func (s *GamificationService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	stats, err := s.stats.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	xp, ok := stats.TotalXp.(int64)
	if !ok {
		xp = 0
	}

	lastActive, err := s.stats.GetLastActiveDate(ctx, userID)
	if err != nil {
		lastActive = nil
	}

	var lastActiveStr string
	if lastActive != nil {
		switch v := lastActive.(type) {
		case time.Time:
			lastActiveStr = v.Format("2006-01-02 15:04")
		case string:
			lastActiveStr = v
		}
	}

	attempts, err := s.attempts.GetAttemptsByUser(ctx, userID)
	if err != nil {
		attempts = nil
	}

	streak := s.calculateStreakFromAttempts(attempts)

	completedQuizzes := make([]string, 0)
	for _, a := range attempts {
		completedQuizzes = append(completedQuizzes, a.QuizID.String())
	}

	return &models.UserStats{
		UserID:           userID.String(),
		XP:               int(xp),
		Streak:           streak,
		LastActiveDate:   lastActiveStr,
		CompletedQuizzes: completedQuizzes,
		CorrectCount:     int(stats.CorrectCnt),
		WrongCount:       int(stats.WrongCnt),
	}, nil
}

func (s *GamificationService) calculateStreakFromAttempts(attempts []db.QuizAttempt) int {
	if len(attempts) == 0 {
		return 0
	}

	dailyAttempts := make(map[string]int)
	for _, a := range attempts {
		day := a.CompletedAt.Format("2006-01-02")
		dailyAttempts[day]++
	}

	streak := 0
	currentDate := s.timeProvider.Now().Format("2006-01-02")
	for {
		if count, ok := dailyAttempts[currentDate]; ok && count > 0 {
			streak++
			t, _ := time.Parse("2006-01-02", currentDate)
			currentDate = t.AddDate(0, 0, -1).Format("2006-01-02")
		} else {
			break
		}
	}

	return streak
}

func (s *GamificationService) UpdateStreak(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *GamificationService) AwardXP(ctx context.Context, userID uuid.UUID, amount int) error {
	return nil
}
