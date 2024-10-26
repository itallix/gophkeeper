package v1

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

type GophkeeperServer struct {
	authService service.AuthenticationService
	authRepo    *storage.UserRepo

	pb.UnimplementedGophkeeperServiceServer
}

func NewGophkeeperServer(authService service.AuthenticationService, authRepo *storage.UserRepo) *GophkeeperServer {
	return &GophkeeperServer{
		authService: authService,
		authRepo:    authRepo,
	}
}

func (srv *GophkeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	token, err := srv.authService.Authenticate(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	resp := &pb.AuthResponse{
		Token:  token,
		UserId: req.GetLogin(),
	}

	return resp, nil
}

func (srv *GophkeeperServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	exists, err := srv.authRepo.Exists(ctx, req.GetLogin())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if exists {
		return nil, status.Errorf(codes.Unauthenticated, "user with login %s already exists", req.GetLogin())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to hash password: %v", err)
	}

	if err = srv.authRepo.CreateUser(ctx, req.GetLogin(), string(hashedPassword)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "error creating a new user %v", err)
	}

	token, err := srv.authService.GetToken(req.GetLogin())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to generate a token: %v", err)
	}

	return &pb.AuthResponse{
		Token:  token,
		UserId: req.GetLogin(),
	}, nil
}
