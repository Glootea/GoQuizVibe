package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/feature/auth/services"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = uuid.Nil

const (
	minPasswordLen = 6
	metadataAuth   = "authorization"
	bearerPrefix   = "Bearer "
)

type AuthServer struct {
	proto.UnimplementedAuthServer
	authService *services.AuthService
	queries     UserLookup
}

type UserLookup interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
}

func NewAuthServer(authService *services.AuthService, queries UserLookup) *AuthServer {
	return &AuthServer{authService: authService, queries: queries}
}

func (s *AuthServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.AuthResponse, error) {
	name := strings.TrimSpace(req.GetName())
	email := strings.TrimSpace(req.GetEmail())
	password := req.GetPassword()

	if name == "" || email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email and password are required")
	}
	if len(password) < minPasswordLen {
		return nil, status.Errorf(codes.InvalidArgument, "password must be at least %d characters", minPasswordLen)
	}

	user, err := s.authService.Register(ctx, name, email, password, db.RoleStudent)
	if err != nil {
		if isEmailConflict(err) {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
		return nil, status.Errorf(codes.Internal, "register user: %v", err)
	}

	return s.buildAuthResponse(ctx, user)
}

func (s *AuthServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	email := strings.TrimSpace(req.GetEmail())
	password := req.GetPassword()
	if email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.authService.Login(ctx, email, password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return s.buildAuthResponse(ctx, user)
}

func (s *AuthServer) Logout(_ context.Context, _ *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	return &proto.LogoutResponse{}, nil
}

func (s *AuthServer) Me(ctx context.Context, _ *proto.MeRequest) (*proto.User, error) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	user, err := s.queries.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not found: %v", err)
	}

	return userToProto(&user), nil
}

func (s *AuthServer) buildAuthResponse(ctx context.Context, user *db.User) (*proto.AuthResponse, error) {
	token, err := s.authService.GenerateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate token: %v", err)
	}

	return &proto.AuthResponse{
		User:        userToProto(user),
		AccessToken: token,
		ExpiresAt:   0,
	}, nil
}

func userToProto(u *db.User) *proto.User {
	if u == nil {
		return nil
	}
	return &proto.User{
		Id:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
		Role:  string(u.Role),
	}
}

func isEmailConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique") ||
		strings.Contains(msg, "users_email_key")
}

func extractBearerToken(md map[string][]string) (string, error) {
	values := md[metadataAuth]
	if len(values) == 0 {
		return "", errors.New("authorization metadata missing")
	}
	raw := values[0]
	if !strings.HasPrefix(raw, bearerPrefix) {
		return "", errors.New("authorization metadata must use Bearer scheme")
	}
	return strings.TrimSpace(strings.TrimPrefix(raw, bearerPrefix)), nil
}
