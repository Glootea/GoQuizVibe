package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	pool         *db.Queries
	jwtSecret    []byte
	jwtExp       time.Duration
}

func NewAuthService(pool *db.Queries, secret string, exp time.Duration) *AuthService {
	return &AuthService{
		pool:         pool,
		jwtSecret:    []byte(secret),
		jwtExp:       exp,
	}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string, role db.Role) (*db.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.pool.CreateUser(ctx, db.CreateUserParams{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*db.User, error) {
	user, err := s.pool.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (s *AuthService) GenerateToken(user *db.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     time.Now().Add(s.jwtExp).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (*models.AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr := claims["user_id"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user_id in token")
	}

	return &models.AuthClaims{
		UserID: userID,
		Email:  claims["email"].(string),
		Role:   db.Role(claims["role"].(string)),
	}, nil
}