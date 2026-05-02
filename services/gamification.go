package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
)

type GamificationService struct {
	pool *db.Queries
}

func NewGamificationService(pool *db.Queries) *GamificationService {
	return &GamificationService{pool: pool}
}

func (s *GamificationService) GetLeaderboard(ctx context.Context, limit int) ([]*models.LeaderboardEntry, error) {
	attempts, err := s.pool.GetRecentAttempts(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	leaderboardMap := make(map[string]*models.LeaderboardEntry)
	for _, attempt := range attempts {
		key := attempt.UserID.String()
		if entry, ok := leaderboardMap[key]; ok {
			entry.XP += attempt.Score
		} else {
			leaderboardMap[key] = &models.LeaderboardEntry{
				UserID:   attempt.UserID.String(),
				UserName: attempt.UserName,
				XP:       attempt.Score,
			}
		}
	}

	entries := make([]*models.LeaderboardEntry, 0, len(leaderboardMap))
	for _, entry := range leaderboardMap {
		entries = append(entries, entry)
	}

	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

func (s *GamificationService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	stats, err := s.pool.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	xp, ok := stats.TotalXp.(int64)
	if !ok {
		xp = 0
	}

	lastActive, err := s.pool.GetLastActiveDate(ctx, userID)
	if err != nil {
		lastActive = nil
	}

	var lastActiveStr string
	if lastActive != nil {
		lastActiveStr = lastActive.(string)
	}

	attempts, err := s.pool.GetAttemptsByUser(ctx, userID)
	if err != nil {
		attempts = nil
	}

	completedQuizzes := make([]string, 0)
	for _, a := range attempts {
		if !a.CompletedAt.IsZero() {
			completedQuizzes = append(completedQuizzes, a.QuizID.String())
		}
	}

	return &models.UserStats{
		UserID:           userID.String(),
		XP:               int(xp),
		LastActiveDate:   lastActiveStr,
		CompletedQuizzes: completedQuizzes,
		CorrectCount:     int(stats.CorrectCnt),
		WrongCount:       int(stats.WrongCnt),
	}, nil
}

func (s *GamificationService) UpdateStreak(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *GamificationService) AwardXP(ctx context.Context, userID uuid.UUID, amount int) error {
	return nil
}