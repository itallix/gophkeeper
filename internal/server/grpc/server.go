package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper.com/internal/common/logger"
	"gophkeeper.com/internal/server"
	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

type contextKey string

const UsernameKey contextKey = "username"

type GophkeeperServer struct {
	authService service.AuthenticationService
	authRepo    *storage.UserRepo
	vault       server.Vault

	pb.UnimplementedGophkeeperServiceServer
}

func NewGophkeeperServer(vault server.Vault, authService service.AuthenticationService,
	authRepo *storage.UserRepo) *GophkeeperServer {
	return &GophkeeperServer{
		authService: authService,
		authRepo:    authRepo,
		vault:       vault,
	}
}

func (srv *GophkeeperServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	pair, err := srv.authService.Authenticate(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	resp := &pb.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		UserId:       req.GetLogin(),
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

	pair, err := srv.authService.GetTokenPair(req.GetLogin())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to generate a token: %v", err)
	}

	return &pb.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		UserId:       req.GetLogin(),
	}, nil
}

func (srv *GophkeeperServer) RefreshToken(_ context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
	pair, err := srv.authService.RefreshTokens(req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	username, err := srv.authService.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate new access token")
	}

	return &pb.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		UserId:       username,
	}, nil
}

func (srv *GophkeeperServer) List(_ context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	var secret models.Secret

	switch req.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		secret = models.NewLogin(nil, nil)
	case pb.DataType_DATA_TYPE_CARD:
		secret = models.NewCard(nil, nil)
	case pb.DataType_DATA_TYPE_NOTE:
		secret = models.NewNote(nil, nil)
	case pb.DataType_DATA_TYPE_BINARY:
		secret = models.NewBinary(nil, nil)
	case pb.DataType_DATA_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.Internal, "unspecified data type is not allowed")
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
	username, ok := ctx.Value(UsernameKey).(string)
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

	switch req.GetData().GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		loginData := data.GetLogin()
		secret = models.NewLogin(
			opts,
			[]models.LoginOption{
				models.WithLogin(loginData.GetLogin()),
				models.WithPassword(loginData.GetPassword()),
			},
		)
	case pb.DataType_DATA_TYPE_CARD:
		cardData := data.GetCard()
		secret = models.NewCard(
			opts,
			[]models.CardOption{
				models.WithCardHolder(cardData.GetCardHolder()),
				models.WithCardNumber(cardData.GetNumber()),
				models.WithExpiry(cardData.GetExpiryMonth(), cardData.GetExpiryYear()),
				models.WithCVC(cardData.GetCvv()),
			})
	case pb.DataType_DATA_TYPE_NOTE:
		secret = models.NewNote(
			opts,
			[]models.NoteOption{
				models.WithText(data.GetNote().GetText()),
			},
		)
	case pb.DataType_DATA_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.Internal, "unspecified data type is not allowed")
	case pb.DataType_DATA_TYPE_BINARY:
		return nil, status.Error(codes.Internal, "binary data type is not allowed")
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetData().GetType())
	}

	if err := srv.vault.StoreSecret(secret); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.CreateResponse{
		Message: fmt.Sprintf("secret with path=%s has been successfully created", path),
	}, nil
}

func (srv *GophkeeperServer) Delete(_ context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
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
		secret = models.NewBinary(opts, nil)
	case pb.DataType_DATA_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.Internal, "unspecified data type is not allowed")
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
	}

	if err := srv.vault.DeleteSecret(secret); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	return &pb.DeleteResponse{
		Message: fmt.Sprintf("secret with path=%s has been successfully deleted", req.GetPath()),
	}, nil
}

