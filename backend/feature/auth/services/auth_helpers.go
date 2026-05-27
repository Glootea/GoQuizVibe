package services

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
)

const cookieNameToken = "token"

func GetUserIDFromRequest(r *http.Request, auth interfaces.Authenticator) (uuid.UUID, error) {
	cookie, err := r.Cookie(cookieNameToken)
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := auth.ValidateToken(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}