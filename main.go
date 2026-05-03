package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/goquizvibe/config"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/database"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
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

	jwtExp := 24 * time.Hour * 7
	authService := services.NewAuthService(queries, cfg.JWTSecret, jwtExp)
	quizService := services.NewQuizService(queries)
	gamification := services.NewGamificationService(queries)
	storageService, err := services.NewStorageService(cfg.Minio)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	if err := storageService.EnsureBucket(ctx); err != nil {
		log.Printf("Warning: Failed to ensure bucket exists: %v", err)
	}

	authHandler := handlers.NewAuth(queries, authService)
	dashboardHandler := handlers.NewDashboard(queries, quizService, gamification, authService)
	quizHandler := handlers.NewQuiz(queries, quizService, gamification, authService)
	adminHandler := handlers.NewAdmin(queries, authService, storageService)

	adminMiddleware := authHandler.RequireRole(models.RoleTeacher)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == "GET" {
			pages.LandingPage().Render(r.Context(), w)
			return
		}
		pages.NotFoundPage().Render(r.Context(), w)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			authHandler.LoginPage(w, r)
		case "POST":
			authHandler.LoginSubmit(w, r)
		}
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			authHandler.RegisterPage(w, r)
		case "POST":
			authHandler.RegisterSubmit(w, r)
		}
	})

	mux.HandleFunc("/logout", handlers.ErrorHandler(authHandler.Logout))

	mux.HandleFunc("/dashboard", handlers.ErrorHandler(dashboardHandler.DashboardPage))

	mux.HandleFunc("/quiz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			quizHandler.QuizPage(w, r)
		}
	})

	mux.HandleFunc("/quiz/", handlers.ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
		path := r.URL.Path
		if r.Method == "POST" && strings.HasSuffix(path, "/submit") {
			return quizHandler.QuizSubmitHTMX(w, r)
		}
		if r.Method == "GET" && strings.Contains(path, "/next") {
			return quizHandler.QuizNextHTMX(w, r)
		}
		if r.Method == "GET" && strings.HasSuffix(path, "/result") {
			return quizHandler.QuizResult(w, r)
		}
		if r.Method == "GET" {
			return quizHandler.QuizPage(w, r)
		}
		return ce.ErrNotFound
	}))

	mux.HandleFunc("/errors", handlers.ErrorHandler(quizHandler.ErrorsPage))

	mux.HandleFunc("/leaderboard", handlers.ErrorHandler(quizHandler.LeaderboardPage))

	mux.Handle("/admin", adminMiddleware(handlers.ErrorHandler(adminHandler.Dashboard)))
	mux.Handle("/admin/quizzes", adminMiddleware(handlers.ErrorHandler(adminHandler.Quizzes)))
	mux.Handle("/admin/quizzes/new", adminMiddleware(handlers.ErrorHandler(adminHandler.QuizzesNew)))

	mux.Handle("/admin/quizzes/", adminMiddleware(handlers.ErrorHandler(adminHandler.QuizEditOp)))

	mux.Handle("/admin/results", adminMiddleware(handlers.ErrorHandler(adminHandler.Results)))
	mux.Handle("/admin/statistics", adminMiddleware(handlers.ErrorHandler(adminHandler.Statistics)))
	mux.Handle("/admin/api/quiz-stats", adminMiddleware(handlers.ErrorHandler(adminHandler.QuizStatsData)))
	mux.Handle("/admin/api/grade-dist", adminMiddleware(handlers.ErrorHandler(adminHandler.GradeDistData)))
	mux.Handle("/admin/api/subject-dist", adminMiddleware(handlers.ErrorHandler(adminHandler.SubjectDistData)))

	mux.Handle("/admin/quizzes/*/restore", adminMiddleware(handlers.ErrorHandler(adminHandler.RestoreQuiz)))

	log.Printf("Server starting on http://localhost:%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, withCommonHeaders(mux)))
}

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

var _ = fmt.Sprintf
