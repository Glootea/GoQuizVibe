package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     r.UserRepository
	jwtSecret []byte
	jwtExp    time.Duration
}

func NewAuthService(users r.UserRepository, secret string, exp time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(secret),
		jwtExp:    exp,
	}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string, role db.Role) (*db.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.users.CreateUser(ctx, db.CreateUserParams{
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
	user, err := s.users.GetUserByEmail(ctx, email)
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
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("missing user_id claim")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user_id in token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("missing email claim")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("missing role claim")
	}

	return &models.AuthClaims{
		UserID: userID,
		Email:  email,
		Role:   db.Role(role),
	}, nil
}
