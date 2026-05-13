package services

import (
	"context"
	"database/sql"
	"errors"
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

func init() {
	rand.NewSource(time.Now().UnixNano())
}

var ErrTimeExpired = errors.New("time expired")

type QuizSession struct {
	UserID       uuid.UUID           `json:"user_id"`
	QuizID       uuid.UUID           `json:"quiz_id"`
	AttemptID    uuid.UUID           `json:"attempt_id"`
	CurrentIndex int                 `json:"current_index"`
	Answers      map[int]AnswerState `json:"answers"`
	CreatedAt    time.Time           `json:"created_at"`
}

type AnswerState = types.AnswerState

func (s *QuizSessionService) isTimeExpired(attempt db.QuizAttempt, timeLimit int) bool {
	return time.Now().After(attempt.StartedAt.Add(time.Duration(timeLimit) * time.Second))
}

type QuizSessionService struct {
	attempts     r.AttemptRepository
	quizzes      r.QuizRepository
	questions    r.QuestionRepository
	images       r.ImageRepository
	users        r.UserRepository
	gamification GamificationService
	cache        CacheService
}

func NewQuizSessionService(
	attempts r.AttemptRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	users r.UserRepository,
	gamification GamificationService,
	cache CacheService,
) *QuizSessionService {
	return &QuizSessionService{
		attempts:     attempts,
		quizzes:      quizzes,
		questions:    questions,
		images:       images,
		users:        users,
		gamification: gamification,
		cache:        cache,
	}
}

func userSessionKey(userID uuid.UUID) string {
	return fmt.Sprintf("quiz:session:user:%s", userID.String())
}

func (s *QuizSessionService) GetSession(ctx context.Context, userID uuid.UUID) *QuizSession {
	var session QuizSession
	if s.cache.Get(ctx, userSessionKey(userID), &session) {
		return &session
	}
	return nil
}

func (s *QuizSessionService) saveSession(ctx context.Context, session *QuizSession, ttl time.Duration) error {
	return s.cache.Set(ctx, userSessionKey(session.UserID), session, ttl)
}

func (s *QuizSessionService) deleteSession(ctx context.Context, userID uuid.UUID) error {
	return s.cache.Delete(ctx, userSessionKey(userID))
}

func shuffleOptions(options []string) []string {
	shuffled := make([]string, len(options))
	copy(shuffled, options)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func (s *QuizSessionService) CreateSession(ctx context.Context, userID, quizID uuid.UUID) (*QuizSession, error) {
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

	session := &QuizSession{
		UserID:       userID,
		QuizID:       quizID,
		AttemptID:    attemptID,
		CurrentIndex: 0,
		Answers:      map[int]AnswerState{},
		CreatedAt:    now,
	}

	if err := s.saveSession(ctx, session, ttl); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	_ = s.cache.Set(ctx, cacheKeyQuestions(userID), selectedQuestions, ttl)
	_ = s.cache.Set(ctx, cacheKeyOrder(userID), order, ttl)
	_ = s.cache.Set(ctx, cacheKeyAnswers(userID), map[int]types.AnswerState{}, ttl)

	return session, nil
}

type QuestionFeedback struct {
	IsCorrect     bool
	CorrectAnswer string
	Explanation   string
	IsLast        bool
}

func (s *QuizSessionService) GetSessionQuestions(ctx context.Context, userID uuid.UUID) ([]models.QuestionWithImages, error) {
	var questions []models.QuestionWithImages
	if s.cache.Get(ctx, cacheKeyQuestions(userID), &questions) {
		return questions, nil
	}

	questions, err := s.loadQuestionsFromDB(ctx, userID)
	if err != nil {
		return nil, err
	}

	ttl := time.Minute
	_ = s.cache.Set(ctx, cacheKeyQuestions(userID), questions, ttl)

	return questions, nil
}

func (s *QuizSessionService) loadQuestionsFromDB(ctx context.Context, userID uuid.UUID) ([]models.QuestionWithImages, error) {
	session := s.GetSession(ctx, userID)
	if session == nil {
		return nil, errors.New("session not found")
	}

	quiz, err := s.getQuizWithQuestions(ctx, session.QuizID)
	if err != nil {
		return nil, err
	}

	var order []int
	if s.cache.Get(ctx, cacheKeyOrder(userID), &order) {
		selected := make([]models.QuestionWithImages, len(order))
		for i, idx := range order {
			selected[i] = quiz.Questions[idx]
		}
		return selected, nil
	}

	return quiz.Questions, nil
}

func (s *QuizSessionService) GetAnswers(ctx context.Context, userID uuid.UUID) (map[int]types.AnswerState, error) {
	var answers map[int]types.AnswerState
	if s.cache.Get(ctx, cacheKeyAnswers(userID), &answers) {
		return answers, nil
	}
	return map[int]types.AnswerState{}, nil
}

func (s *QuizSessionService) SaveAnswer(ctx context.Context, userID uuid.UUID, quizID uuid.UUID, currentIndex int, answer string) (bool, error) {
	session, questions, err := s.getSessionWithQuestions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get session with questions: %w", err)
	}

	if currentIndex < 0 || currentIndex >= len(questions) {
		return false, fmt.Errorf("invalid question index")
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
		return false, fmt.Errorf("create user answer: %w", err)
	}

	return isCorrect, nil
}

func (s *QuizSessionService) NavigateQuestion(ctx context.Context, userID uuid.UUID, quizID uuid.UUID, currentIndex, targetIndex int, answer string, user *db.User) (*types.QuizPageData, error) {
	session, questions, err := s.getSessionWithQuestions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get session with questions: %w", err)
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("no questions in session")
	}

	attempt, err := s.attempts.GetAttemptByID(ctx, session.AttemptID)
	if err != nil {
		return nil, fmt.Errorf("get attempt: %w", err)
	}

	quiz, err := s.getQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	if s.isTimeExpired(attempt, quiz.TimeLimit) {
		return nil, ErrTimeExpired
	}

	if targetIndex < 0 {
		targetIndex = len(questions) - 1
	}
	if targetIndex >= len(questions) {
		targetIndex = 0
	}

	answers := session.Answers
	if answers == nil {
		answers = map[int]AnswerState{}
	}
	answers[currentIndex] = AnswerState{Text: answer, Answered: answer != ""}
	session.Answers = answers
	session.CurrentIndex = targetIndex

	ttl := time.Duration(quiz.TimeLimit+60) * time.Second
	if err := s.saveSession(ctx, &session, ttl); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	_, err = s.SaveAnswer(ctx, userID, quizID, currentIndex, answer)
	if err != nil {
		return nil, fmt.Errorf("save answer: %w", err)
	}

	_ = s.cache.Set(ctx, cacheKeyAnswers(userID), answers, time.Minute)

	remainingSeconds := max(int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit)*time.Second)).Seconds()), 0)

	newIndex := targetIndex
	isLast := newIndex >= len(questions)-1

	answerOptions := make(map[int][]string)
	for i, q := range questions {
		if q.Type == models.QuestionTypeChoice {
			answerOptions[i] = shuffleOptions(q.GetOptions())
		}
	}

	return &types.QuizPageData{
		User:             user,
		Questions:        questions,
		SessionID:        userID.String(),
		CurrentIndex:     newIndex,
		Answers:          answers,
		TotalQuestions:   len(questions),
		RemainingSeconds: remainingSeconds,
		TimeLimitMinutes: quiz.TimeLimit / 60,
		IsLastQuestion:   isLast,
		AnswerOptions:    answerOptions,
	}, nil
}

