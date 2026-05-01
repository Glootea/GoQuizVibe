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
	repo *store.Repository
}

func NewQuizService(r *store.Repository) *QuizService {
	return &QuizService{repo: r}
}

func (s *QuizService) GetQuizzesForUser(userID uuid.UUID) ([]*models.Quiz, error) {
	return s.repo.GetQuizzesForUser(userID)
}

func (s *QuizService) GetQuizByID(id uuid.UUID) (*models.Quiz, error) {
	return s.repo.GetQuizByID(id)
}

func (s *QuizService) SubmitQuizAttempt(userID, quizID uuid.UUID, answers map[uuid.UUID]string) (*models.QuizAttempt, error) {
	quiz, err := s.repo.GetQuizWithQuestions(quizID)
	if err != nil {
		return nil, err
	}

	attempt := &models.QuizAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: time.Now(),
	}

	var score, maxScore int
	for _, q := range quiz.Questions {
		maxScore += q.Points
		userAnswer := answers[q.ID]
		isCorrect := NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		attempt.Answers = append(attempt.Answers, models.UserAnswer{
			ID:         uuid.New(),
			AttemptID:  attempt.ID,
			QuestionID: q.ID,
			UserAnswer: userAnswer,
			IsCorrect:  isCorrect,
		})

		if isCorrect {
			score += q.Points
		}
	}

	attempt.Score = score
	attempt.MaxScore = maxScore
	now := time.Now()
	attempt.CompletedAt = &now

	if err := s.repo.SaveAttempt(attempt); err != nil {
		return nil, err
	}

	for _, a := range attempt.Answers {
		if err := s.repo.SaveUserAnswer(&a); err != nil {
			return nil, err
		}
	}

	return attempt, nil
}

type GamificationService struct {
	repo *store.Repository
}

func NewGamificationService(r *store.Repository) *GamificationService {
	return &GamificationService{repo: r}
}

func (s *GamificationService) UpdateStreak(userID uuid.UUID) error {
	return nil
}

func (s *GamificationService) AwardXP(userID uuid.UUID, amount int) error {
	return nil
}

func (s *GamificationService) GetLeaderboard() []*models.LeaderboardEntry {
	entries, _ := s.repo.GetLeaderboard(100)
	return entries
}

func (s *GamificationService) GetUserStats(userID uuid.UUID) (*models.UserStats, error) {
	return s.repo.GetUserStats(userID)
}