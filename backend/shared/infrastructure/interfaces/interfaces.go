package interfaces

import (
	"context"

	"github.com/goquizvibe/backend/shared/models"
)

type Authenticator interface {
	ValidateToken(token string) (*models.AuthClaims, error)
}

type CacheServiceInterface interface {
	Get(ctx context.Context, key string, dest any) bool
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
