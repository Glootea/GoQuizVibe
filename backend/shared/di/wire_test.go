package di

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	adminHandlers "github.com/goquizvibe/backend/feature/admin/handlers"
	adminServices "github.com/goquizvibe/backend/feature/admin/services"
	authHandlers "github.com/goquizvibe/backend/feature/auth/handlers"
	authServices "github.com/goquizvibe/backend/feature/auth/services"
	dashboardHandlers "github.com/goquizvibe/backend/feature/dashboard/handlers"
	dashboardServices "github.com/goquizvibe/backend/feature/dashboard/services"
	gamificationServices "github.com/goquizvibe/backend/feature/gamification/services"
	quizHandlers "github.com/goquizvibe/backend/feature/quiz/handlers"
	quizServices "github.com/goquizvibe/backend/feature/quiz/services"
	"github.com/goquizvibe/backend/shared/config"
	"github.com/goquizvibe/backend/shared/db"
	cacheServices "github.com/goquizvibe/backend/shared/infrastructure/cache"
	interfaces "github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	storageServices "github.com/goquizvibe/backend/shared/infrastructure/storage"
	timeProvider "github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
	locales "github.com/goquizvibe/backend/shared/locales"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/models"
	r "github.com/goquizvibe/backend/shared/repositories"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goforj/wire"
)

type TestApp struct {
	AuthService         *authServices.AuthService
	QuizService         *quizServices.QuizService
	QuizSessionService  *quizServices.QuizSessionService
	AdminService        *adminServices.AdminService
	DashboardService    *dashboardServices.DashboardService
	GamificationService *gamificationServices.GamificationService
	StorageService      *storageServices.StorageService
	CacheService        *cacheServices.CacheService

	AuthHandler      *authHandlers.AuthHandler
	DashboardHandler *dashboardHandlers.DashboardHandler
	QuizHandler      *quizHandlers.QuizHandler
	AdminHandler     *adminHandlers.AdminHandler

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

func NewMockGamificationService() *gamificationServices.GamificationService {
	return nil
}

func ProvideTestConfig() *config.Config {
	cfg := config.Load()
	return cfg
}

func ProvideTestCacheService() *cacheServices.CacheService {
	return nil
}

func ProvideTestPromptGenerator(cacheService *cacheServices.CacheService) *adminServices.PromptGenerator {
	schemaService := adminServices.NewQuestionSchema(cacheService)
	return adminServices.NewPromptGenerator(schemaService)
}

func ProvideTestStorageService() *storageServices.StorageService {
	return nil
}

func ProvideTestAuthService(queries *db.Queries, cfg *config.Config) *authServices.AuthService {
	return authServices.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
}

func ProvideTestGamificationService(queries *db.Queries, tp timeProvider.TimeProvider) *gamificationServices.GamificationService {
	return gamificationServices.NewGamificationService(queries, queries, tp)
}

func ProvideTestQuizService(queries *db.Queries, cacheService *cacheServices.CacheService) *quizServices.QuizService {
	return quizServices.NewQuizService(queries, queries, queries, queries, cacheService)
}

func ProvideTestAdminService(
	users r.UserRepository,
	quizzes r.QuizRepository,
	questions r.QuestionRepository,
	images r.ImageRepository,
	attempts r.AttemptRepository,
	stats r.StatsRepository,
	materials r.LearningMaterialRepository,
	authService *authServices.AuthService,
	storageService *storageServices.StorageService,
	cacheService *cacheServices.CacheService,
) *adminServices.AdminService {
	return adminServices.NewAdminService(users, quizzes, questions, images, attempts, stats, materials, authService, storageService, cacheService)
}

func ProvideTestQuizSessionService(
	queries *db.Queries,
	gamification *gamificationServices.GamificationService,
) *quizServices.QuizSessionService {
	return nil
}

func ProvideTestDashboardService(
	queries *db.Queries,
	gamification *gamificationServices.GamificationService,
	authService *authServices.AuthService,
	sessionService *quizServices.QuizSessionService,
	cacheService *cacheServices.CacheService,
) *dashboardServices.DashboardService {
	return dashboardServices.NewDashboardService(queries, queries, queries, queries, gamification, authService, sessionService, cacheService)
}

func ProvideMockAuthenticator() interfaces.Authenticator {
	return NewMockAuthenticator()
}

func ProvideMockTimeProvider() timeProvider.TimeProvider {
	return NewMockTimeProvider()
}

func ProvideTestAuthHandler(queries *db.Queries, authService *authServices.AuthService, localeService *locales.Service) *authHandlers.AuthHandler {
	return authHandlers.NewAuth(queries, authService, localeService)
}

func ProvideTestDashboardHandler(dashboardService *dashboardServices.DashboardService) *dashboardHandlers.DashboardHandler {
	return dashboardHandlers.NewDashboard(dashboardService)
}

func ProvideTestQuizHandler(
	queries *db.Queries,
	quizService *quizServices.QuizService,
	authService *authServices.AuthService,
) *quizHandlers.QuizHandler {
	return quizHandlers.NewQuiz(queries, quizService, nil, authService)
}

func ProvideTestAdminHandler(
	adminService *adminServices.AdminService,
	authService *authServices.AuthService,
	localeService *locales.Service,
	promptGenerator *adminServices.PromptGenerator,
) *adminHandlers.AdminHandler {
	return adminHandlers.NewAdmin(adminService, authService, localeService, promptGenerator)
}

func ProvideTestRequireAuthMiddleware(authService *authServices.AuthService) middleware.RequireAuthMiddleware {
	return middleware.NewRequireAuthMiddleware(authService)
}

func ProvideTestRequireRoleMiddleware(authService *authServices.AuthService) middleware.RequireRoleMiddleware {
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
	authService := authServices.NewAuthService(queries, cfg.JWTSecret, 24*time.Hour)
	gamification := gamificationServices.NewGamificationService(queries, queries, mockTime)
	cacheService := cacheServices.NewCacheService(nil, cfg.Redis.CacheTTL)
	quizService := quizServices.NewQuizService(queries, queries, queries, queries, cacheService)
	schemaService := adminServices.NewQuestionSchema(nil)
	promptGenerator := adminServices.NewPromptGenerator(schemaService)

	quizSessionService := quizServices.NewQuizSessionService(queries, queries, queries, queries, queries, gamification, cacheService)
	dashboardService := dashboardServices.NewDashboardService(queries, queries, queries, queries, gamification, authService, quizSessionService, cacheService)

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
		AuthHandler:             authHandlers.NewAuth(queries, authService, localeSvc),
		DashboardHandler:        dashboardHandlers.NewDashboard(dashboardService),
		QuizHandler:             quizHandlers.NewQuiz(queries, quizService, nil, authService),
		AdminHandler:            adminHandlers.NewAdmin(nil, authService, localeSvc, promptGenerator),
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
