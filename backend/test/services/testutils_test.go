package services_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
	mocks "github.com/goquizvibe/backend/shared/mocks/services"
	"github.com/goquizvibe/backend/shared/models"
	"go.uber.org/mock/gomock"
)

type MockTimeProvider struct {
	now time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.now
}

func NewMockTimeProvider(t time.Time) *MockTimeProvider {
	return &MockTimeProvider{now: t}
}

func NewMockTimeProviderWithNow() *MockTimeProvider {
	return &MockTimeProvider{now: time.Now()}
}

type MockAuthenticator struct {
	ctrl     *gomock.Controller
	recorder *MockAuthenticatorRecorder
}

type MockAuthenticatorRecorder struct {
	mock *MockAuthenticator
}

func NewMockAuthenticator(ctrl *gomock.Controller) *MockAuthenticator {
	m := &MockAuthenticator{ctrl: ctrl}
	m.recorder = &MockAuthenticatorRecorder{mock: m}
	return m
}

func (m *MockAuthenticator) EXPECT() *MockAuthenticatorRecorder {
	return m.recorder
}

func (m *MockAuthenticator) ValidateToken(token string) (*models.AuthClaims, error) {
	ret := m.ctrl.Call(m, "ValidateToken", token)
	ret0, _ := ret[0].(*models.AuthClaims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAuthenticatorRecorder) ValidateToken(token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockAuthenticator)(nil).ValidateToken), token)
}

type MockCacheService struct {
	ctrl     *gomock.Controller
	recorder *MockCacheServiceRecorder
}

type MockCacheServiceRecorder struct {
	mock *MockCacheService
}

func NewMockCacheService(ctrl *gomock.Controller) *MockCacheService {
	m := &MockCacheService{ctrl: ctrl}
	m.recorder = &MockCacheServiceRecorder{mock: m}
	return m
}

func (m *MockCacheService) EXPECT() *MockCacheServiceRecorder {
	return m.recorder
}

func (m *MockCacheService) Get(ctx context.Context, key string, dest any) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key, dest)
	return ret[0].(bool)
}

func (mr *MockCacheServiceRecorder) Get(ctx, key, dest any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheService)(nil).Get), ctx, key, dest)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value)
	return ret[0].(error)
}

func (mr *MockCacheServiceRecorder) Set(ctx, key, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheService)(nil).Set), ctx, key, value)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, key)
	return ret[0].(error)
}

func (mr *MockCacheServiceRecorder) Delete(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockCacheService)(nil).Delete), ctx, key)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", ctx, key)
	return ret[0].(bool), ret[1].(error)
}

func (mr *MockCacheServiceRecorder) Exists(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockCacheService)(nil).Exists), ctx, key)
}

func NewTimeProvider(t time.Time) timeprovider.TimeProvider {
	return &MockTimeProvider{now: t}
}

type UserBuilder struct {
	user db.User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: db.User{
			ID:    uuid.New(),
			Name:  "Test User",
			Email: "test@example.com",
			Role:  db.RoleStudent,
		},
	}
}

func (b *UserBuilder) ID(id uuid.UUID) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) Email(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) Name(name string) *UserBuilder {
	b.user.Name = name
	return b
}

func (b *UserBuilder) Role(role db.Role) *UserBuilder {
	b.user.Role = role
	return b
}

func (b *UserBuilder) PasswordHash(hash string) *UserBuilder {
	b.user.PasswordHash = hash
	return b
}

func (b *UserBuilder) Build() db.User {
	return b.user
}

type QuizBuilder struct {
	quiz db.Quiz
}

func NewQuizBuilder() *QuizBuilder {
	return &QuizBuilder{
		quiz: db.Quiz{
			ID:               uuid.New(),
			Title:            "Test Quiz",
			Description:      "Test Description",
			Subject:          "Math",
			Grade:            5,
			TimeLimit:        300,
			QuestionPoolSize: 0,
			Status:           db.QuizStatusAvailable,
		},
	}
}

func (b *QuizBuilder) ID(id uuid.UUID) *QuizBuilder {
	b.quiz.ID = id
	return b
}

