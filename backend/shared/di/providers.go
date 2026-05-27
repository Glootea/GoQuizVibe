package di

import (
	"context"
	"fmt"
	"time"

	"github.com/goquizvibe/backend/shared/config"
	"github.com/goquizvibe/backend/shared/database"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/models"
	"github.com/goquizvibe/backend/shared/locales"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/goforj/wire"

	authServices "github.com/goquizvibe/backend/feature/auth/services"
	authHandlers "github.com/goquizvibe/backend/feature/auth/handlers"
	dashboardServices "github.com/goquizvibe/backend/feature/dashboard/services"
	dashboardHandlers "github.com/goquizvibe/backend/feature/dashboard/handlers"
	quizServices "github.com/goquizvibe/backend/feature/quiz/services"
	quizHandlers "github.com/goquizvibe/backend/feature/quiz/handlers"
	adminServices "github.com/goquizvibe/backend/feature/admin/services"
	adminHandlers "github.com/goquizvibe/backend/feature/admin/handlers"
	learningMaterialsServices "github.com/goquizvibe/backend/feature/learning_materials/services"
	learningMaterialsHandlers "github.com/goquizvibe/backend/feature/learning_materials/handlers"
	editorHandlers "github.com/goquizvibe/backend/feature/editor/handlers"
	gamificationServices "github.com/goquizvibe/backend/feature/gamification/services"
	permissionsServices "github.com/goquizvibe/backend/feature/permissions/services"
	permissionsHandlers "github.com/goquizvibe/backend/feature/permissions/handlers"
	cacheService "github.com/goquizvibe/backend/shared/infrastructure/cache"
	storageService "github.com/goquizvibe/backend/shared/infrastructure/storage"
	timeprovider "github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
)

type App struct {
	Config *config.Config

	AuthService             *authServices.AuthService
	QuizService             *quizServices.QuizService
	QuizSessionService      *quizServices.QuizSessionService
	QuizTimerService        *quizServices.QuizTimerService
	AdminService            *adminServices.AdminService
	DashboardService        *dashboardServices.DashboardService
	GamificationService     *gamificationServices.GamificationService
	StorageService          *storageService.StorageService
	CacheService            *cacheService.CacheService
	LearningMaterialService *learningMaterialsServices.LearningMaterialService
	UserGroupService        *permissionsServices.UserGroupService
	PermissionsService      *permissionsServices.PermissionsService

	AuthHandler              *authHandlers.AuthHandler
	DashboardHandler         *dashboardHandlers.DashboardHandler
	QuizHandler              *quizHandlers.QuizHandler
	AdminHandler             *adminHandlers.AdminHandler
	LearningMaterialsHandler *learningMaterialsHandlers.LearningMaterialsHandler
	EditorHandler            *editorHandlers.EditorHandler
	TypstHandler             *learningMaterialsHandlers.TypstHandler
	GroupsHandler           *permissionsHandlers.GroupsHandler
	PermissionsHandler       *permissionsHandlers.PermissionsHandler

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
		Addr:     fmt.Sprintf("%s:6379", cfg.Redis.Host),
		Password: cfg.Redis.Password,
		DB:       0,
	})
}

func ProvideCacheService(client *redis.Client, cfg *config.Config) *cacheService.CacheService {
	return cacheService.NewCacheService(client, cfg.Redis.CacheTTL)
}

func ProvideAuthService(queries *db.Queries, cfg *config.Config) *authServices.AuthService {
	jwtExp := 24 * time.Hour * 7
	return authServices.NewAuthService(queries, cfg.JWTSecret, jwtExp)
}

func ProvideGamificationService(queries *db.Queries) *gamificationServices.GamificationService {
	return gamificationServices.NewGamificationService(queries, queries, timeprovider.RealTimeProvider{})
}

func ProvideQuizService(queries *db.Queries, cacheService *cacheService.CacheService) *quizServices.QuizService {
	return quizServices.NewQuizService(queries, queries, queries, queries, cacheService)
}

func ProvideStorageService(cfg *config.Config) (*storageService.StorageService, error) {
	return storageService.NewStorageServiceFromConfig(cfg.Minio)
}

func ProvideAdminService(
	queries *db.Queries,
	authService *authServices.AuthService,
	storageService *storageService.StorageService,
	cacheService *cacheService.CacheService,
) *adminServices.AdminService {
	return adminServices.NewAdminService(queries, queries, queries, queries, queries, queries, queries, authService, storageService, cacheService)
}

func ProvideQuizSessionService(
	queries *db.Queries,
	gamification *gamificationServices.GamificationService,
	cacheService *cacheService.CacheService,
) *quizServices.QuizSessionService {
	return quizServices.NewQuizSessionService(queries, queries, queries, queries, queries, gamification, cacheService)
}

func ProvideQuizTimerService(
	queries *db.Queries,
	sessionService *quizServices.QuizSessionService,
	cacheService *cacheService.CacheService,
	redisClient *redis.Client,
) *quizServices.QuizTimerService {
	return quizServices.NewQuizTimerService(queries, sessionService, cacheService, redisClient)
}

func ProvideDashboardService(
	queries *db.Queries,
	gamification *gamificationServices.GamificationService,
	authService *authServices.AuthService,
	quizSessionService *quizServices.QuizSessionService,
	cacheService *cacheService.CacheService,
) *dashboardServices.DashboardService {
	return dashboardServices.NewDashboardService(queries, queries, queries, queries, gamification, authService, quizSessionService, cacheService)
}

