package di

import "context"

func InitializeApp(ctx context.Context) (*App, error) {
	config := ProvideConfig()

	pool, err := ProvideDBPool(ctx, config)
	if err != nil {
		return nil, err
	}
	queries := ProvideDB(pool)

	authService := ProvideAuthService(queries, config)

	client := ProvideRedis(config)
	cacheService := ProvideCacheService(client, config)

	quizService := ProvideQuizService(queries, cacheService)
	gamificationService := ProvideGamificationService(queries)
	quizSessionService := ProvideQuizSessionService(queries, gamificationService, cacheService)
	quizTimerService := ProvideQuizTimerService(queries, quizSessionService, cacheService, client)

	storageService, err := ProvideStorageService(config)
	if err != nil {
		return nil, err
	}

	typstClient, err := ProvideTypstGRPCClient(config)
	if err != nil {
		return nil, err
	}

	userGroupService := ProvideUserGroupService(queries)
	permissionsService := ProvidePermissionsService(queries)
	adminService := ProvideAdminService(queries, queries, authService, storageService, cacheService)
	dashboardService := ProvideDashboardService(queries, gamificationService, authService, quizSessionService, cacheService)

	localeService, err := ProvideLocaleService()
	if err != nil {
		return nil, err
	}

	promptGenerator := ProvidePromptGenerator(cacheService)
	learningMaterialService := ProvideLearningMaterialService(queries, storageService, typstClient, permissionsService)

	authHandler := ProvideAuthHandler(queries, authService, localeService)
	dashboardHandler := ProvideDashboardHandler(dashboardService)
	quizHandler := ProvideQuizHandler(queries, quizService, quizSessionService, authService)
	adminHandler := ProvideAdminHandler(adminService, authService, localeService, promptGenerator)
	editorHandler := ProvideEditorHandler(learningMaterialService)
	typstHandler := ProvideTypstHandler(learningMaterialService, adminService)
	learningMaterialsHandler := ProvideLearningMaterialsHandler(learningMaterialService, adminService, localeService)
	groupsHandler := ProvideGroupsHandler(userGroupService, queries, authService)
	permissionsHandler := ProvidePermissionsHandler(permissionsService, userGroupService, authService, queries)

	requireAuthMiddleware := ProvideRequireAuthMiddleware(authService)
	requireRoleMiddleware := ProvideRequireRoleMiddleware(authService)
	requireAssetOwnerMiddleware := ProvideRequireAssetOwnerMiddleware(authService, permissionsService, localeService)
	requireLearningMaterialAccessMiddleware := ProvideRequireLearningMaterialAccessMiddleware(authService, permissionsService, queries, localeService)
	compressionMiddleware := ProvideCompressionMiddleware()
	commonHeaders := ProvideCommonHeadersMiddleware()
	localeMiddleware := ProvideLocaleMiddleware(localeService)

	return &App{
		Config:                      config,
		AuthService:                 authService,
		QuizService:                 quizService,
		QuizSessionService:          quizSessionService,
		QuizTimerService:            quizTimerService,
		AdminService:                adminService,
		DashboardService:            dashboardService,
		GamificationService:         gamificationService,
		StorageService:              storageService,
		CacheService:                cacheService,
		LearningMaterialService:     learningMaterialService,
		UserGroupService:            userGroupService,
		PermissionsService:          permissionsService,
		AuthHandler:                 authHandler,
		DashboardHandler:            dashboardHandler,
		QuizHandler:                 quizHandler,
		AdminHandler:                adminHandler,
		EditorHandler:               editorHandler,
		TypstHandler:                typstHandler,
		LearningMaterialsHandler:    learningMaterialsHandler,
		GroupsHandler:               groupsHandler,
		PermissionsHandler:          permissionsHandler,
		RequireAuthMiddleware:                  requireAuthMiddleware,
		RequireRoleMiddleware:                  requireRoleMiddleware,
		RequireAssetOwnerMiddleware:            requireAssetOwnerMiddleware,
		RequireLearningMaterialAccessMiddleware: requireLearningMaterialAccessMiddleware,
		CompressionMiddleware:                  compressionMiddleware,
		CommonHeadersMiddleware:                commonHeaders,
		LocaleMiddleware:                       localeMiddleware,
		LocaleService:                          localeService,
	}, nil
}
