package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"github.com/caarlos0/env"

	"gophkeeper.com/internal/common/logger"
	"gophkeeper.com/internal/server"
	pgrpc "gophkeeper.com/internal/server/grpc"
	"gophkeeper.com/internal/server/grpc/middleware"
	"gophkeeper.com/internal/server/s3"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

type config struct {
	Address       string `env:"ADDRESS" envDefault:"localhost:8081"`
	DatabaseDSN   string `env:"DATABASE_DSN" envDefault:"postgres://postgres:P@ssw0rd@localhost/gophkeeper?sslmode=disable"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"DEBUG"`
	AccessSecret  string `env:"ACCESS_SECRET" envDefault:"access_secret"`
	RefreshSecret string `env:"REFRESH_SECRET" envDefault:"refresh_secret"`
}

const (
	AccessTokenTTLHours  = 1
	RefreshTokenTTLHours = 24
)

func main() {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Cannot parse config: %s", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
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
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalf("failed to run gRPC server: %v", err)
	}
	userRepo := storage.NewUserRepo(pool)
	authService := service.NewJWTAuthService(userRepo, []byte(cfg.AccessSecret),
		[]byte(cfg.RefreshSecret), AccessTokenTTLHours*time.Hour, RefreshTokenTTLHours*time.Hour)
	authInterceptor := middleware.NewAuthInterceptor(authService)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)
	pb.RegisterGophkeeperServiceServer(grpcServer, pgrpc.NewGophkeeperServer(vault, authService, userRepo))

	logger.Log().Infof("Starting gRPC server %s...", cfg.Address)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("cannot serve gRPC server: %v", err)
	}
}
