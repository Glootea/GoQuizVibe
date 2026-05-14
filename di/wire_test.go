package di

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/config"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/locales"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/services"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goforj/wire"
)

type TestApp struct {
	AuthService         *services.AuthService
	QuizService         *services.QuizService
	QuizSessionService  *services.QuizSessionService
	AdminService        *services.AdminService
	DashboardService    *services.DashboardService
	GamificationService *services.GamificationService
	StorageService      *services.StorageService
	CacheService        *services.CacheService

	AuthHandler      *handlers.AuthHandler
	DashboardHandler *handlers.DashboardHandler
	QuizHandler      *handlers.QuizHandler
	AdminHandler     *handlers.AdminHandler

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

type MockGamificationService struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*models.UserStats
}

func NewMockGamificationService() *services.GamificationService {
	return nil
}

func ProvideTestConfig() *config.Config {
	cfg := config.Load()
	return cfg
}

func ProvideTestCacheService() *services.CacheService {
	return nil
}

func ProvideTestStorageService() *services.StorageService {
	return nil
}

func ProvideTestAuthService(queries *db.Queries, cfg *config.Config) *services.AuthService {
	return services.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
}

func ProvideTestGamificationService(queries *db.Queries, tp services.TimeProvider) *services.GamificationService {
	return services.NewGamificationService(queries, queries, tp)
}

func ProvideTestQuizService(queries *db.Queries) *services.QuizService {
	return services.NewQuizService(queries, queries, queries, queries)
}

func ProvideTestAdminService(
	queries *db.Queries,
	authService *services.AuthService,
	storageService *services.StorageService,
) *services.AdminService {
	return services.NewAdminService(queries, queries, queries, queries, queries, queries, authService, storageService)
}

func ProvideTestQuizSessionService(
	queries *db.Queries,
	gamification *services.GamificationService,
) *services.QuizSessionService {
	return nil
}

func ProvideTestDashboardService(
	queries *db.Queries,
	gamification *services.GamificationService,
	authService *services.AuthService,
	sessionService *services.QuizSessionService,
) *services.DashboardService {
	return services.NewDashboardService(queries, queries, queries, queries, gamification, authService, sessionService)
}

func ProvideMockAuthenticator() services.Authenticator {
	return NewMockAuthenticator()
}

func ProvideMockTimeProvider() services.TimeProvider {
	return NewMockTimeProvider()
}

func ProvideTestAuthHandler(queries *db.Queries, authService *services.AuthService, localeService *locales.Service) *handlers.AuthHandler {
	return handlers.NewAuth(queries, authService, localeService)
}

func ProvideTestDashboardHandler(dashboardService *services.DashboardService) *handlers.DashboardHandler {
	return handlers.NewDashboard(dashboardService)
}

func ProvideTestQuizHandler(
	queries *db.Queries,
	quizService *services.QuizService,
	authService *services.AuthService,
) *handlers.QuizHandler {
	return handlers.NewQuiz(queries, quizService, nil, authService)
}

func ProvideTestAdminHandler(
	authService *services.AuthService,
	localeService *locales.Service,
) *handlers.AdminHandler {
	return handlers.NewAdmin(nil, authService, localeService)
}

func ProvideTestRequireAuthMiddleware(authService *services.AuthService) middleware.RequireAuthMiddleware {
	return middleware.NewRequireAuthMiddleware(authService)
}

func ProvideTestRequireRoleMiddleware(authService *services.AuthService) middleware.RequireRoleMiddleware {
	return middleware.NewRequireRoleMiddleware(authService, models.RoleTeacher)
}

func ProvideTestCompressionMiddleware() *middleware.CompressionMiddleware {
	return middleware.NewCompressionMiddleware()
}

func ProvideTestCommonHeadersMiddleware() *middleware.CommonHeaders {
	return middleware.NewCommonHeaders()
}

func ProvideTestLocaleMiddleware(localeService *locales.Service) *middleware.LocaleMiddleware {
	return middleware.NewLocaleMiddleware(localeService)
}

var TestProviderSet = wire.NewSet(
	ProvideTestConfig,
	ProvideDBPool,
	ProvideDB,
	ProvideRedis,
	ProvideTestCacheService,
	ProvideTestAuthService,
	ProvideTestGamificationService,
	ProvideTestQuizService,
	ProvideTestStorageService,
	ProvideTestAdminService,
	ProvideTestQuizSessionService,
	ProvideTestDashboardService,
	ProvideLocaleService,
	ProvideMockAuthenticator,
	ProvideMockTimeProvider,
	wire.Bind(new(services.Authenticator), new(*MockAuthenticator)),
	wire.Bind(new(services.TimeProvider), new(*MockTimeProvider)),
)

var TestHandlerSet = wire.NewSet(
	ProvideTestAuthHandler,
	ProvideTestDashboardHandler,
	ProvideTestQuizHandler,
	ProvideTestAdminHandler,
)

var TestMiddlewareSet = wire.NewSet(
	ProvideTestRequireAuthMiddleware,
	ProvideTestRequireRoleMiddleware,
	ProvideTestCompressionMiddleware,
	ProvideTestCommonHeadersMiddleware,
	ProvideTestLocaleMiddleware,
)

var TestAppSet = wire.NewSet(
	TestProviderSet,
	TestHandlerSet,
	TestMiddlewareSet,
	wire.Struct(new(TestApp), "AuthService", "QuizService", "QuizSessionService", "AdminService", "DashboardService", "GamificationService", "StorageService", "CacheService", "AuthHandler", "DashboardHandler", "QuizHandler", "AdminHandler", "RequireAuthMiddleware", "RequireRoleMiddleware", "CompressionMiddleware", "CommonHeadersMiddleware", "LocaleMiddleware", "LocaleService", "MockAuthenticator", "MockTimeProvider"),
)

func InitializeTestApp(ctx context.Context) (*TestApp, error) {
	wire.Build(TestAppSet)
	return nil, nil
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
	authService := services.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
	gamification := services.NewGamificationService(queries, queries, mockTime)
	quizService := services.NewQuizService(queries, queries, queries, queries)

	quizSessionService := services.NewQuizSessionService(queries, queries, queries, queries, queries, *gamification, services.CacheService{})
	dashboardService := services.NewDashboardService(queries, queries, queries, queries, gamification, authService, quizSessionService)

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
		AuthHandler:             handlers.NewAuth(queries, authService, localeSvc),
		DashboardHandler:        handlers.NewDashboard(dashboardService),
		QuizHandler:             handlers.NewQuiz(queries, quizService, nil, authService),
		AdminHandler:            handlers.NewAdmin(nil, authService, localeSvc),
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
