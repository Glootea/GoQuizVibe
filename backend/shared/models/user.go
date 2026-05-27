package models

import "github.com/google/uuid"

type AuthClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   Role      `json:"role"`
}