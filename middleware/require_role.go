package middleware

import (
	"net/http"
	"slices"

	"github.com/goquizvibe/models"
	"github.com/goquizvibe/services"
)

func NewRequireRoleMiddleware(authService *services.AuthService, roles ...models.Role) RequireRoleMiddleware {
	return RequireRoleMiddleware{authService, roles}
}

type RequireRoleMiddleware struct {
	authService *services.AuthService
	roles       []models.Role
}

func (m RequireRoleMiddleware) Wrap(next http.Handler) http.HandlerFunc {
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
		claims, err := m.authService.ValidateToken(cookie.Value)
		if err != nil {
			if r.Header.Get("hx-request") == "true" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if slices.Contains(m.roles, claims.Role) {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("hx-request") == "true" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})

}
