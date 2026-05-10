package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
	"github.com/goquizvibe/types"
)

type QuizSessionService struct {
	attempts     r.AttemptRepository
	sessions     r.SessionRepository
	quizzes      r.QuizRepository
	questions    r.QuestionRepository
	images       r.ImageRepository
	users        r.UserRepository
	gamification *GamificationService
	cache        *CacheService
}

func NewQuizSessionService(
	attempts r.AttemptRepository,
	sessions r.SessionRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	users r.UserRepository,
	gamification *GamificationService,
	cache *CacheService,
) *QuizSessionService {
	return &QuizSessionService{
		attempts:     attempts,
		sessions:     sessions,
		quizzes:      quizzes,
		questions:    questions,
		images:       images,
		users:        users,
		gamification: gamification,
		cache:        cache,
	}
}

func cacheKeyQuestions(sessionID uuid.UUID) string {
	return fmt.Sprintf("quiz:session:%s:questions", sessionID.String())
}

func cacheKeyOrder(sessionID uuid.UUID) string {
	return fmt.Sprintf("quiz:session:%s:order", sessionID.String())
}

func cacheKeyAnswers(sessionID uuid.UUID) string {
	return fmt.Sprintf("quiz:session:%s:answers", sessionID.String())
}

func (s *QuizSessionService) CreateSession(ctx context.Context, userID, quizID uuid.UUID) (*db.QuizSession, error) {
	attemptID := uuid.New()
	now := time.Now()

	_, err := s.attempts.CreateAttempt(ctx, db.CreateAttemptParams{
		ID:        attemptID,
		UserID:    userID,
		QuizID:    quizID,
		StartedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create attempt: %w", err)
	}

	quiz, err := s.getQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	var selectedQuestions []models.QuestionWithImages
	var order []int

	if quiz.QuestionPoolSize > 0 && len(quiz.Questions) > int(quiz.QuestionPoolSize) {
		poolSize := int(quiz.QuestionPoolSize)
		indices := rand.Perm(len(quiz.Questions))[:poolSize]
		order = indices
		selectedQuestions = make([]models.QuestionWithImages, poolSize)
		for i, idx := range indices {
			selectedQuestions[i] = quiz.Questions[idx]
		}
	} else {
		order = make([]int, len(quiz.Questions))
		for i := range quiz.Questions {
			order[i] = i
		}
		selectedQuestions = quiz.Questions
	}

	ttl := time.Duration(quiz.TimeLimit+60) * time.Second

	session := &db.QuizSession{
		ID:           uuid.New(),
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attemptID,
		CurrentIndex: 0,
		Answers:      nil,
		CreatedAt:    now,
	}

	_, err = s.sessions.CreateSession(ctx, db.CreateSessionParams{
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

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKeyQuestions(session.ID), selectedQuestions, ttl)
		_ = s.cache.Set(ctx, cacheKeyOrder(session.ID), order, ttl)
		_ = s.cache.Set(ctx, cacheKeyAnswers(session.ID), map[int]string{}, ttl)
	}

	return session, nil
}

type QuestionFeedback struct {
	IsCorrect     bool
	CorrectAnswer string
	Explanation   string
	IsLast        bool
}

func (s *QuizSessionService) GetSessionQuestions(ctx context.Context, sessionID uuid.UUID) ([]models.QuestionWithImages, error) {
	if s.cache != nil {
		var questions []models.QuestionWithImages
		if s.cache.Get(ctx, cacheKeyQuestions(sessionID), &questions) {
			return questions, nil
		}
	}

	questions, err := s.loadQuestionsFromDB(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		ttl := time.Minute
		_ = s.cache.Set(ctx, cacheKeyQuestions(sessionID), questions, ttl)
	}

	return questions, nil
}

func (s *QuizSessionService) loadQuestionsFromDB(ctx context.Context, sessionID uuid.UUID) ([]models.QuestionWithImages, error) {
	session, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	quiz, err := s.getQuizWithQuestions(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		var order []int
		if s.cache.Get(ctx, cacheKeyOrder(sessionID), &order) {
			selected := make([]models.QuestionWithImages, len(order))
			for i, idx := range order {
				selected[i] = quiz.Questions[idx]
			}
			return selected, nil
		}
	}

	return quiz.Questions, nil
}

func (s *QuizSessionService) GetAnswers(ctx context.Context, sessionID uuid.UUID) (map[int]string, error) {
	if s.cache != nil {
		var answers map[int]string
		if s.cache.Get(ctx, cacheKeyAnswers(sessionID), &answers) {
			return answers, nil
		}
	}
	return map[int]string{}, nil
}

func (s *QuizSessionService) NavigateQuestion(ctx context.Context, sessionID uuid.UUID, quizID uuid.UUID, currentIndex, targetIndex int, answer string, user *db.User) (*types.QuizPageData, error) {
	session, questions, err := s.getSessionWithQuestions(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session with questions: %w", err)
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("no questions in session")
	}

	if targetIndex < 0 {
		targetIndex = len(questions) - 1
	}
	if targetIndex >= len(questions) {
		targetIndex = 0
	}

	answers := make(map[int]string)
	if session.Answers != nil {
		if err := json.Unmarshal(session.Answers, &answers); err != nil {
			return nil, fmt.Errorf("unmarshal answers: %w", err)
		}
	}
	answers[currentIndex] = answer

	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return nil, fmt.Errorf("marshal answers: %w", err)
	}

	_, err = s.sessions.UpdateSession(ctx, db.UpdateSessionParams{
		ID:           session.ID,
		CurrentIndex: targetIndex,
		Answers:      answersJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	question := questions[currentIndex]
	isCorrect := NormalizeAnswer(answer) == NormalizeAnswer(question.CorrectAnswer)

	_, err = s.attempts.CreateUserAnswer(ctx, db.CreateUserAnswerParams{
		ID:         uuid.New(),
		AttemptID:  session.AttemptID,
		QuestionID: question.ID,
		UserAnswer: answer,
		IsCorrect:  isCorrect,
	})
	if err != nil {
		return nil, fmt.Errorf("create user answer: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKeyAnswers(sessionID), answers, time.Minute)
	}

	quiz, err := s.getQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	attempt, err := s.attempts.GetAttemptByID(ctx, session.AttemptID)
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}

	remainingSeconds := max(int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit)*time.Second)).Seconds()), 0)

	newIndex := targetIndex
	isLast := newIndex >= len(questions)-1

	return &types.QuizPageData{
		User:             user,
		Questions:        questions,
		SessionID:        sessionID.String(),
		CurrentIndex:     newIndex,
		Answers:          answers,
		TotalQuestions:   len(questions),
		RemainingSeconds: remainingSeconds,
		TimeLimitMinutes: quiz.TimeLimit / 60,
		IsLastQuestion:   isLast,
	}, nil
}

