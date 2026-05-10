package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/goquizvibe/database"
	"github.com/goquizvibe/di"
	"github.com/goquizvibe/handlers"
)

func main() {
	ctx := context.Background()

	app, err := di.InitializeApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	pool, err := database.Connect(ctx, *app.Config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.SeedData(ctx, pool); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	if err := database.LoadInitialDataFromFolder(ctx, pool, "initial_data"); err != nil {
		log.Fatalf("Failed to load initial data: %v", err)
	}

	if err := app.StorageService.EnsureBucket(ctx); err != nil {
		log.Printf("Warning: Failed to ensure bucket exists: %v", err)
	}

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
		{"GET", "/", app.AuthHandler.LandingPage},
		{"GET", "/login", app.AuthHandler.LoginPage},
		{"GET", "/register", app.AuthHandler.RegisterPage},
		{"POST", "/login", app.AuthHandler.LoginSubmit},
		{"POST", "/register", app.AuthHandler.RegisterSubmit},
		{"GET", "/logout", app.AuthHandler.Logout},
		{"GET", "/dashboard", app.DashboardHandler.DashboardPage},
		{"GET", "/quiz/{id}", app.QuizHandler.QuizStart},
		{"GET", "/quiz/{id}/q/{index}", app.QuizHandler.QuizQuestion},
		{"POST", "/quiz/{id}/navigate", app.QuizHandler.QuizNavigate},
		{"POST", "/quiz/{id}/finish", app.QuizHandler.QuizFinish},
		{"POST", "/quiz/{id}/cancel-session", app.QuizHandler.CancelSession},
		{"GET", "/quiz/{id}/result", app.QuizHandler.QuizResult},
		{"GET", "/errors", app.QuizHandler.ErrorsPage},
		{"GET", "/leaderboard", app.QuizHandler.LeaderboardPage},
		{"GET", "/admin", app.AdminHandler.Dashboard},
		{"GET", "/admin/quizzes", app.AdminHandler.Quizzes},
		{"GET", "/admin/quizzes/new", app.AdminHandler.QuizzesNew},
		{"GET", "/admin/quizzes/{id}", app.AdminHandler.QuizView},
		{"PUT", "/admin/quizzes/{id}", app.AdminHandler.QuizUpdate},
		{"DELETE", "/admin/quizzes/{id}", app.AdminHandler.QuizDelete},
		{"POST", "/admin/quizzes/{id}/question", app.AdminHandler.AddQuestion},
		{"PUT", "/admin/quizzes/{id}/question/{qid}", app.AdminHandler.UpdateQuestion},
		{"DELETE", "/admin/quizzes/{id}/question/{qid}", app.AdminHandler.DeleteQuestion},
		{"POST", "/admin/quizzes/{id}/question/{qid}/image", app.AdminHandler.UploadQuestionImage},
		{"DELETE", "/admin/quizzes/{id}/question/{qid}/image/{imgid}", app.AdminHandler.DeleteQuestionImage},
		{"POST", "/admin/quizzes/{id}/restore", app.AdminHandler.RestoreQuiz},
		{"GET", "/admin/results", app.AdminHandler.Results},
		{"GET", "/admin/statistics", app.AdminHandler.Statistics},
		{"GET", "/admin/api/quiz-stats", app.AdminHandler.QuizStatsData},
		{"GET", "/admin/api/grade-dist", app.AdminHandler.GradeDistData},
		{"GET", "/admin/api/subject-dist", app.AdminHandler.SubjectDistData},
	}

	requireAuthMiddleware := app.RequireAuthMiddleware.Wrap
	requiredRoleMiddleware := app.RequireRoleMiddleware.Wrap
	compressionMiddleware := app.CompressionMiddleware.Wrap
	commonHeadersMiddleware := app.CommonHeadersMiddleware.Wrap
	localeMiddleware := app.LocaleMiddleware.Wrap

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

	log.Printf("Server starting on http://localhost:%s", app.Config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+app.Config.ServerPort, mux))
}
