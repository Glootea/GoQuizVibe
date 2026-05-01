package handlers

import (
	"net/http"

	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/store"
	"github.com/goquizvibe/types"
)

type AuthHandler struct {
	store       *store.MemoryStore
	authService *services.AuthService
}

func NewAuth(s *store.MemoryStore, a *services.AuthService) *AuthHandler {
	return &AuthHandler{store: s, authService: a}
}

func (h *AuthHandler) LandingPage(w http.ResponseWriter, r *http.Request) {
	pages.LandingPage().Render(r.Context(), w)
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.LoginPage(nil).Render(r.Context(), w)
}

func (h *AuthHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.authService.Login(email, password)
	if err != nil {
		pages.LoginPage(&types.LoginError{Message: "Неверный email или пароль"}).Render(r.Context(), w)
		return
	}

	token, _ := h.authService.GenerateToken(user)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	pages.RegisterPage(nil).Render(r.Context(), w)
}

func (h *AuthHandler) RegisterSubmit(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(password) < 6 {
		pages.RegisterPage(&types.RegisterError{Message: "Пароль должен быть минимум 6 символов"}).Render(r.Context(), w)
		return
	}

	user, err := h.authService.Register(name, email, password, models.RoleStudent)
	if err != nil {
		pages.RegisterPage(&types.RegisterError{Message: "Email уже зарегистрирован"}).Render(r.Context(), w)
		return
	}

	token, _ := h.authService.GenerateToken(user)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		_, err = h.authService.ValidateToken(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
