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
	interfaces "github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	storageSvc "github.com/goquizvibe/backend/shared/infrastructure/storage"
	timeProvider "github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
	locales "github.com/goquizvibe/backend/shared/locales"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/models"
	r "github.com/goquizvibe/backend/shared/repositories"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goforj/wire"
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

type MockGamificationService struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*models.UserStats
}

func NewMockGamificationService() *gamificationSvc.GamificationService {
	return nil
}

func ProvideTestConfig() *config.Config {
	cfg := config.Load()
	return cfg
}

func ProvideTestCacheService() *cacheSvc.CacheService {
	return nil
}

func ProvideTestPromptGenerator(cacheService *cacheSvc.CacheService) *adminSvc.PromptGenerator {
	schemaService := adminSvc.NewQuestionSchema(cacheService)
	return adminSvc.NewPromptGenerator(schemaService)
}

func ProvideTestStorageService() *storageSvc.StorageService {
	return nil
}

func ProvideTestAuthService(queries *db.Queries, cfg *config.Config) *authSvc.AuthService {
	return authSvc.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
}

func ProvideTestGamificationService(queries *db.Queries, tp timeProvider.TimeProvider) *gamificationSvc.GamificationService {
	return gamificationSvc.NewGamificationService(queries, queries, tp)
}

func ProvideTestQuizService(queries *db.Queries, cacheService *cacheSvc.CacheService) *quizSvc.QuizService {
	return quizSvc.NewQuizService(queries, queries, queries, queries, cacheService)
}

func ProvideTestAdminService(
	users r.UserRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	attempts r.AttemptRepository,
	stats r.StatsRepository,
	materials r.LearningMaterialRepository,
	authService *authSvc.AuthService,
	storageService *storageSvc.StorageService,
	cacheService *cacheSvc.CacheService,
) *adminSvc.AdminService {
	return adminSvc.NewAdminService(users, quizzes, questions, images, attempts, stats, materials, authService, storageService, cacheService)
}

func ProvideTestQuizSessionService(
	queries *db.Queries,
	gamification *gamificationSvc.GamificationService,
) *quizSvc.QuizSessionService {
	return nil
}

func ProvideTestDashboardService(
	queries *db.Queries,
	gamification *gamificationSvc.GamificationService,
	authService *authSvc.AuthService,
	sessionService *quizSvc.QuizSessionService,
	cacheService *cacheSvc.CacheService,
) *dashboardSvc.DashboardService {
	return dashboardSvc.NewDashboardService(queries, queries, queries, queries, gamification, authService, sessionService, cacheService)
}

func ProvideMockAuthenticator() interfaces.Authenticator {
	return NewMockAuthenticator()
}

func ProvideMockTimeProvider() timeProvider.TimeProvider {
	return NewMockTimeProvider()
}

func ProvideTestAuthHandler(queries *db.Queries, authService *authSvc.AuthService, localeService *locales.Service) *authHndl.AuthHandler {
	return authHndl.NewAuth(queries, authService, localeService)
}

func ProvideTestDashboardHandler(dashboardService *dashboardSvc.DashboardService) *dashboardHndl.DashboardHandler {
	return dashboardHndl.NewDashboard(dashboardService)
}

func ProvideTestQuizHandler(
	queries *db.Queries,
	quizService *quizSvc.QuizService,
	authService *authSvc.AuthService,
) *quizHndl.QuizHandler {
	return quizHndl.NewQuiz(queries, quizService, nil, authService)
}

func ProvideTestAdminHandler(
	adminService *adminSvc.AdminService,
	authService *authSvc.AuthService,
	localeService *locales.Service,
	promptGenerator *adminSvc.PromptGenerator,
) *adminHndl.AdminHandler {
	return adminHndl.NewAdmin(adminService, authService, localeService, promptGenerator)
}

func ProvideTestRequireAuthMiddleware(authService *authSvc.AuthService) middleware.RequireAuthMiddleware {
	return middleware.NewRequireAuthMiddleware(authService)
}

func ProvideTestRequireRoleMiddleware(authService *authSvc.AuthService) middleware.RequireRoleMiddleware {
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
	ProvideTestPromptGenerator,
	ProvideLocaleService,
	ProvideMockAuthenticator,
	ProvideMockTimeProvider,
	wire.Bind(new(interfaces.Authenticator), new(*MockAuthenticator)),
	wire.Bind(new(timeProvider.TimeProvider), new(*MockTimeProvider)),
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
	authService := authSvc.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
	gamification := gamificationSvc.NewGamificationService(queries, queries, mockTime)
	cacheService := cacheSvc.NewCacheService(nil, cfg.Redis.CacheTTL)
	quizService := quizSvc.NewQuizService(queries, queries, queries, queries, cacheService)
	schemaService := adminSvc.NewQuestionSchema(nil)
	promptGenerator := adminSvc.NewPromptGenerator(schemaService)

	quizSessionService := quizSvc.NewQuizSessionService(queries, queries, queries, queries, queries, gamification, cacheService)
	dashboardService := dashboardSvc.NewDashboardService(queries, queries, queries, queries, gamification, authService, quizSessionService, cacheService)

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
