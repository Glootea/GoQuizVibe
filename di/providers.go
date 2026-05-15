package di

import (
	"context"
	"time"

	"github.com/goquizvibe/config"
	"github.com/goquizvibe/database"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/locales"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/goforj/wire"
)

type App struct {
	Config *config.Config

	AuthService         *services.AuthService
	QuizService         *services.QuizService
	QuizSessionService  *services.QuizSessionService
	QuizTimerService    *services.QuizTimerService
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
}

func ProvideConfig() *config.Config {
	return config.Load()
}

func ProvideDBPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return database.Connect(ctx, *cfg)
}

func ProvideDB(pool *pgxpool.Pool) *db.Queries {
	return db.New(pool)
}

func ProvideRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: cfg.Redis.Password,
		DB:       0,
	})
}

func ProvideCacheService(client *redis.Client, cfg *config.Config) *services.CacheService {
	return services.NewCacheService(client, cfg.Redis.CacheTTL)
}

func ProvideAuthService(queries *db.Queries, cfg *config.Config) *services.AuthService {
	jwtExp := 24 * time.Hour * 7
	return services.NewAuthService(queries, cfg.JWTSecret, jwtExp)
}

func ProvideGamificationService(queries *db.Queries) *services.GamificationService {
	return services.NewGamificationService(queries, queries, services.RealTimeProvider{})
}

func ProvideQuizService(queries *db.Queries) *services.QuizService {
	return services.NewQuizService(queries, queries, queries, queries)
}

func ProvideStorageService(cfg *config.Config) (*services.StorageService, error) {
	return services.NewStorageServiceFromConfig(cfg.Minio)
}

func ProvideAdminService(
	queries *db.Queries,
	authService *services.AuthService,
	storageService *services.StorageService,
) *services.AdminService {
	return services.NewAdminService(queries, queries, queries, queries, queries, queries, authService, storageService)
}

func ProvideQuizSessionService(
	queries *db.Queries,
	gamification *services.GamificationService,
	cacheService *services.CacheService,
) *services.QuizSessionService {
	return services.NewQuizSessionService(queries, queries, queries, queries, queries, *gamification, *cacheService)
}

func ProvideQuizTimerService(
	queries *db.Queries,
	sessionService *services.QuizSessionService,
	cacheService *services.CacheService,
	redisClient *redis.Client,
) *services.QuizTimerService {
	return services.NewQuizTimerService(queries, sessionService, cacheService, redisClient)
}

func ProvideDashboardService(
	queries *db.Queries,
	gamification *services.GamificationService,
	authService *services.AuthService,
	quizSessionService *services.QuizSessionService,
) *services.DashboardService {
	return services.NewDashboardService(queries, queries, queries, queries, gamification, authService, quizSessionService)
}

func ProvideLocaleService() (*locales.Service, error) {
	return locales.NewService()
}

func ProvideAuthenticator(authService *services.AuthService) services.Authenticator {
	return authService
}

func ProvideTimeProvider() services.TimeProvider {
	return services.RealTimeProvider{}
}

func ProvideAuthHandler(queries *db.Queries, authService *services.AuthService, localeService *locales.Service) *handlers.AuthHandler {
	return handlers.NewAuth(queries, authService, localeService)
}

func ProvideDashboardHandler(dashboardService *services.DashboardService) *handlers.DashboardHandler {
	return handlers.NewDashboard(dashboardService)
}

func ProvideQuizHandler(
	queries *db.Queries,
	quizService *services.QuizService,
	quizSessionService *services.QuizSessionService,
	authService *services.AuthService,
) *handlers.QuizHandler {
	return handlers.NewQuiz(queries, quizService, quizSessionService, authService)
}

func ProvideAdminHandler(
	adminService *services.AdminService,
	authService *services.AuthService,
	localeService *locales.Service,
	promptGenerator *services.PromptGenerator,
) *handlers.AdminHandler {
	return handlers.NewAdmin(adminService, authService, localeService, promptGenerator)
}

func ProvidePromptGenerator(cacheService *services.CacheService) *services.PromptGenerator {
	schemaService := services.NewQuestionSchema(cacheService)
	return services.NewPromptGenerator(schemaService)
}

func ProvideRequireAuthMiddleware(authService *services.AuthService) middleware.RequireAuthMiddleware {
	return middleware.NewRequireAuthMiddleware(authService)
}

func ProvideRequireRoleMiddleware(authService *services.AuthService) middleware.RequireRoleMiddleware {
	return middleware.NewRequireRoleMiddleware(authService, models.RoleTeacher)
}

func ProvideCompressionMiddleware() *middleware.CompressionMiddleware {
	return middleware.NewCompressionMiddleware()
}

func ProvideCommonHeadersMiddleware() *middleware.CommonHeaders {
	return middleware.NewCommonHeaders()
}

func ProvideLocaleMiddleware(localeService *locales.Service) *middleware.LocaleMiddleware {
	return middleware.NewLocaleMiddleware(localeService)
}

var ServiceSet = wire.NewSet(
	ProvideConfig,
	ProvideDBPool,
	ProvideDB,
	ProvideRedis,
	ProvideCacheService,
	ProvideAuthService,
	ProvideGamificationService,
	ProvideQuizService,
	ProvideStorageService,
	ProvideAdminService,
	ProvideQuizSessionService,
	ProvideQuizTimerService,
	ProvideDashboardService,
	ProvideLocaleService,
	ProvideAuthenticator,
	ProvideTimeProvider,
)

var HandlerSet = wire.NewSet(
	ProvideAuthHandler,
	ProvideDashboardHandler,
	ProvideQuizHandler,
	ProvideAdminHandler,
	ProvidePromptGenerator,
)

var MiddlewareSet = wire.NewSet(
	ProvideRequireAuthMiddleware,
	ProvideRequireRoleMiddleware,
	ProvideCompressionMiddleware,
	ProvideCommonHeadersMiddleware,
	ProvideLocaleMiddleware,
)

var AppSet = wire.NewSet(
	ServiceSet,
	HandlerSet,
	MiddlewareSet,
	wire.Struct(new(App), "Config", "AuthService", "QuizService", "QuizSessionService", "QuizTimerService", "AdminService", "DashboardService", "GamificationService", "StorageService", "CacheService", "AuthHandler", "DashboardHandler", "QuizHandler", "AdminHandler", "RequireAuthMiddleware", "RequireRoleMiddleware", "CompressionMiddleware", "CommonHeadersMiddleware", "LocaleMiddleware", "LocaleService"),
)