func (srv *GophkeeperServer) Get(_ context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
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
	case pb.DataType_DATA_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.Internal, "unspecified data type is not allowed")
	case pb.DataType_DATA_TYPE_BINARY:
		return nil, status.Error(codes.Internal, "binary data type is not allowed")
	default:
		return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
	}

	if err := srv.vault.RetrieveSecret(secret); err != nil {
		return nil, status.Errorf(codes.Internal, "cannot perform the action %v", err)
	}

	switch req.GetType() {
	case pb.DataType_DATA_TYPE_LOGIN:
		login, ok := secret.(*models.Login)
		if !ok {
			return nil, status.Errorf(codes.Internal,
				"invalid type assertion: expected *models.Login, got %T", secret)
		}
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
		card, ok := secret.(*models.Card)
		if !ok {
			return nil, status.Errorf(codes.Internal,
				"invalid type assertion: expected *models.Card, got %T", secret)
		}
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
						Number:      string(card.Number),
						CardHolder:  card.CardholderName,
						ExpiryMonth: card.ExpiryMonth,
						ExpiryYear:  card.ExpiryYear,
						Cvv:         string(card.CVC),
					},
				},
			},
		}, nil
	case pb.DataType_DATA_TYPE_NOTE:
		note, ok := secret.(*models.Note)
		if !ok {
			return nil, status.Errorf(codes.Internal,
				"invalid type assertion: expected *models.Note, got %T", secret)
		}
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
	case pb.DataType_DATA_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.Internal, "unspecified data type is not allowed")
	case pb.DataType_DATA_TYPE_BINARY:
		return nil, status.Error(codes.Internal, "binary data type is not allowed")
	}

	return nil, status.Errorf(codes.Internal, "unknown data type: %v", req.GetType())
}

func (srv *GophkeeperServer) Upload(stream pb.GophkeeperService_UploadServer) error {
	username, ok := stream.Context().Value(UsernameKey).(string)
	if !ok {
		return status.Error(codes.Internal, "username not found in context")
	}

	var (
		lastChunk  *pb.Chunk
		encDataKey []byte
	)
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if chunk.Data == nil {
			lastChunk = chunk
			break
		}
		currentHash := sha256.Sum256(chunk.GetData())
		if chunk.GetHash() != hex.EncodeToString(currentHash[:]) {
			return status.Error(codes.Aborted, "aborted upload due to chunk hash mismatch")
		}

		binary := models.NewBinary(
			[]models.SecretOption{
				models.WithPath(chunk.GetFilename()),
				models.WithEncryptedDataKey(encDataKey),
			},
			[]models.BinaryOption{
				models.WithChunkID(chunk.GetChunkId()),
				models.WithData(chunk.GetData()),
			})
		if err = srv.vault.StoreSecret(binary); err != nil {
			return status.Errorf(codes.Internal, "failed to store chunk: %v", err)
		}
		if encDataKey == nil {
			encDataKey = binary.EncryptedDataKey
		}
	}
	binary := models.NewBinary(
		[]models.SecretOption{
			models.WithPath(lastChunk.GetFilename()),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
			models.WithEncryptedDataKey(encDataKey),
		},
		[]models.BinaryOption{
			models.WithChunks(lastChunk.GetChunkId()),
			models.WithHash(lastChunk.GetHash()),
			models.WithData(nil),
		},
	)
	if err := srv.vault.StoreSecret(binary); err != nil {
		return status.Errorf(codes.Internal, "failed to store chunk: %v", err)
	}
	if err := stream.SendAndClose(&pb.UploadResponse{
		Message: fmt.Sprintf("Upload of %s with %d chunks has been completed",
			lastChunk.GetFilename(), lastChunk.GetChunkId()),
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to close stream: %v", err)
	}
	return nil
}

func (srv *GophkeeperServer) Download(req *pb.DownloadRequest, stream pb.GophkeeperService_DownloadServer) error {
	binary := models.NewBinary(
		[]models.SecretOption{
			models.WithPath(req.GetFilename()),
		},
		nil,
	)
	if err := srv.vault.RetrieveSecret(binary); err != nil {
		return status.Errorf(codes.Internal, "failed to retrieve binary metadata: %v", err)
	}

	for i := range binary.Chunks {
		chunk := models.NewBinary(
			[]models.SecretOption{
				models.WithPath(req.GetFilename()),
				models.WithEncryptedDataKey(binary.EncryptedDataKey),
			},
			[]models.BinaryOption{
				models.WithChunkID(i),
				models.WithChunks(binary.Chunks),
			},
		)
		if err := srv.vault.RetrieveSecret(chunk); err != nil {
			return status.Errorf(codes.Internal, "failed to retrieve chunk data: %v", err)
		}
		logger.Log().Infof("Download chunk: %d %d", chunk.ChunkID, len(chunk.Data))
		chunkHash := sha256.Sum256(chunk.Data)
		if err := stream.Send(&pb.Chunk{
			Filename: binary.Path,
			Data:     chunk.Data,
			ChunkId:  chunk.ChunkID,
			Hash:     hex.EncodeToString(chunkHash[:]),
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk data: %v", err)
		}
	}

	if err := stream.Send(&pb.Chunk{
		Filename: binary.Path,
		Hash:     binary.Hash,
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send chunk data: %v", err)
	}

	return nil
}
