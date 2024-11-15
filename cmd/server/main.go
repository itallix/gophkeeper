package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"github.com/itallix/gophkeeper/internal/common/logger"
	"github.com/itallix/gophkeeper/internal/server"
	pgrpc "github.com/itallix/gophkeeper/internal/server/grpc"
	"github.com/itallix/gophkeeper/internal/server/grpc/middleware"
	"github.com/itallix/gophkeeper/internal/server/s3"
	"github.com/itallix/gophkeeper/internal/server/service"
	"github.com/itallix/gophkeeper/internal/server/storage"
	pb "github.com/itallix/gophkeeper/pkg/generated/api/proto/v1"
)

type config struct {
	Address          string `env:"ADDRESS" envDefault:"localhost:8081"`
	DSN              string `env:"DB_DSN" envDefault:"postgres://postgres:P@ssw0rd@localhost/gophkeeper?sslmode=disable"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"DEBUG"`
	AccessSecret     string `env:"ACCESS_SECRET" envDefault:"access_secret"`
	RefreshSecret    string `env:"REFRESH_SECRET" envDefault:"refresh_secret"`
	MasterKeyPath    string `env:"MASTER_KEY" envDefault:"testdata/private.pem"`
	EncryptedKeyPath string `env:"ENCRYPTED_KEY" envDefault:"testdata/encrypted_key.bin"`
}

const (
	AccessTokenTTLHours  = 1
	RefreshTokenTTLHours = 24
	ShutdownTimeoutSec   = 30
)

func createServer(ctx context.Context, cfg config) (*grpc.Server, net.Listener, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}

	objectStorage, err := s3.NewObjectStorage()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize object storage: %w", err)
	}

	kms, err := service.NewRSAKMS(cfg.MasterKeyPath, cfg.EncryptedKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize kms: %w", err)
	}
	encryptionService := service.NewStandardEncryptionService(kms)
	vault := server.NewVaultImpl(ctx, pool, objectStorage, encryptionService)
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed liseting address: %w", err)
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

	return grpcServer, lis, nil
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("cannot parse config: %w", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("cannot instantiate zap logger: %w", err)
	}
	grpcServer, lis, err := createServer(ctx, cfg)
	if err != nil {
		return err
	}

	go func() {
		logger.Log().Infof("Starting gRPC server %s...", cfg.Address)
		if err = grpcServer.Serve(lis); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Log().Errorf("failed to serve gRPC server: %v", err)
			cancel()
		}
	}()

	select {
	case sig := <-sigChan:
		logger.Log().Infof("Received signal: %v", sig)
	case <-ctx.Done():
		logger.Log().Info("Shutting down due to context cancellation")
	}

	logger.Log().Info("Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), ShutdownTimeoutSec*time.Second)
	defer shutdownCancel()

	done := make(chan bool)
	go func() {
		// Gracefully stop the gRPC server
		grpcServer.GracefulStop()
		done <- true
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-shutdownCtx.Done():
		logger.Log().Warn("Shutdown timed out, forcing exit")
		grpcServer.Stop()
	case <-done:
		logger.Log().Info("Graceful shutdown completed")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		logger.Log().Fatal(err)
	}
}