func (s *QuizSessionService) CompleteSession(ctx context.Context, userID uuid.UUID) (*db.QuizAttempt, error) {
	session := s.GetSession(ctx, userID)
	if session == nil {
		return nil, errors.New("session not found")
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

	if err := s.deleteSession(ctx, userID); err != nil {
		return nil, fmt.Errorf("delete session: %w", err)
	}

	return &updatedAttempt, nil
}

func (s *QuizSessionService) GetQuizResultData(ctx context.Context, quizID uuid.UUID, userID uuid.UUID) (*types.QuizResultData, error) {
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

	attempt, err := s.CompleteSession(ctx, userID)
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

func (s *QuizSessionService) GetQuizQuestionData(ctx context.Context, userID uuid.UUID, quizID uuid.UUID, index int) (*types.QuizPageData, error) {
	session, questions, err := s.getSessionWithQuestions(ctx, userID)
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

	answers, err := s.GetAnswers(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get answers: %w", err)
	}

	remainingSeconds := int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit) * time.Second)).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	isLast := index >= len(questions)-1

	answerOptions := make(map[int][]string)
	for i, q := range questions {
		if q.Type == models.QuestionTypeChoice {
			answerOptions[i] = shuffleOptions(q.GetOptions())
		}
	}

	return &types.QuizPageData{
		Questions:        questions,
		SessionID:        userID.String(),
		CurrentIndex:     index,
		Answers:          answers,
		TotalQuestions:   len(questions),
		RemainingSeconds: remainingSeconds,
		TimeLimitMinutes: quiz.TimeLimit / 60,
		IsLastQuestion:   isLast,
		AnswerOptions:    answerOptions,
	}, nil
}

func (s *QuizSessionService) GetQuizInfoData(ctx context.Context, userID, quizID uuid.UUID) (*types.QuizInfoData, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	quiz, err := s.getQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("get quiz: %w", err)
	}

	return &types.QuizInfoData{
		User:          &user,
		Quiz:          quiz,
		QuestionCount: len(quiz.Questions),
		TimeLimitMin:  quiz.TimeLimit / 60,
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

func (s *QuizSessionService) SessionExists(ctx context.Context, userID uuid.UUID) bool {
	return s.GetSession(ctx, userID) != nil
}

func (s *QuizSessionService) GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) *types.ActiveSessionInfo {
	session := s.GetSession(ctx, userID)
	if session == nil {
		return nil
	}

	quiz, err := s.quizzes.GetQuizByID(ctx, session.QuizID)
	if err != nil {
		return nil
	}

	attempt, err := s.attempts.GetAttemptByID(ctx, session.AttemptID)
	if err != nil {
		return nil
	}

	remainingSeconds := max(int(time.Until(attempt.StartedAt.Add(time.Duration(quiz.TimeLimit)*time.Second)).Seconds()), 0)

	if remainingSeconds == 0 {
		return nil
	}

	return &types.ActiveSessionInfo{
		SessionID:        session.UserID,
		QuizID:           session.QuizID,
		QuizTitle:        quiz.Title,
		CurrentIndex:     session.CurrentIndex,
		RemainingSeconds: remainingSeconds,
	}
}

func (s *QuizSessionService) getSessionWithQuestions(ctx context.Context, userID uuid.UUID) (QuizSession, []models.QuestionWithImages, error) {
	session := s.GetSession(ctx, userID)
	if session == nil {
		return QuizSession{}, nil, errors.New("session not found")
	}

	quiz, err := s.getQuizWithQuestions(ctx, session.QuizID)
	if err != nil {
		return QuizSession{}, nil, err
	}

	return *session, quiz.Questions, nil
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
