package middleware

import "net/http"

func NewCommonHeaders() *CommonHeaders {
	return &CommonHeaders{}
}

type CommonHeaders struct {
}

func (c *CommonHeaders) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	}
}
