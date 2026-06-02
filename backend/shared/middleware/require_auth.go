package middleware

import (
	"net/http"

	"github.com/goquizvibe/backend/feature/auth/services"
)

func NewRequireAuthMiddleware(authService *services.AuthService) RequireAuthMiddleware {
	return RequireAuthMiddleware{authService}
}

type RequireAuthMiddleware struct {
	authService *services.AuthService
}

func (m RequireAuthMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieNameToken)
		if err != nil {
			HandleAuthFailure(w, r)
			return
		}
		_, err = m.authService.ValidateToken(cookie.Value)
		if err != nil {
			HandleAuthFailure(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
