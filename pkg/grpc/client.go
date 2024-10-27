package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
	"gophkeeper.com/pkg/grpc/middleware"
)

type GophkeeperClient struct {
	conn *grpc.ClientConn

	pb.GophkeeperServiceClient
}

func (dc *GophkeeperClient) Close() error {
	return dc.conn.Close()
}

func NewGophkeeperClient(targetURL string, token string) (*GophkeeperClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middleware.AuthInterceptor(token)),
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
