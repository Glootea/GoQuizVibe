package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/config"
	"github.com/goquizvibe/handlers"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
)

func main() {
	cfg := config.Load()

	memStore := store.NewMemoryStore()

	teacherID := uuid.New()
	teacher := &models.User{
		ID:           teacherID,
		Name:         "Учитель",
		Email:        "teacher@example.com",
		Role:         models.RoleTeacher,
		PasswordHash: "$2a$10$dummy",
		CreatedAt:    time.Now(),
	}
	memStore.CreateUser(teacher)

	memStore.SeedMathQuizzes(teacherID)

	authService := services.NewAuthService(memStore, cfg.JWTSecret)
	quizService := services.NewQuizService(memStore)
	gamification := services.NewGamificationService(memStore)

	authHandler := handlers.NewAuth(memStore, authService)
	dashboardHandler := handlers.NewDashboard(memStore, quizService, gamification, authService)
	quizHandler := handlers.NewQuiz(memStore, quizService, gamification, authService)

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
		switch r.Method {
		case "GET":
			if len(r.URL.Path) > len("/quiz/") && r.URL.Path[len(r.URL.Path)-7:] == "/result" {
				quizHandler.QuizResult(w, r)
			} else {
				quizHandler.QuizPage(w, r)
			}
		case "POST":
			quizHandler.QuizSubmitHTMX(w, r)
		default:
			http.NotFound(w, r)
		}
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
