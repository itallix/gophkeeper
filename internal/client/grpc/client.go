package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/itallix/gophkeeper/internal/client/grpc/middleware"
	"github.com/itallix/gophkeeper/internal/client/jwt"
	pb "github.com/itallix/gophkeeper/pkg/generated/api/proto/v1"
)

type GophkeeperClient struct {
	conn *grpc.ClientConn

	pb.GophkeeperServiceClient
}

func (dc *GophkeeperClient) Close() error {
	return dc.conn.Close()
}

func NewGophkeeperClient(targetURL string, tokenProvider *jwt.TokenProvider) (*GophkeeperClient, error) {
	// TODO: implement tokenProvider with refresh tokens
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middleware.AuthInterceptor(tokenProvider)),
		grpc.WithStreamInterceptor(middleware.StreamAuthInterceptor(tokenProvider)),
	}
	conn, err := grpc.NewClient(targetURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new client: %w", err)
	}

	return &GophkeeperClient{
		conn:                    conn,
		GophkeeperServiceClient: pb.NewGophkeeperServiceClient(conn),
	}, err
}
