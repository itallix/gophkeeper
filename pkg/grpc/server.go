package grpc

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper.com/internal/server"
	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

type GophkeeperServer struct {
	authService service.AuthenticationService
	authRepo    *storage.UserRepo
	vault       *server.Vault

	pb.UnimplementedGophkeeperServiceServer
}

func NewGophkeeperServer(vault *server.Vault, authService service.AuthenticationService,
	authRepo *storage.UserRepo) *GophkeeperServer {
	return &GophkeeperServer{
		authService: authService,
		authRepo:    authRepo,
		vault:       vault,
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

func (srv *GophkeeperServer) ListLogins(ctx context.Context, req *pb.ListLoginRequest) (*pb.ListLoginResponse, error) {
	login := models.NewLogin(nil, nil)
	list, err := srv.vault.ListSecrets(login)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.ListLoginResponse{
		Secrets: list,
	}, nil
}

func (srv *GophkeeperServer) CreateLogin(ctx context.Context, req *pb.CreateLoginRequest) (*pb.CreateLoginResponse, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, status.Error(codes.Internal, "username not found in context")
	}

	login := models.NewLogin(
		[]models.SecretOption{
			models.WithPath(req.GetPath()),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		},
		[]models.LoginOption{
			models.WithLogin(req.GetLogin()),
			models.WithPassword(req.GetPassword()),
		},
	)
	if err := srv.vault.StoreSecret(login); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.CreateLoginResponse{
		Message: fmt.Sprintf("login with path=%s has been successfully created", req.GetPath()),
	}, nil
}

func (srv *GophkeeperServer) DeleteLogin(ctx context.Context, req *pb.DeleteLoginRequest) (*pb.DeleteLoginResponse, error) {
	login := models.NewLogin(
		[]models.SecretOption{
			models.WithPath(req.GetPath()),
		},
		nil,
	)
	if err := srv.vault.DeleteSecret(login); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.DeleteLoginResponse{
		Message: fmt.Sprintf("login with path=%s has been successfully deleted", req.GetPath()),
	}, nil
}

func (srv *GophkeeperServer) GetLogin(ctx context.Context, req *pb.GetLoginRequest) (*pb.GetLoginResponse, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, status.Error(codes.Internal, "username not found in context")
	}

	login := models.NewLogin(
		[]models.SecretOption{
			models.WithPath(req.GetPath()),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		},
		nil,
	)
	if err := srv.vault.RetrieveSecret(login); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.GetLoginResponse{
		Login:     login.Login,
		Password:  string(login.Password),
		CreatedBy: login.CreatedBy,
		CreatedAt: login.CreatedAt.Format("2006-01-02 15:04:05"),
		Path:      login.Path,
		Metadata:  fmt.Sprintf("%v", login.CustomMeta),
	}, nil
}