func (b *QuizBuilder) Title(title string) *QuizBuilder {
	b.quiz.Title = title
	return b
}

func (b *QuizBuilder) Description(desc string) *QuizBuilder {
	b.quiz.Description = desc
	return b
}

func (b *QuizBuilder) Subject(subject string) *QuizBuilder {
	b.quiz.Subject = subject
	return b
}

func (b *QuizBuilder) Grade(grade int) *QuizBuilder {
	b.quiz.Grade = grade
	return b
}

func (b *QuizBuilder) TimeLimit(seconds int) *QuizBuilder {
	b.quiz.TimeLimit = seconds
	return b
}

func (b *QuizBuilder) PoolSize(size int) *QuizBuilder {
	b.quiz.QuestionPoolSize = size
	return b
}

func (b *QuizBuilder) Status(status db.QuizStatus) *QuizBuilder {
	b.quiz.Status = status
	return b
}

func (b *QuizBuilder) CreatedBy(userID uuid.UUID) *QuizBuilder {
	b.quiz.CreatedBy = userID
	return b
}

func (b *QuizBuilder) Build() db.Quiz {
	return b.quiz
}

func (b *QuizBuilder) BuildWithQuestions(questions []models.QuestionWithImages) *models.QuizWithQuestionsAndImages {
	return &models.QuizWithQuestionsAndImages{
		Quiz:      b.quiz,
		Questions: questions,
	}
}

type QuestionBuilder struct {
	question db.Question
	options  []string
	images   []db.QuestionImage
}

func NewQuestionBuilder() *QuestionBuilder {
	return &QuestionBuilder{
		question: db.Question{
			ID:            uuid.New(),
			Text:          "Test Question",
			Type:          db.QuestionTypeChoice,
			CorrectAnswer: "A",
			Points:        10,
		},
		options: []string{"A", "B", "C", "D"},
	}
}

func (b *QuestionBuilder) ID(id uuid.UUID) *QuestionBuilder {
	b.question.ID = id
	return b
}

func (b *QuestionBuilder) QuizID(quizID uuid.UUID) *QuestionBuilder {
	b.question.QuizID = quizID
	return b
}

func (b *QuestionBuilder) Text(text string) *QuestionBuilder {
	b.question.Text = text
	return b
}

func (b *QuestionBuilder) Type(qType db.QuestionType) *QuestionBuilder {
	b.question.Type = qType
	return b
}

func (b *QuestionBuilder) Options(opts []string) *QuestionBuilder {
	b.options = opts
	optsJSON, _ := json.Marshal(opts)
	b.question.Options = optsJSON
	return b
}

func (b *QuestionBuilder) CorrectAnswer(answer string) *QuestionBuilder {
	b.question.CorrectAnswer = answer
	return b
}

func (b *QuestionBuilder) Explanation(exp string) *QuestionBuilder {
	b.question.Explanation = exp
	return b
}

func (b *QuestionBuilder) Points(points int) *QuestionBuilder {
	b.question.Points = points
	return b
}

func (b *QuestionBuilder) Image(url string) *QuestionBuilder {
	b.images = append(b.images, db.QuestionImage{
		ID:         uuid.New(),
		QuestionID: b.question.ID,
		Url:        url,
	})
	return b
}

func (b *QuestionBuilder) Build() db.Question {
	return b.question
}

func (b *QuestionBuilder) WithImages() models.QuestionWithImages {
	return models.QuestionWithImages{
		Question: b.question,
		Images:   b.images,
	}
}

type AttemptBuilder struct {
	attempt db.QuizAttempt
}

func NewAttemptBuilder() *AttemptBuilder {
	return &AttemptBuilder{
		attempt: db.QuizAttempt{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			QuizID:   uuid.New(),
			Score:    0,
			MaxScore: 100,
		},
	}
}

func (b *AttemptBuilder) ID(id uuid.UUID) *AttemptBuilder {
	b.attempt.ID = id
	return b
}

func (b *AttemptBuilder) UserID(userID uuid.UUID) *AttemptBuilder {
	b.attempt.UserID = userID
	return b
}

func (b *AttemptBuilder) QuizID(quizID uuid.UUID) *AttemptBuilder {
	b.attempt.QuizID = quizID
	return b
}

