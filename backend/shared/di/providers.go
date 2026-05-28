package di

import (
	"context"
	"fmt"
	"time"

	"github.com/goquizvibe/backend/shared/config"
	"github.com/goquizvibe/backend/shared/database"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/locales"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/goforj/wire"

	adminHndl "github.com/goquizvibe/backend/feature/admin/handlers"
	adminSvc "github.com/goquizvibe/backend/feature/admin/services"
	authHndl "github.com/goquizvibe/backend/feature/auth/handlers"
	authSvc "github.com/goquizvibe/backend/feature/auth/services"
	dashboardHnld "github.com/goquizvibe/backend/feature/dashboard/handlers"
	dashboardSrv "github.com/goquizvibe/backend/feature/dashboard/services"
	editorHndl "github.com/goquizvibe/backend/feature/editor/handlers"
	gamificationSvc "github.com/goquizvibe/backend/feature/gamification/services"
	learningMaterialsHdl "github.com/goquizvibe/backend/feature/learning_materials/handlers"
	learningMaterialsSvc "github.com/goquizvibe/backend/feature/learning_materials/services"
	permissionsHdl "github.com/goquizvibe/backend/feature/permissions/handlers"
	permissionsSvc "github.com/goquizvibe/backend/feature/permissions/services"
	quizHndl "github.com/goquizvibe/backend/feature/quiz/handlers"
	quizSvc "github.com/goquizvibe/backend/feature/quiz/services"
	cacheService "github.com/goquizvibe/backend/shared/infrastructure/cache"
	storageService "github.com/goquizvibe/backend/shared/infrastructure/storage"
	timeprovider "github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
)

