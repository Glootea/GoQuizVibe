package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	gamificationSvc "github.com/goquizvibe/backend/feature/gamification/services"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/models"
)

func TestGamificationService_CalculateStreak(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	baseTime := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	t.Run("no attempts returns zero", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)
		tp := NewMockTimeProvider(baseTime)

		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, tp)
		streak, err := svc.CalculateStreak(ctx, userID)
		if err != nil {
			t.Fatalf("CalculateStreak() error = %v, want nil", err)
		}
		if streak != 0 {
			t.Errorf("CalculateStreak() = %d, want 0", streak)
		}
	})

	t.Run("consecutive days returns correct streak", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)
		tp := NewMockTimeProvider(baseTime)

		attempts := []db.QuizAttempt{
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime, Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -1), Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -2), Valid: true}},
		}
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, tp)
		streak, err := svc.CalculateStreak(ctx, userID)
		if err != nil {
			t.Fatalf("CalculateStreak() error = %v, want nil", err)
		}
		if streak != 3 {
			t.Errorf("CalculateStreak() = %d, want 3", streak)
		}
	})

	t.Run("gap breaks streak", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)
		tp := NewMockTimeProvider(baseTime)

		attempts := []db.QuizAttempt{
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime, Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -1), Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -3), Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -4), Valid: true}},
		}
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, tp)
		streak, err := svc.CalculateStreak(ctx, userID)
		if err != nil {
			t.Fatalf("CalculateStreak() error = %v, want nil", err)
		}
		if streak != 2 {
			t.Errorf("CalculateStreak() = %d, want 2", streak)
		}
	})

	t.Run("yesterday only not today returns zero", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)
		tp := NewMockTimeProvider(baseTime)

		attempts := []db.QuizAttempt{
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -1), Valid: true}},
		}
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, tp)
		streak, err := svc.CalculateStreak(ctx, userID)
		if err != nil {
			t.Fatalf("CalculateStreak() error = %v, want nil", err)
		}
		if streak != 0 {
			t.Errorf("CalculateStreak() = %d, want 0", streak)
		}
	})

	t.Run("today only returns one", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)
		tp := NewMockTimeProvider(baseTime)

		attempts := []db.QuizAttempt{
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime, Valid: true}},
		}
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, tp)
		streak, err := svc.CalculateStreak(ctx, userID)
		if err != nil {
			t.Fatalf("CalculateStreak() error = %v, want nil", err)
		}
		if streak != 1 {
			t.Errorf("CalculateStreak() = %d, want 1", streak)
		}
	})
}

func TestGamificationService_GetLeaderboard(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("empty attempts returns empty slice", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		m.Attempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return([]db.GetRecentAttemptsRow{}, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		entries, err := svc.GetLeaderboard(ctx, 100)
		if err != nil {
			t.Fatalf("GetLeaderboard() error = %v, want nil", err)
		}
		if len(entries) != 0 {
			t.Errorf("GetLeaderboard() len = %d, want 0", len(entries))
		}
	})

	t.Run("single user multiple attempts sums XP", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		rows := []db.GetRecentAttemptsRow{
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 50, MaxScore: 100},
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 30, MaxScore: 100},
		}
		m.Attempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(rows, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		entries, err := svc.GetLeaderboard(ctx, 100)
		if err != nil {
			t.Fatalf("GetLeaderboard() error = %v, want nil", err)
		}
		if len(entries) != 1 {
			t.Fatalf("GetLeaderboard() len = %d, want 1", len(entries))
		}
		if entries[0].XP != 80 {
			t.Errorf("GetLeaderboard() XP = %d, want 80", entries[0].XP)
		}
	})

	t.Run("multiple users sorted by XP desc", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		user2ID := uuid.New()
		rows := []db.GetRecentAttemptsRow{
			{ID: uuid.New(), UserID: userID, UserName: "User1", Score: 100, MaxScore: 100},
			{ID: uuid.New(), UserID: user2ID, UserName: "User2", Score: 50, MaxScore: 100},
		}
		m.Attempts.EXPECT().GetRecentAttempts(ctx, int32(100)).Return(rows, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		entries, err := svc.GetLeaderboard(ctx, 100)
		if err != nil {
			t.Fatalf("GetLeaderboard() error = %v, want nil", err)
		}
		if len(entries) != 2 {
			t.Fatalf("GetLeaderboard() len = %d, want 2", len(entries))
		}
		if entries[0].XP != 100 || entries[0].UserName != "User1" {
			t.Errorf("GetLeaderboard()[0] = %s with XP %d, want User1 with XP 100", entries[0].UserName, entries[0].XP)
		}
		if entries[1].XP != 50 || entries[1].UserName != "User2" {
			t.Errorf("GetLeaderboard()[1] = %s with XP %d, want User2 with XP 50", entries[1].UserName, entries[1].XP)
		}
	})
}

