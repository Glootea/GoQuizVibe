package custerrors

import (
	"errors"
	"net/http"
)

type statusError struct {
	err    error
	status int
}

func (e statusError) Error() string   { return e.err.Error() }
func (e statusError) HTTPStatus() int { return e.status }
func (e statusError) Unwrap() error   { return e.err }

func WithHTTPStatus(err error, status int) error {
	if err == nil {
		return nil
	}
	return statusError{err: err, status: status}
}

func HTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	var se interface{ HTTPStatus() int }
	if errors.As(err, &se) {
		return se.HTTPStatus()
	}
	return http.StatusInternalServerError
}

func UserMessage(err error) string {
	if err == nil {
		return ""
	}
	var se interface{ UserMessage() string }
	if errors.As(err, &se) {
		return se.UserMessage()
	}
	return "Что-то пошло не так. Попробуйте позже."
}

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not found")
	ErrInternal       = errors.New("internal error")
)
