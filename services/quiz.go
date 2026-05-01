package services

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/store"
)

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

type QuizService struct {
	store *store.MemoryStore
}

func NewQuizService(s *store.MemoryStore) *QuizService {
	return &QuizService{store: s}
}

func (s *QuizService) GetQuizzesForUser(userID uuid.UUID) []*models.Quiz {
	return s.store.GetQuizzesForUser(userID)
}

func (s *QuizService) GetQuizByID(id uuid.UUID) (*models.Quiz, error) {
	return s.store.GetQuizByID(id)
}

func (s *QuizService) SubmitQuizAttempt(userID, quizID uuid.UUID, answers map[uuid.UUID]string) (*models.QuizAttempt, error) {
	quiz, err := s.store.GetQuizByID(quizID)
	if err != nil {
		return nil, err
	}

	attempt := &models.QuizAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		Answers:   make([]models.UserAnswer, 0),
		StartedAt: time.Now(),
	}

	var score, maxScore int
	for _, q := range quiz.Questions {
		maxScore += q.Points
		userAnswer := answers[q.ID]
		isCorrect := NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		attempt.Answers = append(attempt.Answers, models.UserAnswer{
			QuestionID: q.ID,
			UserAnswer: userAnswer,
			IsCorrect:  isCorrect,
		})

		if isCorrect {
			score += q.Points
		} else {
			s.store.AddWrongAnswer(userID, models.WrongAnswer{
				ID:            uuid.New(),
				QuestionID:    q.ID,
				QuizID:        quizID,
				UserAnswer:    userAnswer,
				CorrectAnswer: q.CorrectAnswer,
				Timestamp:     time.Now(),
			})
		}
	}

	attempt.Score = score
	attempt.MaxScore = maxScore
	attempt.CompletedAt = time.Now()

	s.store.SaveAttempt(attempt)

	progress, _ := s.store.GetProgress(userID)
	if progress != nil {
		progress.CompletedQuizzes = append(progress.CompletedQuizzes, quizID)
		s.store.UpdateProgress(progress)
	}

	return attempt, nil
}

type GamificationService struct {
	store *store.MemoryStore
}

func NewGamificationService(s *store.MemoryStore) *GamificationService {
	return &GamificationService{store: s}
}

func (s *GamificationService) UpdateStreak(userID uuid.UUID) error {
	progress, err := s.store.GetProgress(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	lastActive := time.Date(progress.LastActiveDate.Year(), progress.LastActiveDate.Month(), progress.LastActiveDate.Day(), 0, 0, 0, 0, progress.LastActiveDate.Location())

	if lastActive.Equal(today) {
		return nil
	}

	if lastActive.Equal(yesterday) {
		progress.Streak++
	} else if !lastActive.IsZero() {
		progress.Streak = 1
	} else {
		progress.Streak = 1
	}

	progress.LastActiveDate = now
	return s.store.UpdateProgress(progress)
}

func (s *GamificationService) AwardXP(userID uuid.UUID, amount int) error {
	progress, err := s.store.GetProgress(userID)
	if err != nil {
		return err
	}

	progress.XP += amount
	return s.store.UpdateProgress(progress)
}

func (s *GamificationService) GetLeaderboard() []*models.LeaderboardEntry {
	return s.store.GetLeaderboard()
}

func (s *GamificationService) GetUserStats(userID uuid.UUID) (*models.UserProgress, error) {
	return s.store.GetProgress(userID)
}
