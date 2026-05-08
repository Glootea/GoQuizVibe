package handlers

import (
	"log"
	"net/http"

	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/types"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func ErrorHandler(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {

			log.Printf("--- ERROR START ---\n%v\nRequest url: %s\nRequest method: %s\n--- ERROR END ---", err, r.URL.String(), r.Method)

			isHTMX := r.Header.Get("hx-request") == "true"
			status := ce.HTTPStatus(err)
			redirectTo := redirectPathForStatus(status)

			if isHTMX {
				w.WriteHeader(status)
				t := middleware.GetTranslator(r.Context())
				pages.ErrorAlert(ce.UserMessage(err), redirectTo, t).Render(r.Context(), w)
			} else {
				w.WriteHeader(status)
				t := middleware.GetTranslator(r.Context())
				pages.ErrorPage(types.ErrorData{
					Message:    ce.UserMessage(err),
					RedirectTo: redirectTo,
				}, t).Render(r.Context(), w)
			}
		}
	}
}

func ErrorHandlerFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return ErrorHandler(f)
}

func redirectPathForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "/login"
	case http.StatusForbidden, http.StatusNotFound:
		return "/dashboard"
	default:
		return "/dashboard"
	}
}
