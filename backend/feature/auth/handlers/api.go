package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	custerrors "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/middleware"
	"github.com/goquizvibe/backend/shared/models"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AuthResponse struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
	ExpiresAt   string  `json:"expires_at"`
}

type MeResponse struct {
	User UserDTO `json:"user"`
}

func toUserDTO(u *db.User) UserDTO {
	return UserDTO{
		ID:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
		Role:  string(u.Role),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func (h *AuthHandler) RegisterJSON(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrInvalidRequest, err),
			http.StatusBadRequest,
		)
	}

	if len(req.Password) < 6 {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrInvalidRequest, errors.New("password must be at least 6 characters")),
			http.StatusBadRequest,
		)
	}

	user, err := h.authService.Register(r.Context(), req.Name, req.Email, req.Password, models.RoleStudent)
	if err != nil {
		if isEmailExists(err) {
			return custerrors.WithHTTPStatus(
				errors.Join(custerrors.ErrInvalidRequest, err),
				http.StatusConflict,
			)
		}
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrInternal, err),
			http.StatusInternalServerError,
		)
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrInternal, err),
			http.StatusInternalServerError,
		)
	}

	setTokenCookie(w, token)
	writeJSON(w, http.StatusCreated, AuthResponse{
		User:        toUserDTO(user),
		AccessToken: token,
		ExpiresAt:   "0",
	})
	return nil
}

func (h *AuthHandler) LoginJSON(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrInvalidRequest, err),
			http.StatusBadRequest,
		)
	}

	user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrUnauthorized, err),
			http.StatusUnauthorized,
		)
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrInternal, err),
			http.StatusInternalServerError,
		)
	}

	setTokenCookie(w, token)
	writeJSON(w, http.StatusOK, AuthResponse{
		User:        toUserDTO(user),
		AccessToken: token,
		ExpiresAt:   "0",
	})
	return nil
}

func (h *AuthHandler) LogoutJSON(w http.ResponseWriter, r *http.Request) error {
	middleware.ClearTokenCookie(w)
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

func (h *AuthHandler) MeJSON(w http.ResponseWriter, r *http.Request) error {
	user, err := h.GetUserFromRequest(r)
	if err != nil {
		return custerrors.WithHTTPStatus(
			errors.Join(custerrors.ErrUnauthorized, err),
			http.StatusUnauthorized,
		)
	}
	writeJSON(w, http.StatusOK, MeResponse{User: toUserDTO(user)})
	return nil
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieNameToken,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func isEmailExists(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique") ||
		strings.Contains(msg, "users_email_key")
}
