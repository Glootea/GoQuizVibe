package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
	"github.com/goquizvibe/types"
)

type DashboardService struct {
	users          r.UserRepository
	quizzes        r.QuizRepository
	questions      r.QuestionRepository
	images         r.ImageRepository
	gamification   *GamificationService
	auth           Authenticator
	sessionService *QuizSessionService
}

func NewDashboardService(
	users r.UserRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	gamification *GamificationService,
	auth Authenticator,
	sessionService *QuizSessionService,
) *DashboardService {
	return &DashboardService{
		users:          users,
		quizzes:        quizzes,
		questions:      questions,
		images:         images,
		gamification:   gamification,
		auth:           auth,
		sessionService: sessionService,
	}
}

func (s *DashboardService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*types.DashboardData, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quizzes, err := s.getQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get quizzes: %w", err)
	}

	stats, err := s.gamification.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	leaderboard, err := s.gamification.GetLeaderboard(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}

	var activeSession *types.ActiveSessionInfo
	if s.sessionService != nil {
		activeSession, _ = s.sessionService.GetActiveSessionForUser(ctx, userID)
	}

	return &types.DashboardData{
		User:          &user,
		Quizzes:       quizzes,
		Stats:         stats,
		Leaderboard:   leaderboard,
		ActiveSession: activeSession,
	}, nil
}

func (s *DashboardService) GetUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, fmt.Errorf("get cookie: %w", err)
	}
	claims, err := s.auth.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate token: %w", err)
	}
	return claims.UserID, nil
}

func (s *DashboardService) getQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]*models.QuizWithQuestionsAndImages, error) {
	quizzes, err := s.quizzes.GetQuizzesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.QuizWithQuestionsAndImages, len(quizzes))
	for i, q := range quizzes {
		questions, _ := s.questions.GetQuestionsByQuizID(ctx, q.ID)
		questionsWithImages := s.attachImagesToQuestions(ctx, questions)
		result[i] = &models.QuizWithQuestionsAndImages{
			Quiz:      q,
			Questions: questionsWithImages,
		}
	}
	return result, err
}

func (s *DashboardService) attachImagesToQuestions(ctx context.Context, questions []db.Question) []models.QuestionWithImages {
	result := make([]models.QuestionWithImages, len(questions))
	for i, q := range questions {
		images, _ := s.images.GetImagesByQuestionID(ctx, q.ID)
		result[i] = models.QuestionWithImages{
			Question: q,
			Images:   images,
		}
	}
	return result
}
