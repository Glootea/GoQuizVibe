package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/services"
)

func GetUserIDFromCookie(r *http.Request, auth services.Authenticator) (uuid.UUID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := auth.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}