func TestGamificationService_GetUserStats(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("valid stats returns correct data", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		statsRow := db.GetUserStatsRow{
			TotalXp:    int64(500),
			CorrectCnt: 45,
			WrongCnt:   10,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.XP != 500 {
			t.Errorf("GetUserStats() XP = %d, want 500", stats.XP)
		}
		if stats.CorrectCount != 45 {
			t.Errorf("GetUserStats() CorrectCount = %d, want 45", stats.CorrectCount)
		}
		if stats.WrongCount != 10 {
			t.Errorf("GetUserStats() WrongCount = %d, want 10", stats.WrongCount)
		}
	})

	t.Run("nil total xp returns zero", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		statsRow := db.GetUserStatsRow{
			TotalXp:    nil,
			CorrectCnt: 0,
			WrongCnt:   0,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.XP != 0 {
			t.Errorf("GetUserStats() XP = %d, want 0", stats.XP)
		}
	})

	t.Run("GetLastActiveDate error handled gracefully", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		statsRow := db.GetUserStatsRow{
			TotalXp:    int64(100),
			CorrectCnt: 10,
			WrongCnt:   2,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, errors.New("no last active"))

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.XP != 100 {
			t.Errorf("GetUserStats() XP = %d, want 100", stats.XP)
		}
	})

	t.Run("lastActive as time.Time formats correctly", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		lastActiveTime := time.Date(2026, 5, 1, 14, 30, 0, 0, time.UTC)
		statsRow := db.GetUserStatsRow{
			TotalXp:    int64(200),
			CorrectCnt: 20,
			WrongCnt:   5,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(lastActiveTime, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.LastActiveDate != "2026-05-01 14:30" {
			t.Errorf("GetUserStats() LastActiveDate = %s, want 2026-05-01 14:30", stats.LastActiveDate)
		}
	})

	t.Run("lastActive as string passes through", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		statsRow := db.GetUserStatsRow{
			TotalXp:    int64(300),
			CorrectCnt: 30,
			WrongCnt:   8,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return([]db.QuizAttempt{}, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return("2026-04-28 10:00", nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.LastActiveDate != "2026-04-28 10:00" {
			t.Errorf("GetUserStats() LastActiveDate = %s, want 2026-04-28 10:00", stats.LastActiveDate)
		}
	})

	t.Run("GetAttemptsByUser error handled gracefully", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		statsRow := db.GetUserStatsRow{
			TotalXp:    int64(150),
			CorrectCnt: 15,
			WrongCnt:   3,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(nil, errors.New("db error"))
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.XP != 150 {
			t.Errorf("GetUserStats() XP = %d, want 150", stats.XP)
		}
		if len(stats.CompletedQuizzes) != 0 {
			t.Errorf("GetUserStats() CompletedQuizzes = %v, want empty or nil", stats.CompletedQuizzes)
		}
	})

	t.Run("calculateStreakFromAttempts via GetUserStats", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		baseTime := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
		attempts := []db.QuizAttempt{
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime, Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -1), Valid: true}},
			{ID: uuid.New(), UserID: userID, QuizID: uuid.New(), CompletedAt: sql.NullTime{Time: baseTime.AddDate(0, 0, -2), Valid: true}},
		}

		statsRow := db.GetUserStatsRow{
			TotalXp:    int64(250),
			CorrectCnt: 25,
			WrongCnt:   6,
		}
		m.Stats.EXPECT().GetUserStats(ctx, userID).Return(statsRow, nil)
		m.Attempts.EXPECT().GetAttemptsByUser(ctx, userID).Return(attempts, nil)
		m.Stats.EXPECT().GetLastActiveDate(ctx, userID).Return(nil, nil)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(baseTime))
		stats, err := svc.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats() error = %v, want nil", err)
		}
		if stats.Streak != 3 {
			t.Errorf("GetUserStats() Streak = %d, want 3", stats.Streak)
		}
	})
}

func TestSortByXP(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()
		entries := []*models.LeaderboardEntry{}
		gamificationSvc.SortByXP(entries)
		if len(entries) != 0 {
			t.Errorf("SortByXP() len = %d, want 0", len(entries))
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		entries := []*models.LeaderboardEntry{
			{UserID: "1", XP: 100},
		}
		gamificationSvc.SortByXP(entries)
		if len(entries) != 1 {
			t.Errorf("SortByXP() len = %d, want 1", len(entries))
		}
		if entries[0].XP != 100 {
			t.Errorf("SortByXP() entries[0].XP = %d, want 100", entries[0].XP)
		}
	})

	t.Run("already sorted", func(t *testing.T) {
		t.Parallel()
		entries := []*models.LeaderboardEntry{
			{UserID: "1", XP: 300},
			{UserID: "2", XP: 200},
			{UserID: "3", XP: 100},
		}
		gamificationSvc.SortByXP(entries)
		if entries[0].XP != 300 || entries[1].XP != 200 || entries[2].XP != 100 {
			t.Errorf("SortByXP() entries not sorted correctly")
		}
	})

	t.Run("reverse sorted", func(t *testing.T) {
		t.Parallel()
		entries := []*models.LeaderboardEntry{
			{UserID: "1", XP: 100},
			{UserID: "2", XP: 200},
			{UserID: "3", XP: 300},
		}
		gamificationSvc.SortByXP(entries)
		if entries[0].XP != 300 || entries[1].XP != 200 || entries[2].XP != 100 {
			t.Errorf("SortByXP() entries not sorted correctly")
		}
	})

	t.Run("ties maintain order", func(t *testing.T) {
		t.Parallel()
		entries := []*models.LeaderboardEntry{
			{UserID: "1", XP: 100},
			{UserID: "2", XP: 100},
			{UserID: "3", XP: 100},
		}
		gamificationSvc.SortByXP(entries)
		if entries[0].XP != 100 || entries[1].XP != 100 || entries[2].XP != 100 {
			t.Errorf("SortByXP() ties not handled correctly")
		}
	})
}

func TestUpdateStreak(t *testing.T) {
	t.Run("returns nil", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		err := svc.UpdateStreak(context.Background(), uuid.New())
		if err != nil {
			t.Errorf("UpdateStreak() error = %v, want nil", err)
		}
	})
}

func TestAwardXP(t *testing.T) {
	t.Run("returns nil", func(t *testing.T) {
		t.Parallel()
		m := NewGamificationServiceMocks(t)

		svc := gamificationSvc.NewGamificationService(m.Attempts, m.Stats, NewMockTimeProvider(time.Now()))
		err := svc.AwardXP(context.Background(), uuid.New(), 50)
		if err != nil {
			t.Errorf("AwardXP() error = %v, want nil", err)
		}
	})
}
