package middleware

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"gophkeeper.com/internal/client/jwt"
)

func AuthInterceptor(tokenProvider *jwt.TokenProvider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Skip auth for login/register
		if method == "/api.v1.GophkeeperService/Login" || method == "/api.v1.GophkeeperService/Register" {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		tokenData, err := tokenProvider.LoadToken()
		if err != nil {
			return fmt.Errorf("failed to load token: %w", err)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+tokenData.AccessToken)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamAuthInterceptor(tokenProvider *jwt.TokenProvider) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		tokenData, err := tokenProvider.LoadToken()
		if err != nil {
			return nil, fmt.Errorf("failed to load token: %w", err)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+tokenData.AccessToken)

		return streamer(ctx, desc, cc, method, opts...)
	}
}
