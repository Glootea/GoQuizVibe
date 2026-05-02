package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/goquizvibe/config"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/database"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(*cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.SeedData(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	repo := store.NewRepository(db)

	authService := services.NewAuthService(repo, cfg.JWTSecret)
	quizService := services.NewQuizService(repo)
	gamification := services.NewGamificationService(repo)

	authHandler := handlers.NewAuth(repo, authService)
	dashboardHandler := handlers.NewDashboard(repo, quizService, gamification, authService)
	quizHandler := handlers.NewQuiz(repo, quizService, gamification, authService)
	adminHandler := handlers.NewAdmin(repo, authService)

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
		"":                  adminHandler.Dashboard,
		"quizzes":           adminHandler.QuizzesCreate,
		"quizzes/":          adminHandler.QuizOp,
		"quizzes/*/restore": adminHandler.RestoreQuiz,
		"results":           adminHandler.Results,
		"statistics":        adminHandler.Statistics,
	}

	for path, handler := range adminRoutes {
		errorHandler := handlers.ErrorHandlerFunc(handler)
		mux.Handle("/admin/"+path, adminMiddleware(errorHandler))
	}

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