func ProvideLocaleService() (*locales.Service, error) {
	return locales.NewService()
}

func ProvideTimeProvider() timeprovider.TimeProvider {
	return timeprovider.RealTimeProvider{}
}

func ProvideAuthHandler(queries *db.Queries, authService *authServices.AuthService, localeService *locales.Service) *authHandlers.AuthHandler {
	return authHandlers.NewAuth(queries, authService, localeService)
}

func ProvideDashboardHandler(dashboardService *dashboardServices.DashboardService) *dashboardHandlers.DashboardHandler {
	return dashboardHandlers.NewDashboard(dashboardService)
}

func ProvideQuizHandler(
	queries *db.Queries,
	quizService *quizServices.QuizService,
	quizSessionService *quizServices.QuizSessionService,
	authService *authServices.AuthService,
) *quizHandlers.QuizHandler {
	return quizHandlers.NewQuiz(queries, quizService, quizSessionService, authService)
}

func ProvideAdminHandler(
	adminService *adminServices.AdminService,
	authService *authServices.AuthService,
	localeService *locales.Service,
	promptGenerator *adminServices.PromptGenerator,
) *adminHandlers.AdminHandler {
	return adminHandlers.NewAdmin(adminService, authService, localeService, promptGenerator)
}

func ProvidePromptGenerator(cacheService *cacheService.CacheService) *adminServices.PromptGenerator {
	schemaService := adminServices.NewQuestionSchema(cacheService)
	return adminServices.NewPromptGenerator(schemaService)
}

func ProvideTypstGRPCClient(cfg *config.Config) (*learningMaterialsServices.TypstGRPCClient, error) {
	return learningMaterialsServices.NewTypstGRPCClient(cfg)
}

func ProvideLearningMaterialService(
	queries *db.Queries,
	storageService *storageService.StorageService,
	typstClient *learningMaterialsServices.TypstGRPCClient,
) *learningMaterialsServices.LearningMaterialService {
	return learningMaterialsServices.NewLearningMaterialService(queries, storageService, typstClient)
}

func ProvideLearningMaterialsHandler(
	materialService *learningMaterialsServices.LearningMaterialService,
	adminService *adminServices.AdminService,
	localeService *locales.Service,
) *learningMaterialsHandlers.LearningMaterialsHandler {
	return learningMaterialsHandlers.NewLearningMaterialsHandler(materialService, adminService, localeService)
}

func ProvideEditorHandler(materialService *learningMaterialsServices.LearningMaterialService) *editorHandlers.EditorHandler {
	return editorHandlers.NewEditor(materialService)
}

func ProvideTypstHandler(
	materialService *learningMaterialsServices.LearningMaterialService,
	adminService *adminServices.AdminService,
) *learningMaterialsHandlers.TypstHandler {
	return learningMaterialsHandlers.NewTypstHandler(materialService, adminService)
}

func ProvideUserGroupService(queries *db.Queries) *permissionsServices.UserGroupService {
	return permissionsServices.NewUserGroupService(queries, queries)
}

func ProvidePermissionsService(queries *db.Queries) *permissionsServices.PermissionsService {
	return permissionsServices.NewPermissionsService(queries, queries)
}

func ProvideGroupsHandler(
	groupService *permissionsServices.UserGroupService,
	authService *authServices.AuthService,
) *permissionsHandlers.GroupsHandler {
	return permissionsHandlers.NewGroupsHandler(groupService, authService)
}

func ProvidePermissionsHandler(
	permService *permissionsServices.PermissionsService,
	groupService *permissionsServices.UserGroupService,
	authService *authServices.AuthService,
) *permissionsHandlers.PermissionsHandler {
	return permissionsHandlers.NewPermissionsHandler(permService, groupService, authService)
}

func ProvideRequireAuthMiddleware(authService *authServices.AuthService) middleware.RequireAuthMiddleware {
	return middleware.NewRequireAuthMiddleware(authService)
}

func ProvideRequireRoleMiddleware(authService *authServices.AuthService) middleware.RequireRoleMiddleware {
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
	ProvideTimeProvider,
	ProvideTypstGRPCClient,
	ProvideLearningMaterialService,
	ProvideUserGroupService,
	ProvidePermissionsService,
)

var HandlerSet = wire.NewSet(
	ProvideAuthHandler,
	ProvideDashboardHandler,
	ProvideQuizHandler,
	ProvideAdminHandler,
	ProvideEditorHandler,
	ProvideTypstHandler,
	ProvidePromptGenerator,
	ProvideLearningMaterialsHandler,
	ProvideGroupsHandler,
	ProvidePermissionsHandler,
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
	wire.Struct(new(App), "Config", "AuthService", "QuizService", "QuizSessionService", "QuizTimerService", "AdminService", "DashboardService", "GamificationService", "StorageService", "CacheService", "LearningMaterialService", "UserGroupService", "PermissionsService", "AuthHandler", "DashboardHandler", "QuizHandler", "AdminHandler", "EditorHandler", "TypstHandler", "GroupsHandler", "PermissionsHandler", "RequireAuthMiddleware", "RequireRoleMiddleware", "CompressionMiddleware", "CommonHeadersMiddleware", "LocaleMiddleware", "LocaleService"),
)
