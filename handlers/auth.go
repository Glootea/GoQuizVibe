package handlers

import (
	"errors"
	"net/http"
	"slices"

	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"
)

type AuthHandler struct {
	pool        *db.Queries
	authService *services.AuthService
}

func NewAuth(pool *db.Queries, a *services.AuthService) *AuthHandler {
	return &AuthHandler{pool: pool, authService: a}
}

func (h *AuthHandler) LandingPage(w http.ResponseWriter, r *http.Request) error {
	return pages.LandingPage().Render(r.Context(), w)

}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) error {
	return pages.LoginPage(nil).Render(r.Context(), w)
}

func (h *AuthHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.authService.Login(r.Context(), email, password)
	if err != nil {
		return pages.LoginPage(&types.LoginError{Message: "Неверный email или пароль"}).Render(r.Context(), w)
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
	return pages.RegisterPage(nil).Render(r.Context(), w)
}

func (h *AuthHandler) RegisterSubmit(w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(password) < 6 {
		return pages.RegisterPage(&types.RegisterError{Message: "Пароль должен быть минимум 6 символов"}).Render(r.Context(), w)
	}

	user, err := h.authService.Register(r.Context(), name, email, password, models.RoleStudent)
	if err != nil {
		return pages.RegisterPage(&types.RegisterError{Message: "Email уже зарегистрирован"}).Render(r.Context(), w)
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

func (h *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if r.Header.Get("hx-request") == "true" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		_, err = h.authService.ValidateToken(cookie.Value)
		if err != nil {
			if r.Header.Get("hx-request") == "true" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AuthHandler) RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				if r.Header.Get("hx-request") == "true" {
					http.NotFound(w, r)
					return
				}
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			claims, err := h.authService.ValidateToken(cookie.Value)
			if err != nil {
				if r.Header.Get("hx-request") == "true" {
					http.NotFound(w, r)
					return
				}
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			if slices.Contains(roles, claims.Role) {
				next.ServeHTTP(w, r)
				return
			}
			if r.Header.Get("hx-request") == "true" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		})
	}
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
