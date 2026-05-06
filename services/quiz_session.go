package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/types"
)

type QuizSessionService struct {
	pool         *db.Queries
	quizService  *QuizService
	gamification *GamificationService
	authService  *AuthService
}

func NewQuizSessionService(pool *db.Queries, qs *QuizService, gs *GamificationService, as *AuthService) *QuizSessionService {
	return &QuizSessionService{
		pool:         pool,
		quizService:  qs,
		gamification: gs,
		authService:  as,
	}
}

func (s *QuizSessionService) CreateSession(ctx context.Context, userID, quizID uuid.UUID) (*db.QuizSession, error) {
	attemptID := uuid.New()
	now := time.Now()

	_, err := s.pool.CreateAttempt(ctx, db.CreateAttemptParams{
		ID:        attemptID,
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create attempt: %w", err)
	}

	session := &db.QuizSession{
		ID:           uuid.New(),
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attemptID,
		CurrentIndex: 0,
		Answers:      nil,
		CreatedAt:    now,
	}

	_, err = s.pool.CreateSession(ctx, db.CreateSessionParams{
		ID:           session.ID,
		UserID:       session.UserID,
		QuizID:       session.QuizID,
		AttemptID:    session.AttemptID,
		CurrentIndex: session.CurrentIndex,
		Answers:      session.Answers,
		CreatedAt:    session.CreatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

type QuestionFeedback struct {
	IsCorrect     bool
	CorrectAnswer string
	Explanation   string
	IsLast        bool
}

func (s *QuizSessionService) SubmitAnswer(ctx context.Context, sessionID uuid.UUID, quizID uuid.UUID, questionIndex int, answer string) (*QuestionFeedback, error) {
	quiz, err := s.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	if questionIndex >= len(quiz.Questions) {
		return nil, fmt.Errorf("question index out of range")
	}

	session, err := s.pool.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	question := quiz.Questions[questionIndex]
	isCorrect := NormalizeAnswer(answer) == NormalizeAnswer(question.CorrectAnswer)

	var answers map[int]string
	if session.Answers != nil {
		if err := json.Unmarshal(session.Answers, &answers); err != nil {
			return nil, fmt.Errorf("unmarshal answers: %w", err)
		}
	} else {
		answers = make(map[int]string)
	}
	answers[questionIndex] = answer

	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return nil, fmt.Errorf("marshal answers: %w", err)
	}

	_, err = s.pool.UpdateSession(ctx, db.UpdateSessionParams{
		ID:           session.ID,
		CurrentIndex: questionIndex + 1,
		Answers:      answersJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	_, err = s.pool.CreateUserAnswer(ctx, db.CreateUserAnswerParams{
		ID:         uuid.New(),
		AttemptID:  session.AttemptID,
		QuestionID: question.ID,
		UserAnswer: answer,
		IsCorrect:  isCorrect,
	})
	if err != nil {
		return nil, fmt.Errorf("create user answer: %w", err)
	}

	return &QuestionFeedback{
		IsCorrect:     isCorrect,
		CorrectAnswer: question.CorrectAnswer,
		Explanation:   question.Explanation,
		IsLast:        questionIndex >= len(quiz.Questions)-1,
	}, nil
}

func (s *QuizSessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) (*db.QuizAttempt, error) {
	session, err := s.pool.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	quiz, err := s.quizService.GetQuizByID(ctx, session.QuizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	answers, err := s.pool.GetAnswersByAttempt(ctx, session.AttemptID)
	if err != nil {
		return nil, fmt.Errorf("get answers: %w", err)
	}

	var score, maxScore int
	for _, q := range quiz.Questions {
		maxScore += int(q.Points)
		for _, a := range answers {
			if a.QuestionID == q.ID && a.IsCorrect {
				score += int(q.Points)
				break
			}
		}
	}

	now := time.Now()
	updatedAttempt, err := s.pool.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          session.AttemptID,
		Score:       score,
		MaxScore:    maxScore,
		CompletedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("update attempt: %w", err)
	}

	err = s.pool.DeleteSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("delete session: %w", err)
	}

	return &updatedAttempt, nil
}

func (s *QuizSessionService) GetQuizResultData(ctx context.Context, quizID, sessionID, userID uuid.UUID) (*types.QuizResultData, error) {
	user, err := s.pool.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quiz, err := s.quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	stats, err := s.gamification.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	attempt, err := s.CompleteSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("complete session: %w", err)
	}

	answers, err := s.pool.GetAnswersByAttempt(ctx, attempt.ID)
	if err != nil {
		return nil, fmt.Errorf("get answers: %w", err)
	}

	answersMap := make(map[uuid.UUID]string)
	var correctCount, wrongCount int
	for _, a := range answers {
		answersMap[a.QuestionID] = a.UserAnswer
		if a.IsCorrect {
			correctCount++
		} else {
			wrongCount++
		}
	}

	var answerDetails []types.AnswerDetail
	for _, q := range quiz.Questions {
		userAnswer := answersMap[q.ID]
		if userAnswer == "" {
			userAnswer = "Нет ответа"
		}
		isCorrect := userAnswer != "" && NormalizeAnswer(userAnswer) == NormalizeAnswer(q.CorrectAnswer)

		answerDetails = append(answerDetails, types.AnswerDetail{
			Question:      q.Text,
			UserAnswer:    userAnswer,
			CorrectAnswer: q.CorrectAnswer,
			IsCorrect:     isCorrect,
		})
	}

	return &types.QuizResultData{
		User:         &user,
		Quiz:         quiz,
		Stats:        stats,
		Score:        int(attempt.Score),
		MaxScore:     int(attempt.MaxScore),
		CorrectCount: correctCount,
		WrongCount:   wrongCount,
		Answers:      answerDetails,
	}, nil
}

func (s *QuizSessionService) GetErrorsPageData(ctx context.Context, userID uuid.UUID) (*types.ErrorsPageData, error) {
	user, err := s.pool.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	stats, err := s.gamification.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	attempts, err := s.pool.GetQuizErrors(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get quiz errors: %w", err)
	}

	var quizErrors []types.QuizErrors
	for _, attempt := range attempts {
		quiz, err := s.quizService.GetQuizByID(ctx, attempt.QuizID)
		if err != nil {
			continue
		}

		answers, err := s.pool.GetAnswersByAttempt(ctx, attempt.ID)
		if err != nil {
			continue
		}

		questionMap := make(map[uuid.UUID]models.Question)
		for _, q := range quiz.Questions {
			questionMap[q.ID] = q
		}

		var wrongAnswers []models.WrongAnswer
		for _, a := range answers {
			if !a.IsCorrect {
				q := questionMap[a.QuestionID]
				wrongAnswers = append(wrongAnswers, models.WrongAnswer{
					ID:            a.ID,
					QuestionID:    a.QuestionID,
					QuizID:        attempt.QuizID,
					UserAnswer:    a.UserAnswer,
					CorrectAnswer: q.CorrectAnswer,
					Explanation:   q.Explanation,
					Timestamp:     attempt.StartedAt,
				})
			}
		}

		quizErrors = append(quizErrors, types.QuizErrors{
			Quiz:         quiz,
			WrongAnswers: wrongAnswers,
		})
	}

	return &types.ErrorsPageData{
		User:       &user,
		QuizErrors: quizErrors,
		Stats:      stats,
	}, nil
}

func (s *QuizSessionService) GetLeaderboardData(ctx context.Context, userID uuid.UUID) (*types.LeaderboardPageData, error) {
	user, err := s.pool.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	entries, err := s.gamification.GetLeaderboard(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}

	return &types.LeaderboardPageData{
		User:    &user,
		Entries: entries,
	}, nil
}

func (s *QuizSessionService) GetUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, fmt.Errorf("get cookie: %w", err)
	}
	claims, err := s.authService.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate token: %w", err)
	}
	return claims.UserID, nil
}

func (s *QuizSessionService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	return s.gamification.GetUserStats(ctx, userID)
}
