package middleware

import (
	"net/http"

	"github.com/goquizvibe/services"
)

func NewRequireAuthMiddleware(authService *services.AuthService) RequireAuthMiddleware {
	return RequireAuthMiddleware{authService}
}

type RequireAuthMiddleware struct {
	authService *services.AuthService
}

func (m RequireAuthMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		_, err = m.authService.ValidateToken(cookie.Value)
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