func (s *QuizSessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) (*db.QuizAttempt, error) {
	session, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	quiz, err := s.getQuizWithQuestions(ctx, session.QuizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	answers, err := s.attempts.GetAnswersByAttempt(ctx, session.AttemptID)
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
	updatedAttempt, err := s.attempts.UpdateAttempt(ctx, db.UpdateAttemptParams{
		ID:          session.AttemptID,
		Score:       score,
		MaxScore:    maxScore,
		CompletedAt: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update attempt: %w", err)
	}

	err = s.sessions.DeleteSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("delete session: %w", err)
	}

	return &updatedAttempt, nil
}

func (s *QuizSessionService) GetQuizResultData(ctx context.Context, quizID, sessionID, userID uuid.UUID) (*types.QuizResultData, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quiz, err := s.getQuizWithQuestions(ctx, quizID)
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

	answers, err := s.attempts.GetAnswersByAttempt(ctx, attempt.ID)
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
			Explanation:   q.Explanation,
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
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	attempts, err := s.attempts.GetQuizErrors(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get quiz errors: %w", err)
	}

	stats, err := s.gamification.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	var quizErrors []types.QuizErrors
	for _, attempt := range attempts {
		quiz, err := s.getQuizWithQuestions(ctx, attempt.QuizID)
		if err != nil {
			continue
		}

		answers, err := s.attempts.GetAnswersByAttempt(ctx, attempt.ID)
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
	user, err := s.users.GetUserByID(ctx, userID)
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

func (s *QuizSessionService) GetUserIDFromRequest(r *http.Request, auth Authenticator) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, fmt.Errorf("get cookie: %w", err)
	}
	claims, err := auth.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate token: %w", err)
	}
	return claims.UserID, nil
}

