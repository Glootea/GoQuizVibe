package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/models"
	r "github.com/goquizvibe/backend/shared/repositories"
	cache "github.com/goquizvibe/backend/shared/infrastructure/cache"
)

func NormalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}

func CheckFillAnswer(userAnswers []string, correctAnswerJSON string) bool {
	if len(userAnswers) == 0 {
		return false
	}

	segments, err := models.ParseFillSegments(correctAnswerJSON)
	if err != nil || len(segments) == 0 {
		return false
	}

	gapCount := 0
	for _, seg := range segments {
		if seg.Type == "gap" {
			gapCount++
		}
	}

	if gapCount == 0 {
		return false
	}

	if len(userAnswers) != gapCount {
		return false
	}

	gapIndex := 0
	for _, seg := range segments {
		if seg.Type == "gap" {
			if NormalizeAnswer(userAnswers[gapIndex]) != NormalizeAnswer(seg.Content) {
				return false
			}
			gapIndex++
		}
	}
	return true
}

type QuizService struct {
	quizzes   r.QuizRepository
	questions r.QuestionRepository
	images    r.ImageRepository
	attempts  r.AttemptRepository
	groups    r.UserGroupRepository
	cache     *cache.CacheService
}

func NewQuizService(quizzes r.QuizRepository, questions r.QuestionRepository, images r.ImageRepository, attempts r.AttemptRepository, groups r.UserGroupRepository, cache *cache.CacheService) *QuizService {
	return &QuizService{
		quizzes:   quizzes,
		questions: questions,
		images:    images,
		attempts:  attempts,
		groups:    groups,
		cache:     cache,
	}
}

func (s *QuizService) GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*models.QuizWithQuestionsAndImages, error) {
	cacheKey := "quizzes:user:" + userID.String()
	quizzes, err := cache.GetOrFetch(ctx, s.cache, cacheKey, func() ([]db.Quiz, error) {
		groups, err := s.groups.GetUserGroupsByAdmin(ctx, userID)
		if err != nil {
			return nil, err
		}
		groupIDs := make([]uuid.UUID, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.ID
		}
		return s.quizzes.GetQuizzesForUser(ctx, db.GetQuizzesForUserParams{
			RecipientID: userID,
			Column2:     groupIDs,
		})
	})
	if err != nil {
		return nil, err
	}
	result := make([]*models.QuizWithQuestionsAndImages, len(quizzes))
	for i, q := range quizzes {
		questions, _ := s.questions.GetQuestionsByQuizID(ctx, q.ID)
		questionsWithImages := cache.AttachImagesToQuestions(ctx, questions, s.images)
		result[i] = &models.QuizWithQuestionsAndImages{
			Quiz:      q,
			Questions: questionsWithImages,
		}
	}
	return result, err
}

func (s *QuizService) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.QuizWithQuestionsAndImages, error) {
	cacheKey := "quiz:" + id.String()
	quiz, err := cache.GetOrFetch(ctx, s.cache, cacheKey, func() (db.Quiz, error) {
		return s.quizzes.GetQuizByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	questionsCacheKey := "questions:quiz:" + id.String()
	questions, err := cache.GetOrFetch(ctx, s.cache, questionsCacheKey, func() ([]db.Question, error) {
		return s.questions.GetQuestionsByQuizID(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	questionsWithImages := cache.AttachImagesToQuestions(ctx, questions, s.images)
	return &models.QuizWithQuestionsAndImages{
		Quiz:      quiz,
		Questions: questionsWithImages,
	}, nil
}

func (s *QuizService) SubmitQuizAttempt(ctx context.Context, userID, quizID uuid.UUID, answers map[uuid.UUID]string) (*db.QuizAttempt, error) {
	_, err := s.quizzes.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	questions, err := s.questions.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	attemptID := uuid.New()
	now := time.Now()

	if _, err := s.attempts.CreateAttempt(ctx, db.CreateAttemptParams{
		ID:        attemptID,
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: now,
	}); err != nil {
		return nil, err
	}

	var score, maxScore int

	for _, q := range questions {
		maxScore += int(q.Points)
		userAnswer := answers[q.ID]

		isCorrect := NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		if _, err := s.attempts.UpsertUserAnswer(ctx, db.UpsertUserAnswerParams{
			AttemptID:  attemptID,
			QuestionID: q.ID,
			UserAnswer: userAnswer,
			IsCorrect:  isCorrect,
		}); err != nil {
			return nil, err
		}

		if isCorrect {
			score += int(q.Points)
		}
	}

	completedAt := sql.NullTime{Time: time.Now(), Valid: true}
	attempt, err := s.attempts.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          attemptID,
		Score:       score,
		MaxScore:    maxScore,
		CompletedAt: completedAt,
	})
	if err != nil {
		return nil, err
	}

	return &attempt, nil
}
