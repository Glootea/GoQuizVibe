package services

import "github.com/goquizvibe/models"

type Authenticator interface {
	ValidateToken(token string) (*models.AuthClaims, error)
}