func (s *QuizSessionService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	return s.gamification.GetUserStats(ctx, userID)
}

func (s *QuizSessionService) GetQuizQuestionData(ctx context.Context, sessionID uuid.UUID, quizID uuid.UUID, index int) (*types.QuizPageData, error) {
	session, questions, err := s.getSessionWithQuestions(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session with questions: %w", err)
	}

	if index >= len(questions) {
		index = 0
	}

	attempt, err := s.attempts.GetAttemptByID(ctx, session.AttemptID)
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}

	quiz, err := s.getQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	answers, err := s.GetAnswers(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get answers: %w", err)
	}

	remainingSeconds := int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit) * time.Second)).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	isLast := index >= len(questions)-1

	return &types.QuizPageData{
		Questions:        questions,
		SessionID:        sessionID.String(),
		CurrentIndex:     index,
		Answers:          answers,
		TotalQuestions:   len(questions),
		RemainingSeconds: remainingSeconds,
		TimeLimitMinutes: quiz.TimeLimit / 60,
		IsLastQuestion:   isLast,
	}, nil
}

func (s *QuizSessionService) getQuizWithQuestions(ctx context.Context, quizID uuid.UUID) (*models.QuizWithQuestionsAndImages, error) {
	quiz, err := s.quizzes.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	questions, err := s.questions.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	questionsWithImages := s.attachImagesToQuestions(ctx, questions)
	return &models.QuizWithQuestionsAndImages{
		Quiz:      quiz,
		Questions: questionsWithImages,
	}, nil
}

func (s *QuizSessionService) attachImagesToQuestions(ctx context.Context, questions []db.Question) []models.QuestionWithImages {
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

func (s *QuizSessionService) SessionExists(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return s.sessions.SessionExists(ctx, sessionID)
}

func (s *QuizSessionService) GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*types.ActiveSessionInfo, error) {
	attempts, err := s.attempts.GetIncompleteAttemptsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get incomplete attempts: %w", err)
	}

	if len(attempts) == 0 {
		return nil, nil
	}

	for _, attempt := range attempts {
		session, err := s.sessions.GetSessionByAttemptID(ctx, attempt.ID)
		if err != nil {
			continue
		}

		quiz, err := s.quizzes.GetQuizByID(ctx, attempt.QuizID)
		if err != nil {
			continue
		}

		remainingSeconds := int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit) * time.Second)).Seconds())
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}

		return &types.ActiveSessionInfo{
			SessionID:        session.ID,
			QuizID:           quiz.ID,
			QuizTitle:        quiz.Title,
			CurrentIndex:     session.CurrentIndex,
			RemainingSeconds: remainingSeconds,
		}, nil
	}

	return nil, nil
}

func (s *QuizSessionService) getSessionWithQuestions(ctx context.Context, sessionID uuid.UUID) (db.QuizSession, []models.QuestionWithImages, error) {
	session, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil {
		return db.QuizSession{}, nil, err
	}

	quiz, err := s.getQuizWithQuestions(ctx, session.QuizID)
	if err != nil {
		return db.QuizSession{}, nil, err
	}

	return session, quiz.Questions, nil
}
