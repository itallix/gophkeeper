package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"gophkeeper.com/internal/server/service"
)

type AuthInterceptor struct {
	authService service.AuthenticationService
	// List of methods that don't require auth
	noAuthMethods map[string]bool
}

func NewAuthInterceptor(authService service.AuthenticationService) *AuthInterceptor {
	noAuthMethods := map[string]bool{
		"/api.v1.GophkeeperService/Register": true,
		"/api.v1.GophkeeperService/Login":    true,
	}

	return &AuthInterceptor{
		authService:   authService,
		noAuthMethods: noAuthMethods,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for whitelisted methods
		if i.noAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := i.extractToken(ctx)
		if err != nil {
			return nil, err
		}

		// Validate token
		username, err := i.authService.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add claims to context
		newCtx := context.WithValue(ctx, "username", username)

		return handler(newCtx, req)
	}
}

func (i *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}
