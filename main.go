package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/goquizvibe/config"
	"github.com/goquizvibe/database"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/locales"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/services"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := database.Connect(ctx, *cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	if err := database.SeedData(ctx, pool); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	if err := database.LoadInitialDataFromFolder(ctx, pool, "initial_data"); err != nil {
		log.Fatalf("Failed to load initial data: %v", err)
	}

	jwtExp := 24 * time.Hour * 7
	authService := services.NewAuthService(queries, cfg.JWTSecret, jwtExp)
	quizService := services.NewQuizService(queries, queries, queries, queries)
	gamification := services.NewGamificationService(queries, queries, services.RealTimeProvider{})
	storageService, err := services.NewStorageServiceFromConfig(cfg.Minio)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	if err := storageService.EnsureBucket(ctx); err != nil {
		log.Printf("Warning: Failed to ensure bucket exists: %v", err)
	}

	adminService := services.NewAdminService(queries, queries, queries, queries, queries, queries, authService, storageService)
	quizSessionService := services.NewQuizSessionService(queries, queries, queries, queries, queries, queries, gamification)
	dashboardService := services.NewDashboardService(queries, queries, queries, queries, gamification, authService)

	localeService, err := locales.NewService()
	if err != nil {
		log.Fatalf("Failed to initialize locale service: %v", err)
	}

	authHandler := handlers.NewAuth(queries, authService, localeService)
	dashboardHandler := handlers.NewDashboard(dashboardService)
	quizHandler := handlers.NewQuiz(queries, quizService, quizSessionService, authService)
	adminHandler := handlers.NewAdmin(adminService, authService, localeService)

	requireAuthMiddleware := middleware.NewRequireAuthMiddleware(authService).Wrap
	requiredRoleMiddleware := middleware.NewRequireRoleMiddleware(authService, models.RoleTeacher).Wrap
	compressionMiddleware := middleware.NewCompressionMiddleware().Wrap
	commonHeadersMiddleware := middleware.NewCommonHeaders().Wrap
	localeMiddleware := middleware.NewLocaleMiddleware(localeService).Wrap

	mux := http.NewServeMux()

	type Route struct {
		Method  string
		Pattern string
		Handler func(w http.ResponseWriter, r *http.Request) error
	}

	wrapHandler := func(handler any) http.HandlerFunc {
		switch h := handler.(type) {
		case func(w http.ResponseWriter, r *http.Request):
			return h
		case func(w http.ResponseWriter, r *http.Request) error:
			return handlers.ErrorHandler(h)
		case http.Handler:
			return h.ServeHTTP
		default:
			panic("unknown handler type")
		}
	}

	routes := []Route{
		{"GET", "/", authHandler.LandingPage},
		{"GET", "/login", authHandler.LoginPage},
		{"GET", "/register", authHandler.RegisterPage},
		{"POST", "/login", authHandler.LoginSubmit},
		{"POST", "/register", authHandler.RegisterSubmit},
		{"GET", "/logout", authHandler.Logout},
		{"GET", "/dashboard", dashboardHandler.DashboardPage},
		{"GET", "/quiz/{id}", quizHandler.QuizStart},
		{"GET", "/quiz/{id}/q/{index}", quizHandler.QuizQuestion},
		{"POST", "/quiz/{id}/submit", quizHandler.QuizSubmitHTMX},
		{"GET", "/quiz/{id}/result", quizHandler.QuizResult},
		{"GET", "/errors", quizHandler.ErrorsPage},
		{"GET", "/leaderboard", quizHandler.LeaderboardPage},
		{"GET", "/admin", adminHandler.Dashboard},
		{"GET", "/admin/quizzes", adminHandler.Quizzes},
		{"GET", "/admin/quizzes/new", adminHandler.QuizzesNew},
		{"GET", "/admin/quizzes/{id}", adminHandler.QuizView},
		{"PUT", "/admin/quizzes/{id}", adminHandler.QuizUpdate},
		{"DELETE", "/admin/quizzes/{id}", adminHandler.QuizDelete},
		{"POST", "/admin/quizzes/{id}/question", adminHandler.AddQuestion},
		{"PUT", "/admin/quizzes/{id}/question/{qid}", adminHandler.UpdateQuestion},
		{"DELETE", "/admin/quizzes/{id}/question/{qid}", adminHandler.DeleteQuestion},
		{"POST", "/admin/quizzes/{id}/question/{qid}/image", adminHandler.UploadQuestionImage},
		{"DELETE", "/admin/quizzes/{id}/question/{qid}/image/{imgid}", adminHandler.DeleteQuestionImage},
		{"POST", "/admin/quizzes/{id}/restore", adminHandler.RestoreQuiz},
		{"GET", "/admin/results", adminHandler.Results},
		{"GET", "/admin/statistics", adminHandler.Statistics},
		{"GET", "/admin/api/quiz-stats", adminHandler.QuizStatsData},
		{"GET", "/admin/api/grade-dist", adminHandler.GradeDistData},
		{"GET", "/admin/api/subject-dist", adminHandler.SubjectDistData},
	}

	wrapRoute := func(r Route) http.HandlerFunc {
		wrapped := wrapHandler(r.Handler)
		if r.Pattern == "/" || r.Pattern == "/login" || r.Pattern == "/register" {

		} else if strings.HasPrefix(r.Pattern, "/admin") {
			wrapped = requiredRoleMiddleware(wrapped)
		} else {
			wrapped = requireAuthMiddleware(wrapped)
		}
		wrapped = compressionMiddleware(wrapped)
		wrapped = commonHeadersMiddleware(wrapped)
		wrapped = localeMiddleware(wrapped)

		return wrapped
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	for _, r := range routes {
		mux.HandleFunc(r.Method+" "+r.Pattern, wrapRoute(r))
	}

	log.Printf("Server starting on http://localhost:%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, mux))
}
