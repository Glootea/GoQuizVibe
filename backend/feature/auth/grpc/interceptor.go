package grpc

import (
	"context"

	"github.com/goquizvibe/backend/feature/auth/services"
	"github.com/goquizvibe/backend/shared/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type claimsContextKey struct{}

func ClaimsContextKey() claimsContextKey { return claimsContextKey{} }

func ClaimsFromContext(ctx context.Context) (*models.AuthClaims, bool) {
	v, ok := ctx.Value(claimsContextKey{}).(*models.AuthClaims)
	return v, ok && v != nil
}

func AuthInterceptor(authService *services.AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublicAuthMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token, err := extractBearerToken(md)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, claimsContextKey{}, claims)
		return handler(ctx, req)
	}
}

func isPublicAuthMethod(fullMethod string) bool {
	switch fullMethod {
	case "/auth.Auth/Register", "/auth.Auth/Login", "/auth.Auth/Logout":
		return true
	default:
		return false
	}
}
