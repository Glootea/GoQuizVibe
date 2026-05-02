package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/goquizvibe/config"
	"github.com/goquizvibe/database"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
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

	repo := store.NewRepository(queries)

	authService := services.NewAuthService(repo, cfg.JWTSecret)
	quizService := services.NewQuizService(repo)
	gamification := services.NewGamificationService(repo)

	authHandler := handlers.NewAuth(repo, authService)
	dashboardHandler := handlers.NewDashboard(repo, quizService, gamification, authService)
	quizHandler := handlers.NewQuiz(repo, quizService, gamification, authService)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == "GET" {
			pages.LandingPage().Render(r.Context(), w)
			return
		}
		http.NotFound(w, r)
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
