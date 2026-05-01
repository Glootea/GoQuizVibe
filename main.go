package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/goquizvibe/config"
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
		http.NotFound(w, r)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			authHandler.LoginPage(w, r)
		} else if r.Method == "POST" {
			authHandler.LoginSubmit(w, r)
		}
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			authHandler.RegisterPage(w, r)
		} else if r.Method == "POST" {
			authHandler.RegisterSubmit(w, r)
		}
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		authHandler.Logout(w, r)
	})

	mux.HandleFunc("/dashboard", dashboardHandler.DashboardPage)

	mux.HandleFunc("/quiz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			quizHandler.QuizPage(w, r)
		}
	})

	mux.HandleFunc("/quiz/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method == "POST" && strings.HasSuffix(path, "/submit") {
			quizHandler.QuizSubmitHTMX(w, r)
			return
		}
		if r.Method == "GET" && strings.Contains(path, "/next") {
			quizHandler.QuizNextHTMX(w, r)
			return
		}
		if r.Method == "GET" && strings.HasSuffix(path, "/result") {
			quizHandler.QuizResult(w, r)
			return
		}
		if r.Method == "GET" {
			quizHandler.QuizPage(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/errors", func(w http.ResponseWriter, r *http.Request) {
		quizHandler.ErrorsPage(w, r)
	})

	mux.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		quizHandler.LeaderboardPage(w, r)
	})

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/", adminHandler.Dashboard)
	adminMux.HandleFunc("/quizzes", adminHandler.QuizzesCreate)
	adminMux.HandleFunc("/quizzes/", adminHandler.QuizOp)
	mux.Handle("/admin/", adminMiddleware(adminMux))

	mux.Handle("/admin/results", adminMiddleware(http.HandlerFunc(adminHandler.Results)))
	mux.Handle("/admin/statistics", adminMiddleware(http.HandlerFunc(adminHandler.Statistics)))

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
