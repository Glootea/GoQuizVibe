package handlers

import (
	"errors"
	"net/http"

	"github.com/goquizvibe/backend/feature/auth/services"
	"github.com/goquizvibe/backend/feature/auth/ui"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/locales"
	"github.com/goquizvibe/backend/shared/models"
	"github.com/goquizvibe/backend/shared/types"
)

type AuthHandler struct {
	pool        *db.Queries
	authService *services.AuthService
	localeSvc   *locales.Service
}

func NewAuth(pool *db.Queries, a *services.AuthService, svc *locales.Service) *AuthHandler {
	return &AuthHandler{pool: pool, authService: a, localeSvc: svc}
}

func (h *AuthHandler) LandingPage(w http.ResponseWriter, r *http.Request) error {
	t := locales.GetTranslator(r.Context())
	return ui.LandingPage(t).Render(r.Context(), w)
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) error {
	t := locales.GetTranslator(r.Context())
	return ui.LoginPage(nil, t).Render(r.Context(), w)
}

func (h *AuthHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.authService.Login(r.Context(), email, password)
	if err != nil {
		t := locales.GetTranslator(r.Context())
		return ui.LoginPage(&types.LoginError{Message: "Invalid email or password"}, t).Render(r.Context(), w)
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

	if user.Role == models.RoleTeacher {
		http.Redirect(w, r, "/admin", http.StatusFound)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
	return nil
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) error {
	t := locales.GetTranslator(r.Context())
	return ui.RegisterPage(nil, t).Render(r.Context(), w)
}

func (h *AuthHandler) RegisterSubmit(w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(password) < 6 {
		t := locales.GetTranslator(r.Context())
		return ui.RegisterPage(&types.RegisterError{Message: "Password must be at least 6 characters"}, t).Render(r.Context(), w)
	}

	user, err := h.authService.Register(r.Context(), name, email, password, models.RoleStudent)
	if err != nil {
		t := locales.GetTranslator(r.Context())
		return ui.RegisterPage(&types.RegisterError{Message: "Email already registered"}, t).Render(r.Context(), w)
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
	return nil
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

func (h *AuthHandler) GetUserFromRequest(r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, errors.Join(errors.New("get cookie"), err)
	}
	claims, err := h.authService.ValidateToken(cookie.Value)
	if err != nil {
		return nil, errors.Join(errors.New("validate token"), err)
	}
	user, err := h.pool.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		return nil, errors.Join(errors.New("get user"), err)
	}
	return &user, nil
}
