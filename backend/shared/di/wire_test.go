package di

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	adminHndl "github.com/goquizvibe/backend/feature/admin/handlers"
	adminSvc "github.com/goquizvibe/backend/feature/admin/services"
	authHndl "github.com/goquizvibe/backend/feature/auth/handlers"
	authSvc "github.com/goquizvibe/backend/feature/auth/services"
	dashboardHndl "github.com/goquizvibe/backend/feature/dashboard/handlers"
	dashboardSvc "github.com/goquizvibe/backend/feature/dashboard/services"
	gamificationSvc "github.com/goquizvibe/backend/feature/gamification/services"
	quizHndl "github.com/goquizvibe/backend/feature/quiz/handlers"
	quizSvc "github.com/goquizvibe/backend/feature/quiz/services"
	"github.com/goquizvibe/backend/shared/config"
	"github.com/goquizvibe/backend/shared/db"
	cacheSvc "github.com/goquizvibe/backend/shared/infrastructure/cache"
	storageSvc "github.com/goquizvibe/backend/shared/infrastructure/storage"
	locales "github.com/goquizvibe/backend/shared/locales"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestApp struct {
	AuthService         *authSvc.AuthService
	QuizService         *quizSvc.QuizService
	QuizSessionService  *quizSvc.QuizSessionService
	AdminService        *adminSvc.AdminService
	DashboardService    *dashboardSvc.DashboardService
	GamificationService *gamificationSvc.GamificationService
	StorageService      *storageSvc.StorageService
	CacheService        *cacheSvc.CacheService

	AuthHandler      *authHndl.AuthHandler
	DashboardHandler *dashboardHndl.DashboardHandler
	QuizHandler      *quizHndl.QuizHandler
	AdminHandler     *adminHndl.AdminHandler

	RequireAuthMiddleware   middleware.RequireAuthMiddleware
	RequireRoleMiddleware   middleware.RequireRoleMiddleware
	CompressionMiddleware   *middleware.CompressionMiddleware
	CommonHeadersMiddleware *middleware.CommonHeaders
	LocaleMiddleware        *middleware.LocaleMiddleware

	LocaleService *locales.Service

	MockAuthenticator *MockAuthenticator
	MockTimeProvider  *MockTimeProvider
}

type MockAuthenticator struct {
	mu     sync.RWMutex
	tokens map[string]*models.AuthClaims
}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{
		tokens: make(map[string]*models.AuthClaims),
	}
}

func (m *MockAuthenticator) ValidateToken(token string) (*models.AuthClaims, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if claims, ok := m.tokens[token]; ok {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

func (m *MockAuthenticator) AddToken(token string, claims *models.AuthClaims) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[token] = claims
}

var ErrInvalidToken = &mockError{msg: "invalid token"}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

type MockTimeProvider struct {
	mu  sync.RWMutex
	now time.Time
}

func NewMockTimeProvider() *MockTimeProvider {
	return &MockTimeProvider{
		now: time.Now(),
	}
}

func (m *MockTimeProvider) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.now
}

func (m *MockTimeProvider) SetNow(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = t
}

type MockCacheService struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		data: make(map[string][]byte),
	}
}

func (m *MockCacheService) Get(ctx context.Context, key string, dest any) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.data[key]
	if !ok {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

func (m *MockCacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = data
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

func CreateTestApp(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) *TestApp {
	queries := db.New(pool)

	mockAuth := NewMockAuthenticator()
	mockAuth.AddToken("test-token", &models.AuthClaims{
		UserID: uuid.New(),
		Email:  "test@example.com",
		Role:   models.RoleStudent,
	})

	mockTime := NewMockTimeProvider()
	authService := authSvc.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
	gamification := gamificationSvc.NewGamificationService(queries, queries, mockTime)
	cacheService := cacheSvc.NewCacheService(nil, cfg.Redis.CacheTTL)
	quizService := quizSvc.NewQuizService(queries, queries, queries, queries, queries, cacheService)
	schemaService := adminSvc.NewQuestionSchema(nil)
	promptGenerator := adminSvc.NewPromptGenerator(schemaService)

	quizSessionService := quizSvc.NewQuizSessionService(queries, queries, queries, queries, queries, gamification, cacheService)
	dashboardService := dashboardSvc.NewDashboardService(queries, queries, queries, queries, queries, gamification, authService, quizSessionService, cacheService)

	localeSvc, _ := locales.NewService()

	return &TestApp{
		AuthService:             authService,
		QuizService:             quizService,
		QuizSessionService:      quizSessionService,
		AdminService:            nil,
		DashboardService:        dashboardService,
		GamificationService:     gamification,
		StorageService:          nil,
		CacheService:            nil,
		AuthHandler:             authHndl.NewAuth(queries, authService, localeSvc),
		DashboardHandler:        dashboardHndl.NewDashboard(dashboardService),
		QuizHandler:             quizHndl.NewQuiz(queries, quizService, nil, authService),
		AdminHandler:            adminHndl.NewAdmin(nil, authService, localeSvc, promptGenerator),
		RequireAuthMiddleware:   middleware.NewRequireAuthMiddleware(authService),
		RequireRoleMiddleware:   middleware.NewRequireRoleMiddleware(authService, models.RoleTeacher),
		CompressionMiddleware:   middleware.NewCompressionMiddleware(),
		CommonHeadersMiddleware: middleware.NewCommonHeaders(),
		LocaleMiddleware:        middleware.NewLocaleMiddleware(localeSvc),
		LocaleService:           localeSvc,
		MockAuthenticator:       mockAuth,
		MockTimeProvider:        mockTime,
	}
}