func (b *AttemptBuilder) Score(score int) *AttemptBuilder {
	b.attempt.Score = score
	return b
}

func (b *AttemptBuilder) MaxScore(maxScore int) *AttemptBuilder {
	b.attempt.MaxScore = maxScore
	return b
}

func (b *AttemptBuilder) StartedAt(t time.Time) *AttemptBuilder {
	b.attempt.StartedAt = t
	return b
}

func (b *AttemptBuilder) CompletedAt(t time.Time) *AttemptBuilder {
	b.attempt.CompletedAt = sql.NullTime{Time: t, Valid: true}
	return b
}

func (b *AttemptBuilder) NotCompleted() *AttemptBuilder {
	b.attempt.CompletedAt = sql.NullTime{}
	return b
}

func (b *AttemptBuilder) Build() db.QuizAttempt {
	return b.attempt
}

func MustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

func NewRequest(method, urlStr string) *http.Request {
	return httptest.NewRequest(method, urlStr, nil)
}

func NewRequestWithToken(method, urlStr, token string) *http.Request {
	req := httptest.NewRequest(method, urlStr, nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}
	return req
}

func NewRequestWithCookie(method, urlStr string, cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(method, urlStr, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	return req
}

func ExpectQuizSetup(mockQuizzes *mocks.MockQuizRepository, mockQuestions *mocks.MockQuestionRepository, mockImages *mocks.MockImageRepository, quiz *models.QuizWithQuestionsAndImages) {
	mockQuizzes.EXPECT().GetQuizByID(gomock.Any(), quiz.Quiz.ID).Return(quiz.Quiz, nil).AnyTimes()
	if len(quiz.Questions) > 0 {
		q := quiz.Questions[0]
		mockQuestions.EXPECT().GetQuestionsByQuizID(gomock.Any(), quiz.Quiz.ID).Return([]db.Question{q.Question}, nil).AnyTimes()
		mockImages.EXPECT().GetImagesByQuestionID(gomock.Any(), q.ID).Return(q.Images, nil).AnyTimes()
	}
}

type SessionBuilder struct {
	session db.QuizSession
}

func NewSessionBuilder() *SessionBuilder {
	return &SessionBuilder{
		session: db.QuizSession{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			QuizID:       uuid.New(),
			AttemptID:    uuid.New(),
			CurrentIndex: 0,
			Answers:      nil,
			CreatedAt:    time.Now(),
		},
	}
}

func (b *SessionBuilder) ID(id uuid.UUID) *SessionBuilder {
	b.session.ID = id
	return b
}

func (b *SessionBuilder) UserID(userID uuid.UUID) *SessionBuilder {
	b.session.UserID = userID
	return b
}

func (b *SessionBuilder) QuizID(quizID uuid.UUID) *SessionBuilder {
	b.session.QuizID = quizID
	return b
}

func (b *SessionBuilder) AttemptID(attemptID uuid.UUID) *SessionBuilder {
	b.session.AttemptID = attemptID
	return b
}

func (b *SessionBuilder) CurrentIndex(index int) *SessionBuilder {
	b.session.CurrentIndex = index
	return b
}

func (b *SessionBuilder) CreatedAt(t time.Time) *SessionBuilder {
	b.session.CreatedAt = t
	return b
}

func (b *SessionBuilder) Build() db.QuizSession {
	return b.session
}

type StatsBuilder struct {
	row db.GetUserStatsRow
}

func NewStatsBuilder() *StatsBuilder {
	return &StatsBuilder{
		row: db.GetUserStatsRow{
			TotalXp:    0,
			CorrectCnt: 0,
			WrongCnt:   0,
		},
	}
}

func (b *StatsBuilder) XP(xp int64) *StatsBuilder {
	b.row.TotalXp = xp
	return b
}

func (b *StatsBuilder) Correct(count int64) *StatsBuilder {
	b.row.CorrectCnt = count
	return b
}

func (b *StatsBuilder) Wrong(count int64) *StatsBuilder {
	b.row.WrongCnt = count
	return b
}

func (b *StatsBuilder) Build() db.GetUserStatsRow {
	return b.row
}
