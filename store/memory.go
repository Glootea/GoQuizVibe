package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/models"
)

var (
	ErrEmailExists  = fmt.Errorf("email already registered")
	ErrUserNotFound = fmt.Errorf("user not found")
	ErrQuizNotFound = fmt.Errorf("quiz not found")
)

type MemoryStore struct {
	mu               sync.RWMutex
	Users            map[uuid.UUID]*models.User
	Quizzes          map[uuid.UUID]*models.Quiz
	Progress         map[uuid.UUID]*models.UserProgress
	Attempts         map[uuid.UUID]*models.QuizAttempt
	EmailIndex       map[string]uuid.UUID
	Sessions         map[string]*QuizSession
	CompletedResults map[string]*models.QuizAttempt
}

type QuizSession struct {
	AttemptID    uuid.UUID
	QuizID       uuid.UUID
	UserID       uuid.UUID
	CurrentIndex int
	Answers      map[int]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Users:            make(map[uuid.UUID]*models.User),
		Quizzes:          make(map[uuid.UUID]*models.Quiz),
		Progress:         make(map[uuid.UUID]*models.UserProgress),
		Attempts:         make(map[uuid.UUID]*models.QuizAttempt),
		EmailIndex:       make(map[string]uuid.UUID),
		Sessions:         make(map[string]*QuizSession),
		CompletedResults: make(map[string]*models.QuizAttempt),
	}
}

func (s *MemoryStore) CreateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.EmailIndex[u.Email]; exists {
		return ErrEmailExists
	}

	s.Users[u.ID] = u
	s.EmailIndex[u.Email] = u.ID
	s.Progress[u.ID] = &models.UserProgress{
		UserID:           u.ID,
		XP:               0,
		Streak:           0,
		LastActiveDate:   time.Time{},
		CompletedQuizzes: []uuid.UUID{},
		WrongAnswers:     []models.WrongAnswer{},
	}
	return nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.EmailIndex[email]
	if !ok {
		return nil, ErrUserNotFound
	}

	u, ok := s.Users[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func (s *MemoryStore) GetUserByID(id uuid.UUID) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.Users[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func (s *MemoryStore) GetQuizzes() []*models.Quiz {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quizzes := make([]*models.Quiz, 0, len(s.Quizzes))
	for _, q := range s.Quizzes {
		quizzes = append(quizzes, q)
	}
	return quizzes
}

func (s *MemoryStore) GetQuizByID(id uuid.UUID) (*models.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q, ok := s.Quizzes[id]
	if !ok {
		return nil, ErrQuizNotFound
	}
	return q, nil
}

func (s *MemoryStore) GetQuizzesForUser(userID uuid.UUID) []*models.Quiz {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var quizzes []*models.Quiz
	for _, q := range s.Quizzes {
		if q.Status == models.QuizStatusAvailable {
			quizzes = append(quizzes, q)
			continue
		}
		for _, assigned := range q.AssignedTo {
			if assigned == userID {
				quizzes = append(quizzes, q)
				break
			}
		}
	}
	return quizzes
}

func (s *MemoryStore) SaveAttempt(attempt *models.QuizAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Attempts[attempt.ID] = attempt
	return nil
}

func (s *MemoryStore) GetAttemptsByUser(userID uuid.UUID) []*models.QuizAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var attempts []*models.QuizAttempt
	for _, a := range s.Attempts {
		if a.UserID == userID {
			attempts = append(attempts, a)
		}
	}
	return attempts
}

func (s *MemoryStore) GetProgress(userID uuid.UUID) (*models.UserProgress, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.Progress[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return p, nil
}

func (s *MemoryStore) UpdateProgress(p *models.UserProgress) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Progress[p.UserID] = p
	return nil
}

func (s *MemoryStore) AddWrongAnswer(userID uuid.UUID, wa models.WrongAnswer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.Progress[userID]
	if !ok {
		return ErrUserNotFound
	}

	p.WrongAnswers = append(p.WrongAnswers, wa)
	s.Progress[userID] = p
	return nil
}

func (s *MemoryStore) GetLeaderboard() []*models.LeaderboardEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*models.LeaderboardEntry, 0, len(s.Progress))
	for userID, p := range s.Progress {
		if user, ok := s.Users[userID]; ok {
			entries = append(entries, &models.LeaderboardEntry{
				UserID:   userID,
				UserName: user.Name,
				XP:       p.XP,
				Streak:   p.Streak,
				Rank:     0,
			})
		}
	}

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].XP > entries[i].XP {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	for i, e := range entries {
		e.Rank = i + 1
	}

	return entries
}

func (s *MemoryStore) CreateSession(sessionID string, userID, quizID uuid.UUID) *QuizSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Sessions == nil {
		s.Sessions = make(map[string]*QuizSession)
	}

	attempt := &models.QuizAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		QuizID:    quizID,
		Answers:   make([]models.UserAnswer, 0),
		StartedAt: time.Now(),
	}
	s.Attempts[attempt.ID] = attempt

	session := &QuizSession{
		AttemptID:    attempt.ID,
		QuizID:       quizID,
		UserID:       userID,
		CurrentIndex: 0,
		Answers:      make(map[int]string),
	}

	if sessionID == "" {
		sessionID = attempt.ID.String()
	}
	s.Sessions[sessionID] = session
	return session
}

func (s *MemoryStore) GetSession(sessionID string) (*QuizSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.Sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (s *MemoryStore) UpdateSessionAnswer(sessionID string, questionIndex int, answer string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.Sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	session.Answers[questionIndex] = answer
	session.CurrentIndex = questionIndex + 1
	return nil
}

func (s *MemoryStore) CompleteSession(sessionID string) (*models.QuizAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if attempt, ok := s.CompletedResults[sessionID]; ok {
		return attempt, nil
	}

	session, ok := s.Sessions[sessionID]
	if !ok {
		println("Session not found")
		return nil, fmt.Errorf("session not found")
	}

	attempt, ok := s.Attempts[session.AttemptID]
	if !ok {
		return nil, fmt.Errorf("attempt not found")
	}

	quiz, ok := s.Quizzes[session.QuizID]
	if !ok {
		return nil, fmt.Errorf("quiz not found")
	}

	var score, maxScore int
	for i, q := range quiz.Questions {
		maxScore += q.Points
		if userAnswer, ok := session.Answers[i]; ok {
			isCorrect := normalizeAnswer(userAnswer) == normalizeAnswer(q.CorrectAnswer)
			attempt.Answers = append(attempt.Answers, models.UserAnswer{
				QuestionID: q.ID,
				UserAnswer: userAnswer,
				IsCorrect:  isCorrect,
			})
			if isCorrect {
				score += q.Points
			}
		}
	}

	attempt.Score = score
	attempt.MaxScore = maxScore
	attempt.CompletedAt = time.Now()

	s.CompletedResults[sessionID] = attempt
	delete(s.Sessions, sessionID)
	return attempt, nil
}

func (s *MemoryStore) GetCompletedResult(sessionID string) (*models.QuizAttempt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attempt, ok := s.CompletedResults[sessionID]
	if !ok {
		return nil, fmt.Errorf("result not found")
	}
	return attempt, nil
}

func normalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.TrimRight(answer, ".")
	return answer
}
