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

func (srv *GophkeeperServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	var secret models.Secret

	switch req.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		secret = models.NewLogin(nil, nil)
	case pb.DataType_DATA_TYPE_CARD:
		secret = models.NewCard(nil, nil)
	case pb.DataType_DATA_TYPE_NOTE:
		secret = models.NewNote(nil, nil)
	case pb.DataType_DATA_TYPE_BINARY:
		secret = models.NewNote(nil, nil)
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
	}

	list, err := srv.vault.ListSecrets(secret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.ListResponse{
		Secrets: list,
	}, nil
}

func (srv *GophkeeperServer) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, status.Error(codes.Internal, "username not found in context")
	}
	var secret models.Secret
	data := req.GetData()
	path := data.GetBase().GetPath()
	opts := []models.SecretOption{
		models.WithPath(path),
		models.WithCreatedBy(username),
		models.WithModifiedBy(username),
	}

	switch req.Data.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		loginData := data.GetLogin()
		secret = models.NewLogin(
			opts,
			[]models.LoginOption{
				models.WithLogin(loginData.Login),
				models.WithPassword(loginData.Password),
			},
		)
	case pb.DataType_DATA_TYPE_CARD:
		cardData := data.GetCard()
		secret = models.NewCard(
			opts, 
			[]models.CardOption{
				models.WithCardHolder(cardData.CardHolder),
				models.WithCardNumber(cardData.Number),
				models.WithExpiry(int8(cardData.ExpiryMonth), int16(cardData.ExpiryYear)),
				models.WithCVC(cardData.Cvv),
			})
	case pb.DataType_DATA_TYPE_NOTE:
		secret = models.NewNote(
			opts,
			[]models.NoteOption{
				models.WithText(data.GetNote().Text),
			},
		)
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.Data.GetType())
	}

	if err := srv.vault.StoreSecret(secret); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.CreateResponse{
		Message: fmt.Sprintf("login with path=%s has been successfully created", path),
	}, nil
}

func (srv *GophkeeperServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	var secret models.Secret
	opts := []models.SecretOption{
		models.WithPath(req.GetPath()),
	}

	switch req.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		secret = models.NewLogin(opts, nil)
	case pb.DataType_DATA_TYPE_CARD:
		secret = models.NewCard(opts, nil)
	case pb.DataType_DATA_TYPE_NOTE:
		secret = models.NewNote(opts, nil)
	case pb.DataType_DATA_TYPE_BINARY:
		secret = models.NewNote(opts, nil)
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
	}

	if err := srv.vault.DeleteSecret(secret); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.DeleteResponse{
		Message: fmt.Sprintf("login with path=%s has been successfully deleted", req.GetPath()),
	}, nil
}

func (srv *GophkeeperServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	_, ok := ctx.Value("username").(string)
	if !ok {
		return nil, status.Error(codes.Internal, "username not found in context")
	}
	var secret models.Secret
	opts := []models.SecretOption{
		models.WithPath(req.GetPath()),
	}

	switch req.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		secret = models.NewLogin(opts, nil)
	case pb.DataType_DATA_TYPE_CARD:
		secret = models.NewCard(opts, nil)
	case pb.DataType_DATA_TYPE_NOTE:
		secret = models.NewNote(opts, nil)
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
	}

	if err := srv.vault.RetrieveSecret(secret); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	switch req.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		login := secret.(*models.Login)
		return &pb.GetResponse{
			Data: &pb.TypedData{
				Base: &pb.Metadata{
					CreatedBy: login.CreatedBy,
					CreatedAt: login.CreatedAt.Format("2006-01-02 15:04:05"),
					Path:      login.Path,
					Metadata:  fmt.Sprintf("%v", login.CustomMeta),
				},
				Data: &pb.TypedData_Login{
					Login: &pb.LoginData{
						Login:    login.Login,
						Password: string(login.Password),
					},
				},
			},
		}, nil
	case pb.DataType_DATA_TYPE_CARD:
		card := secret.(*models.Card)
		return &pb.GetResponse{
			Data: &pb.TypedData{
				Base: &pb.Metadata{
					CreatedBy: card.CreatedBy,
					CreatedAt: card.CreatedAt.Format("2006-01-02 15:04:05"),
					Path:      card.Path,
					Metadata:  fmt.Sprintf("%v", card.CustomMeta),
				},
				Data: &pb.TypedData_Card{
					Card: &pb.CardData{
						Number: string(card.Number),
						CardHolder: card.CardholderName,
						ExpiryMonth: int32(card.ExpiryMonth),
						ExpiryYear: int32(card.ExpiryYear),
						Cvv: string(card.CVC),
					},
				},
			},
		}, nil
	case pb.DataType_DATA_TYPE_NOTE:
		note := secret.(*models.Note)
		return &pb.GetResponse{
			Data: &pb.TypedData{
				Base: &pb.Metadata{
					CreatedBy: note.CreatedBy,
					CreatedAt: note.CreatedAt.Format("2006-01-02 15:04:05"),
					Path:      note.Path,
					Metadata:  fmt.Sprintf("%v", note.CustomMeta),
				},
				Data: &pb.TypedData_Note{
					Note: &pb.NoteData{
						Text: string(note.Text),
					},
				},
			},
		}, nil
	}

	return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
}
