package middleware

import (
	"net/http"
	"slices"

	"github.com/goquizvibe/backend/shared/models"
	"github.com/goquizvibe/backend/feature/auth/services"
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
		cookie, err := r.Cookie(cookieNameToken)
		if err != nil {
			HandleAuthFailure(w, r)
			return
		}
		claims, err := m.authService.ValidateToken(cookie.Value)
		if err != nil {
			HandleAuthFailure(w, r)
			return
		}
		if slices.Contains(m.roles, claims.Role) {
			next.ServeHTTP(w, r)
			return
		}
		HandleAuthFailure(w, r)
	})
}