type App struct {
	Config *config.Config

	AuthService             *authSvc.AuthService
	QuizService             *quizSvc.QuizService
	QuizSessionService      *quizSvc.QuizSessionService
	QuizTimerService        *quizSvc.QuizTimerService
	AdminService            *adminSvc.AdminService
	DashboardService        *dashboardSrv.DashboardService
	GamificationService     *gamificationSvc.GamificationService
	StorageService          *storageService.StorageService
	CacheService            *cacheService.CacheService
	LearningMaterialService *learningMaterialsSvc.LearningMaterialService
	UserGroupService        *permissionsSvc.UserGroupService
	PermissionsService      *permissionsSvc.PermissionsService

	AuthHandler              *authHndl.AuthHandler
	DashboardHandler         *dashboardHnld.DashboardHandler
	QuizHandler              *quizHndl.QuizHandler
	AdminHandler             *adminHndl.AdminHandler
	LearningMaterialsHandler *learningMaterialsHdl.LearningMaterialsHandler
	EditorHandler            *editorHndl.EditorHandler
	TypstHandler             *learningMaterialsHdl.TypstHandler
	GroupsHandler            *permissionsHdl.GroupsHandler
	PermissionsHandler       *permissionsHdl.PermissionsHandler

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

func ProvideAuthService(queries *db.Queries, cfg *config.Config) *authSvc.AuthService {
	jwtExp := 24 * time.Hour * 7
	return authSvc.NewAuthService(queries, cfg.JWTSecret, jwtExp)
}

func ProvideGamificationService(queries *db.Queries) *gamificationSvc.GamificationService {
	return gamificationSvc.NewGamificationService(queries, queries, timeprovider.RealTimeProvider{})
}

func ProvideQuizService(queries *db.Queries, cacheService *cacheService.CacheService) *quizSvc.QuizService {
	return quizSvc.NewQuizService(queries, queries, queries, queries, cacheService)
}

func ProvideStorageService(cfg *config.Config) (*storageService.StorageService, error) {
	return storageService.NewStorageServiceFromConfig(cfg.Minio)
}

func ProvideAdminService(
	queries *db.Queries,
	authService *authSvc.AuthService,
	storageService *storageService.StorageService,
	cacheService *cacheService.CacheService,
) *adminSvc.AdminService {
	return adminSvc.NewAdminService(queries, queries, queries, queries, queries, queries, queries, authService, storageService, cacheService)
}

func ProvideQuizSessionService(
	queries *db.Queries,
	gamification *gamificationSvc.GamificationService,
	cacheService *cacheService.CacheService,
) *quizSvc.QuizSessionService {
	return quizSvc.NewQuizSessionService(queries, queries, queries, queries, queries, gamification, cacheService)
}

func ProvideQuizTimerService(
	queries *db.Queries,
	sessionService *quizSvc.QuizSessionService,
	cacheService *cacheService.CacheService,
	redisClient *redis.Client,
) *quizSvc.QuizTimerService {
	return quizSvc.NewQuizTimerService(queries, sessionService, cacheService, redisClient)
}

func ProvideDashboardService(
	queries *db.Queries,
	gamification *gamificationSvc.GamificationService,
	authService *authSvc.AuthService,
	quizSessionService *quizSvc.QuizSessionService,
	cacheService *cacheService.CacheService,
) *dashboardSrv.DashboardService {
	return dashboardSrv.NewDashboardService(queries, queries, queries, queries, gamification, authService, quizSessionService, cacheService)
}

func ProvideLocaleService() (*locales.Service, error) {
	return locales.NewService()
}

func ProvideTimeProvider() timeprovider.TimeProvider {
	return timeprovider.RealTimeProvider{}
}

func ProvideAuthHandler(queries *db.Queries, authService *authSvc.AuthService, localeService *locales.Service) *authHndl.AuthHandler {
	return authHndl.NewAuth(queries, authService, localeService)
}

func ProvideDashboardHandler(dashboardService *dashboardSrv.DashboardService) *dashboardHnld.DashboardHandler {
	return dashboardHnld.NewDashboard(dashboardService)
}

func ProvideQuizHandler(
	queries *db.Queries,
	quizService *quizSvc.QuizService,
	quizSessionService *quizSvc.QuizSessionService,
	authService *authSvc.AuthService,
) *quizHndl.QuizHandler {
	return quizHndl.NewQuiz(queries, quizService, quizSessionService, authService)
}

func ProvideAdminHandler(
	adminService *adminSvc.AdminService,
	authService *authSvc.AuthService,
	localeService *locales.Service,
	promptGenerator *adminSvc.PromptGenerator,
) *adminHndl.AdminHandler {
	return adminHndl.NewAdmin(adminService, authService, localeService, promptGenerator)
}

func ProvidePromptGenerator(cacheService *cacheService.CacheService) *adminSvc.PromptGenerator {
	schemaService := adminSvc.NewQuestionSchema(cacheService)
	return adminSvc.NewPromptGenerator(schemaService)
}

func ProvideTypstGRPCClient(cfg *config.Config) (*learningMaterialsSvc.TypstGRPCClient, error) {
	return learningMaterialsSvc.NewTypstGRPCClient(cfg)
}

func ProvideLearningMaterialService(
	queries *db.Queries,
	storageService *storageService.StorageService,
	typstClient *learningMaterialsSvc.TypstGRPCClient,
) *learningMaterialsSvc.LearningMaterialService {
	return learningMaterialsSvc.NewLearningMaterialService(queries, storageService, typstClient)
}

func ProvideLearningMaterialsHandler(
	materialService *learningMaterialsSvc.LearningMaterialService,
	adminService *adminSvc.AdminService,
	localeService *locales.Service,
) *learningMaterialsHdl.LearningMaterialsHandler {
	return learningMaterialsHdl.NewLearningMaterialsHandler(materialService, adminService, localeService)
}

func ProvideEditorHandler(materialService *learningMaterialsSvc.LearningMaterialService) *editorHndl.EditorHandler {
	return editorHndl.NewEditor(materialService)
}

func ProvideTypstHandler(
	materialService *learningMaterialsSvc.LearningMaterialService,
	adminService *adminSvc.AdminService,
) *learningMaterialsHdl.TypstHandler {
	return learningMaterialsHdl.NewTypstHandler(materialService, adminService)
}

func ProvideRequireAuthMiddleware(authService *authSvc.AuthService) middleware.RequireAuthMiddleware {
	return middleware.NewRequireAuthMiddleware(authService)
}

func ProvideRequireRoleMiddleware(authService *authSvc.AuthService) middleware.RequireRoleMiddleware {
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

func ProvideUserGroupService(queries *db.Queries) *permissionsSvc.UserGroupService {
	return permissionsSvc.NewUserGroupService(queries, queries)
}

func ProvidePermissionsService(queries *db.Queries) *permissionsSvc.PermissionsService {
	return permissionsSvc.NewPermissionsService(queries, queries)
}

func ProvideGroupsHandler(userGroupService *permissionsSvc.UserGroupService, authService *authSvc.AuthService) *permissionsHdl.GroupsHandler {
	return permissionsHdl.NewGroupsHandler(userGroupService, authService)
}

func ProvidePermissionsHandler(permissionsService *permissionsSvc.PermissionsService, userGroupService *permissionsSvc.UserGroupService, authService *authSvc.AuthService) *permissionsHdl.PermissionsHandler {
	return permissionsHdl.NewPermissionsHandler(permissionsService, userGroupService, authService)
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
	wire.Struct(new(App), "Config", "AuthService", "QuizService", "QuizSessionService", "QuizTimerService", "AdminService", "DashboardService", "GamificationService", "StorageService", "CacheService", "LearningMaterialService", "AuthHandler", "DashboardHandler", "QuizHandler", "AdminHandler", "EditorHandler", "TypstHandler", "RequireAuthMiddleware", "RequireRoleMiddleware", "CompressionMiddleware", "CommonHeadersMiddleware", "LocaleMiddleware", "LocaleService", "UserGroupService", "PermissionsService", "GroupsHandler", "PermissionsHandler"),
)
