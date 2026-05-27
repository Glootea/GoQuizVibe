package middleware

import (
	"log"
	"net/http"

	ce "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/locales"

	"github.com/goquizvibe/backend/shared/types"
	"github.com/goquizvibe/backend/shared/ui"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func ErrorHandler(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {

			log.Printf("--- ERROR START ---\n%v\nRequest url: %s\nRequest method: %s\n--- ERROR END ---", err, r.URL.String(), r.Method)

			isHTMX := IsHTMXRequest(r)
			status := ce.HTTPStatus(err)
			redirectTo := redirectPathForStatus(status)

			if isHTMX {
				w.WriteHeader(status)
				t := locales.GetTranslator(r.Context())
				ui.ErrorAlert(ce.UserMessage(err), redirectTo, t).Render(r.Context(), w)
			} else {
				w.WriteHeader(status)
				t := locales.GetTranslator(r.Context())
				ui.ErrorPage(types.ErrorData{
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
