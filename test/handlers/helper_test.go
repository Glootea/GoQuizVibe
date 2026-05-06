package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/handlers"
	mocks "github.com/goquizvibe/mocks/services"
	"github.com/goquizvibe/models"
	"go.uber.org/mock/gomock"
)

func TestGetUserIDFromCookie(t *testing.T) {
	t.Run("valid token returns user ID", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuthenticator(ctrl)
		expectedUserID := uuid.New()

		mockAuth.EXPECT().ValidateToken("valid-token").Return(&models.AuthClaims{
			UserID: expectedUserID,
			Email:  "test@example.com",
			Role:   db.RoleStudent,
		}, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})

		userID, err := handlers.GetUserIDFromCookie(req, mockAuth)
		if err != nil {
			t.Fatalf("GetUserIDFromCookie() error = %v, want nil", err)
		}
		if userID != expectedUserID {
			t.Errorf("GetUserIDFromCookie() = %v, want %v", userID, expectedUserID)
		}
	})

	t.Run("missing cookie returns error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuthenticator(ctrl)

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, err := handlers.GetUserIDFromCookie(req, mockAuth)
		if err == nil {
			t.Fatal("GetUserIDFromCookie() error = nil, want error")
		}
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuthenticator(ctrl)

		mockAuth.EXPECT().ValidateToken("invalid-token").Return(nil, custerrors.ErrUnauthorized)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})

		_, err := handlers.GetUserIDFromCookie(req, mockAuth)
		if err == nil {
			t.Fatal("GetUserIDFromCookie() error = nil, want error")
		}
	})
}
