package services_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	authgrpc "github.com/goquizvibe/backend/feature/auth/grpc"
	"github.com/goquizvibe/backend/feature/auth/services"
	"github.com/goquizvibe/backend/shared/db"
	authproto "github.com/goquizvibe/backend/shared/grpc/proto"
	mocks "github.com/goquizvibe/backend/shared/mocks/services"
	"github.com/goquizvibe/backend/shared/models"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const (
	testJWTSecret = "test-secret-for-grpc"
	testJWTExp    = time.Hour
)

type grpcTestEnv struct {
	client  authproto.AuthClient
	cleanup func()
}

func newGRPCTestEnv(t *testing.T, mockUsers *mocks.MockUserRepository) *grpcTestEnv {
	t.Helper()

	authSvc := services.NewAuthService(mockUsers, testJWTSecret, testJWTExp)
	srv := authgrpc.NewAuthServer(authSvc, mockUsers)
	interceptor := authgrpc.AuthInterceptor(authSvc)

	lis := bufconn.Listen(1024 * 64)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	authproto.RegisterAuthServer(grpcServer, srv)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		grpcServer.Stop()
		t.Fatalf("grpc.NewClient: %v", err)
	}

	client := authproto.NewAuthClient(conn)
	return &grpcTestEnv{
		client: client,
		cleanup: func() {
			conn.Close()
			grpcServer.Stop()
			_ = lis.Close()
			<-serveErr
		},
	}
}

func newTestUser(t *testing.T, email, password string, role db.Role) db.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return db.User{
		ID:           uuid.New(),
		Name:         "Test User",
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}
}

func TestAuthGRPC_Register(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		user := newTestUser(t, "new@example.com", "secret", db.RoleStudent)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockUsers.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(user, nil)

		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		resp, err := env.client.Register(context.Background(), &authproto.RegisterRequest{
			Name:     user.Name,
			Email:    user.Email,
			Password: "secret",
		})
		if err != nil {
			t.Fatalf("Register error = %v", err)
		}
		if resp.GetUser().GetEmail() != user.Email {
			t.Errorf("Register email = %q, want %q", resp.GetUser().GetEmail(), user.Email)
		}
		if resp.GetAccessToken() == "" {
			t.Error("Register access_token must not be empty")
		}
	})

	t.Run("short password", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		_, err := env.client.Register(context.Background(), &authproto.RegisterRequest{
			Name:     "x",
			Email:    "x@y.z",
			Password: "12345",
		})
		if err == nil {
			t.Fatal("Register should fail on short password")
		}
	})

	t.Run("email conflict", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockUsers.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(db.User{}, errors.New("ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"))

		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		_, err := env.client.Register(context.Background(), &authproto.RegisterRequest{
			Name:     "x",
			Email:    "dup@y.z",
			Password: "secret",
		})
		if err == nil {
			t.Fatal("Register should fail on duplicate email")
		}
	})
}

func TestAuthGRPC_Login(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		user := newTestUser(t, "ok@example.com", "secret", db.RoleStudent)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockUsers.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)

		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		resp, err := env.client.Login(context.Background(), &authproto.LoginRequest{
			Email:    user.Email,
			Password: "secret",
		})
		if err != nil {
			t.Fatalf("Login error = %v", err)
		}
		if resp.GetUser().GetEmail() != user.Email {
			t.Errorf("Login email = %q, want %q", resp.GetUser().GetEmail(), user.Email)
		}
		if resp.GetAccessToken() == "" {
			t.Error("Login access_token must not be empty")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		user := newTestUser(t, "bad@example.com", "secret", db.RoleStudent)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockUsers.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)

		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		_, err := env.client.Login(context.Background(), &authproto.LoginRequest{
			Email:    user.Email,
			Password: "WRONG",
		})
		if err == nil {
			t.Fatal("Login should fail on wrong password")
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		_, err := env.client.Login(context.Background(), &authproto.LoginRequest{
			Email:    "",
			Password: "",
		})
		if err == nil {
			t.Fatal("Login should fail when fields are empty")
		}
	})
}

func TestAuthGRPC_Logout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsers := mocks.NewMockUserRepository(ctrl)
	env := newGRPCTestEnv(t, mockUsers)
	defer env.cleanup()

	resp, err := env.client.Logout(context.Background(), &authproto.LogoutRequest{})
	if err != nil {
		t.Fatalf("Logout error = %v", err)
	}
	if resp == nil {
		t.Fatal("Logout response must not be nil")
	}
}

func TestAuthGRPC_Me(t *testing.T) {
	t.Parallel()

	t.Run("authenticated", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		user := newTestUser(t, "me@example.com", "secret", db.RoleStudent)
		mockUsers := mocks.NewMockUserRepository(ctrl)
		mockUsers.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)

		authSvc := services.NewAuthService(mockUsers, testJWTSecret, testJWTExp)
		token, err := authSvc.GenerateToken(&user)
		if err != nil {
			t.Fatalf("GenerateToken: %v", err)
		}

		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
		resp, err := env.client.Me(ctx, &authproto.MeRequest{})
		if err != nil {
			t.Fatalf("Me error = %v", err)
		}
		if resp.GetEmail() != user.Email {
			t.Errorf("Me email = %q, want %q", resp.GetEmail(), user.Email)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		_, err := env.client.Me(context.Background(), &authproto.MeRequest{})
		if err == nil {
			t.Fatal("Me should fail without auth metadata")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer not-a-jwt")
		_, err := env.client.Me(ctx, &authproto.MeRequest{})
		if err == nil {
			t.Fatal("Me should fail with invalid token")
		}
	})

	t.Run("malformed auth header", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUsers := mocks.NewMockUserRepository(ctrl)
		env := newGRPCTestEnv(t, mockUsers)
		defer env.cleanup()

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Basic abc")
		_, err := env.client.Me(ctx, &authproto.MeRequest{})
		if err == nil {
			t.Fatal("Me should fail with non-Bearer scheme")
		}
	})
}

func TestClaimsContextRoundTrip(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	claims := &models.AuthClaims{
		UserID: uid,
		Email:  "ctx@example.com",
		Role:   db.RoleStudent,
	}

	ctx := context.WithValue(context.Background(), authgrpc.ClaimsContextKey(), claims)
	got, ok := authgrpc.ClaimsFromContext(ctx)
	if !ok {
		t.Fatal("ClaimsFromContext should return true for context with claims")
	}
	if got.UserID != uid {
		t.Errorf("UserID mismatch: got %v want %v", got.UserID, uid)
	}

	_, ok = authgrpc.ClaimsFromContext(context.Background())
	if ok {
		t.Fatal("ClaimsFromContext should return false for empty context")
	}
}
