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

	authHandler := handlers.NewAuth(queries, authService)
	dashboardHandler := handlers.NewDashboard(queries, quizService, gamification, authService)
	quizHandler := handlers.NewQuiz(queries, quizService, gamification, authService)
	adminHandler := handlers.NewAdmin(queries, authService)

	adminMiddleware := authHandler.RequireRole(models.RoleTeacher)

	mux := http.NewServeMux()

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

	adminRoutes := map[string]func(w http.ResponseWriter, r *http.Request) error{
		"":                 adminHandler.Dashboard,
		"quizzes":          adminHandler.QuizzesCreate,
		"quizzes/":         adminHandler.QuizOp,
		"results":          adminHandler.Results,
		"statistics":       adminHandler.Statistics,
		"api/quiz-stats":   adminHandler.QuizStatsData,
		"api/grade-dist":   adminHandler.GradeDistData,
		"api/subject-dist": adminHandler.SubjectDistData,
	}

	for path, handler := range adminRoutes {
		fullPath := "/admin/" + path
		if path == "" {
			fullPath = "/admin"
		}
		mux.Handle(fullPath, adminMiddleware(handlers.ErrorHandlerFunc(handler)))
	}

	mux.Handle("/admin/quizzes/*/restore", adminMiddleware(handlers.ErrorHandlerFunc(adminHandler.RestoreQuiz)))

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
