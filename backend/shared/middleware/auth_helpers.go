package middleware

import (
	"net/http"
)

const (
	RedirectToLogin  = "/login"
	cookieNameToken = "token"
)

func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func HandleAuthFailure(w http.ResponseWriter, r *http.Request) {
	if IsHTMXRequest(r) {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, RedirectToLogin, http.StatusFound)
}