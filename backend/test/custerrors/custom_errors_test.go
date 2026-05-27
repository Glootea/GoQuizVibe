package custerrors_test

import (
	"errors"
	"net/http"
	"testing"

	ce "github.com/goquizvibe/backend/shared/custom_errors"
)

func TestWithHTTPStatus(t *testing.T) {
	t.Run("wraps error with status", func(t *testing.T) {
		t.Parallel()
		err := errors.New("test error")
		wrapped := ce.WithHTTPStatus(err, http.StatusBadRequest)

		se, ok := wrapped.(interface{ HTTPStatus() int })
		if !ok {
			t.Fatal("wrapped error does not implement HTTPStatus() int")
		}
		if se.HTTPStatus() != http.StatusBadRequest {
			t.Errorf("HTTPStatus() = %d, want %d", se.HTTPStatus(), http.StatusBadRequest)
		}
	})

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		wrapped := ce.WithHTTPStatus(nil, http.StatusBadRequest)
		if wrapped != nil {
			t.Errorf("WithHTTPStatus(nil) = %v, want nil", wrapped)
		}
	})
}

func TestHTTPStatus(t *testing.T) {
	t.Run("returns 200 for nil error", func(t *testing.T) {
		t.Parallel()
		status := ce.HTTPStatus(nil)
		if status != http.StatusOK {
			t.Errorf("HTTPStatus(nil) = %d, want %d", status, http.StatusOK)
		}
	})

	t.Run("returns status from custom error", func(t *testing.T) {
		t.Parallel()
		err := ce.WithHTTPStatus(errors.New("not found"), http.StatusNotFound)
		status := ce.HTTPStatus(err)
		if status != http.StatusNotFound {
			t.Errorf("HTTPStatus() = %d, want %d", status, http.StatusNotFound)
		}
	})

	t.Run("returns 500 for plain error", func(t *testing.T) {
		t.Parallel()
		err := errors.New("plain error")
		status := ce.HTTPStatus(err)
		if status != http.StatusInternalServerError {
			t.Errorf("HTTPStatus() = %d, want %d", status, http.StatusInternalServerError)
		}
	})
}

func TestUserMessage(t *testing.T) {
	t.Run("returns empty for nil error", func(t *testing.T) {
		t.Parallel()
		msg := ce.UserMessage(nil)
		if msg != "" {
			t.Errorf("UserMessage(nil) = %q, want empty", msg)
		}
	})

	t.Run("returns default message for error wrapped with WithHTTPStatus", func(t *testing.T) {
		t.Parallel()
		err := ce.WithHTTPStatus(errors.New("custom message"), http.StatusBadRequest)
		msg := ce.UserMessage(err)
		if msg != "Что-то пошло не так. Попробуйте позже." {
			t.Errorf("UserMessage() = %q, want default message", msg)
		}
	})

	t.Run("returns default message for plain error", func(t *testing.T) {
		t.Parallel()
		err := errors.New("plain error")
		msg := ce.UserMessage(err)
		if msg != "Что-то пошло не так. Попробуйте позже." {
			t.Errorf("UserMessage() = %q, want default message", msg)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Run("ErrInvalidRequest has correct message", func(t *testing.T) {
		t.Parallel()
		if ce.ErrInvalidRequest.Error() != "invalid request" {
			t.Errorf("ErrInvalidRequest.Error() = %q, want %q", ce.ErrInvalidRequest.Error(), "invalid request")
		}
	})

	t.Run("ErrUnauthorized has correct message", func(t *testing.T) {
		t.Parallel()
		if ce.ErrUnauthorized.Error() != "unauthorized" {
			t.Errorf("ErrUnauthorized.Error() = %q, want %q", ce.ErrUnauthorized.Error(), "unauthorized")
		}
	})

	t.Run("ErrForbidden has correct message", func(t *testing.T) {
		t.Parallel()
		if ce.ErrForbidden.Error() != "forbidden" {
			t.Errorf("ErrForbidden.Error() = %q, want %q", ce.ErrForbidden.Error(), "forbidden")
		}
	})

	t.Run("ErrNotFound has correct message", func(t *testing.T) {
		t.Parallel()
		if ce.ErrNotFound.Error() != "not found" {
			t.Errorf("ErrNotFound.Error() = %q, want %q", ce.ErrNotFound.Error(), "not found")
		}
	})

	t.Run("ErrInternal has correct message", func(t *testing.T) {
		t.Parallel()
		if ce.ErrInternal.Error() != "internal error" {
			t.Errorf("ErrInternal.Error() = %q, want %q", ce.ErrInternal.Error(), "internal error")
		}
	})
}
