package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"gophkeeper.com/internal/common/logger"
	"gophkeeper.com/internal/server"
	pgrpc "gophkeeper.com/internal/server/grpc"
	"gophkeeper.com/internal/server/grpc/middleware"
	"gophkeeper.com/internal/server/s3"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

const (
	ServerPort           = "8081"
	AccessTokenTTLHours  = 1
	RefreshTokenTTLHours = 24
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}

	ctx := context.Background()
	dsn := "postgres://postgres:P@ssw0rd@localhost/gophkeeper?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to initialize connection pool: %s", err)
	}

	objectStorage, err := s3.NewObjectStorage()
	if err != nil {
		log.Fatalf("Failed to initialize object storage: %s", err)
	}

	kms, err := service.NewRSAKMS()
	if err != nil {
		log.Fatalf("Failed to initialize kms: %s", err)
	}
	encryptionService := service.NewStandardEncryptionService(kms)
	vault := server.NewVault(ctx, pool, objectStorage, encryptionService)
	lis, err := net.Listen("tcp", "localhost:"+ServerPort)
	if err != nil {
		log.Fatalf("failed to run gRPC server: %v", err)
	}
	userRepo := storage.NewUserRepo(pool)
	authService := service.NewJWTAuthService(userRepo, []byte("access_secret"),
		[]byte("refresh_secret"), AccessTokenTTLHours*time.Hour, RefreshTokenTTLHours*time.Hour)
	authInterceptor := middleware.NewAuthInterceptor(authService)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)
	pb.RegisterGophkeeperServiceServer(grpcServer, pgrpc.NewGophkeeperServer(vault, authService, userRepo))

	logger.Log().Infof("Starting gRPC server on port %s...", ServerPort)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("cannot serve gRPC server: %v", err)
	}
}
