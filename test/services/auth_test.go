package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	mocks "github.com/goquizvibe/mocks/services"
	"github.com/goquizvibe/services"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	exp := time.Hour * 24

	t.Run("valid registration", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		mockUsers.EXPECT().CreateUser(ctx, gomock.Any()).Return(db.User{
			ID:    uuid.New(),
			Name:  "Test User",
			Email: "test@example.com",
			Role:  db.RoleStudent,
		}, nil)

		user, err := svc.Register(ctx, "Test User", "test@example.com", "password123", db.RoleStudent)
		if err != nil {
			t.Fatalf("Register() error = %v, want nil", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Register() email = %v, want %v", user.Email, "test@example.com")
		}
	})

	t.Run("database error on create", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		mockUsers.EXPECT().CreateUser(ctx, gomock.Any()).Return(db.User{}, errors.New("database error"))

		_, err := svc.Register(ctx, "Test", "test@example.com", "password123", db.RoleStudent)
		if err == nil {
			t.Fatal("Register() error = nil, want error")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret"
	exp := time.Hour * 24

	t.Run("valid login", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		mockUsers.EXPECT().GetUserByEmail(ctx, "test@example.com").Return(db.User{
			ID:           uuid.New(),
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         db.RoleStudent,
		}, nil)

		user, err := svc.Login(ctx, "test@example.com", "correctpassword")
		if err != nil {
			t.Fatalf("Login() error = %v, want nil", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Login() email = %v, want %v", user.Email, "test@example.com")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		mockUsers.EXPECT().GetUserByEmail(ctx, "notfound@example.com").Return(db.User{}, errors.New("sql: no rows"))

		_, err := svc.Login(ctx, "notfound@example.com", "password123")
		if err == nil {
			t.Fatal("Login() error = nil, want error")
		}
		if err.Error() != "invalid credentials" {
			t.Errorf("Login() error message = %v, want %v", err.Error(), "invalid credentials")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		mockUsers.EXPECT().GetUserByEmail(ctx, "test@example.com").Return(db.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         db.RoleStudent,
		}, nil)

		_, err := svc.Login(ctx, "test@example.com", "wrongpassword")
		if err == nil {
			t.Fatal("Login() error = nil, want error")
		}
		if err.Error() != "invalid credentials" {
			t.Errorf("Login() error message = %v, want %v", err.Error(), "invalid credentials")
		}
	})
}

func TestAuthService_GenerateToken(t *testing.T) {
	secret := "test-secret"
	exp := time.Hour * 24

	t.Run("valid token generation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		user := &db.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Role:  db.RoleStudent,
		}

		token, err := svc.GenerateToken(user)
		if err != nil {
			t.Fatalf("GenerateToken() error = %v, want nil", err)
		}
		if token == "" {
			t.Fatal("GenerateToken() returned empty token")
		}

		claims, err := svc.ValidateToken(token)
		if err != nil {
			t.Fatalf("ValidateToken() error = %v, want nil", err)
		}
		if claims.UserID != user.ID {
			t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, user.ID)
		}
		if claims.Email != user.Email {
			t.Errorf("ValidateToken() email = %v, want %v", claims.Email, user.Email)
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	secret := "test-secret"
	exp := time.Hour * 24

	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		_, err := svc.ValidateToken("invalid-token")
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, -time.Hour)

		user := &db.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Role:  db.RoleStudent,
		}

		token, _ := svc.GenerateToken(user)

		_, err := svc.ValidateToken(token)
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error for expired token")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers1 := mocks.NewMockUserRepository(ctrl)
		mockUsers2 := mocks.NewMockUserRepository(ctrl)
		svc1 := services.NewAuthService(mockUsers1, "secret1", exp)
		svc2 := services.NewAuthService(mockUsers2, "secret2", exp)

		user := &db.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Role:  db.RoleStudent,
		}

		token, _ := svc1.GenerateToken(user)

		_, err := svc2.ValidateToken(token)
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error for wrong secret")
		}
	})

	t.Run("missing user_id claim", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		tokenStr, _ := svc.GenerateToken(&db.User{ID: uuid.New(), Email: "test@example.com", Role: db.RoleStudent})
		token, _ := jwt.Parse(tokenStr, nil)
		claims := token.Claims.(jwt.MapClaims)
		delete(claims, "user_id")

		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ = tokenObj.SignedString([]byte(secret))
		_, err := svc.ValidateToken(tokenStr)
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error for missing user_id")
		}
	})

	t.Run("invalid UUID in user_id claim", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		tokenStr, _ := svc.GenerateToken(&db.User{ID: uuid.New(), Email: "test@example.com", Role: db.RoleStudent})
		token, _ := jwt.Parse(tokenStr, nil)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = "not-a-valid-uuid"

		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ = tokenObj.SignedString([]byte(secret))
		_, err := svc.ValidateToken(tokenStr)
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error for invalid UUID")
		}
	})

	t.Run("missing email claim", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		tokenStr, _ := svc.GenerateToken(&db.User{ID: uuid.New(), Email: "test@example.com", Role: db.RoleStudent})
		token, _ := jwt.Parse(tokenStr, nil)
		claims := token.Claims.(jwt.MapClaims)
		delete(claims, "email")

		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ = tokenObj.SignedString([]byte(secret))
		_, err := svc.ValidateToken(tokenStr)
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error for missing email")
		}
	})

	t.Run("missing role claim", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		svc := services.NewAuthService(mockUsers, secret, exp)

		tokenStr, _ := svc.GenerateToken(&db.User{ID: uuid.New(), Email: "test@example.com", Role: db.RoleStudent})
		token, _ := jwt.Parse(tokenStr, nil)
		claims := token.Claims.(jwt.MapClaims)
		delete(claims, "role")

		tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ = tokenObj.SignedString([]byte(secret))
		_, err := svc.ValidateToken(tokenStr)
		if err == nil {
			t.Fatal("ValidateToken() error = nil, want error for missing role")
		}
	})
}
