package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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

			if isJSONRequest(r) {
				WriteJSONError(w, r, err)
				return
			}

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

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

func isJSONRequest(r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		return true
	}
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	return strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html")
}

func WriteJSONError(w http.ResponseWriter, r *http.Request, err error) {
	status := ce.HTTPStatus(err)
	envelope := errorEnvelope{
		Error: errorBody{
			Code:    codeForStatus(status),
			Message: ce.UserMessage(err),
		},
	}
	if fields, ok := err.(interface{ Fields() map[string]string }); ok {
		envelope.Error.Fields = fields.Fields()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusInternalServerError:
		return "INTERNAL"
	default:
		return "ERROR"
	}
}
