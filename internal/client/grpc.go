package client

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

type GophkeeperClient struct {
	conn *grpc.ClientConn

	pb.GophkeeperServiceClient
}

func (dc *GophkeeperClient) Close() {
	dc.conn.Close()
}

func NewGophkeeperClient(targetURL string) (*GophkeeperClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
