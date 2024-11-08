package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper.com/internal/server/grpc"
	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/service"
	mocksrv "gophkeeper.com/mocks/internal_/server"
	mocks "gophkeeper.com/mocks/internal_/server/service"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mocks.AuthenticationService)
		request   *pb.LoginRequest
		wantError bool
		errorCode codes.Code
	}{
		{
			name: "successful_login",
			setup: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					Authenticate(mock.Anything, "user1", "pass1").
					Return(&service.TokenPair{AccessToken: "at", RefreshToken: "rt"}, nil)
			},
			request:   &pb.LoginRequest{Login: "user1", Password: "pass1"},
			wantError: false,
		},
		{
			name: "invalid_credentials",
			setup: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					Authenticate(mock.Anything, "user1", "wrong").
					Return(nil, errors.New("invalid credentials"))
			},
			request:   &pb.LoginRequest{Login: "user1", Password: "wrong"},
			wantError: true,
			errorCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := mocks.NewAuthenticationService(t)
			tt.setup(authService)

			server := grpc.NewGophkeeperServer(nil, authService, nil)
			resp, err := server.Login(context.Background(), tt.request)

			if tt.wantError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.GetAccessToken())
				assert.NotEmpty(t, resp.GetRefreshToken())
				assert.Equal(t, tt.request.GetLogin(), resp.GetUserId())
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mocksrv.Vault)
		request   *pb.CreateRequest
		username  string
		wantError bool
		errorCode codes.Code
	}{
		{
			name: "create_login",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					StoreSecret(mock.MatchedBy(func(s models.Secret) bool {
						login, ok := s.(*models.Login)
						return ok && login.Path == "/test/path"
					})).
					Return(nil)
			},
			request: &pb.CreateRequest{
				Data: &pb.TypedData{
					Base: &pb.Metadata{Path: "/test/path"},
					Data: &pb.TypedData_Login{
						Login: &pb.LoginData{
							Login:    "testuser",
							Password: "testpass",
						},
					},
					Type: pb.DataType_DATA_TYPE_LOGIN,
				},
			},
			username:  "testuser",
			wantError: false,
		},
		{
			name: "create_card",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					StoreSecret(mock.MatchedBy(func(s models.Secret) bool {
						note, ok := s.(*models.Card)
						return ok && note.Path == "/test/card"
					})).
					Return(nil)
			},
			request: &pb.CreateRequest{
				Data: &pb.TypedData{
					Base: &pb.Metadata{Path: "/test/card"},
					Data: &pb.TypedData_Card{
						Card: &pb.CardData{
							CardHolder:  "Adam Smith",
							Number:      "2233445566778899",
							ExpiryMonth: 8,
							ExpiryYear:  int64(time.Now().Year() + 2),
							Cvv:         "237",
						},
					},
					Type: pb.DataType_DATA_TYPE_CARD,
				},
			},
			username:  "testuser",
			wantError: false,
		},
		{
			name: "create_note",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					StoreSecret(mock.MatchedBy(func(s models.Secret) bool {
						note, ok := s.(*models.Note)
						return ok && note.Path == "/test/note"
					})).
					Return(nil)
			},
			request: &pb.CreateRequest{
				Data: &pb.TypedData{
					Base: &pb.Metadata{Path: "/test/note"},
					Data: &pb.TypedData_Note{
						Note: &pb.NoteData{Text: "test note"},
					},
					Type: pb.DataType_DATA_TYPE_NOTE,
				},
			},
			username:  "testuser",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault := mocksrv.NewVault(t)
			tt.setup(vault)

			server := grpc.NewGophkeeperServer(vault, nil, nil)
			ctx := context.WithValue(context.Background(), grpc.UsernameKey, tt.username)
			resp, err := server.Create(ctx, tt.request)

			if tt.wantError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Contains(t, resp.GetMessage(), "successfully created")
			}
		})
	}
}

func TestGet(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name      string
		setup     func(*mocksrv.Vault)
		request   *pb.GetRequest
		wantError bool
		errorCode codes.Code
	}{
		{
			name: "get_login",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					RetrieveSecret(mock.MatchedBy(func(s models.Secret) bool {
						login, ok := s.(*models.Login)
						if ok {
							login.Login = "testuser"
							login.Password = []byte("testpass")
							login.CreatedAt = testTime
							login.CreatedBy = "creator"
							return true
						}
						return false
					})).
					Return(nil)
			},
			request: &pb.GetRequest{
				Path: "/test/path",
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			wantError: false,
		},
		{
			name: "secret_not_found",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					RetrieveSecret(mock.Anything).
					Return(errors.New("secret not found"))
			},
			request: &pb.GetRequest{
				Path: "/nonexistent",
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			wantError: true,
			errorCode: codes.Internal,
		},
		{
			name: "get_card",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					RetrieveSecret(mock.MatchedBy(func(s models.Secret) bool {
						card, ok := s.(*models.Card)
						if ok {
							card.CardholderName = "testuser"
							card.Number = []byte("2211443366558877")
							card.ExpiryMonth = 8
							card.ExpiryYear = int64(time.Now().Year() + 2)
							card.CVC = []byte("264")
							card.CreatedAt = testTime
							card.CreatedBy = "creator"
							return true
						}
						return false
					})).
					Return(nil)
			},
			request: &pb.GetRequest{
				Path: "/test/path",
				Type: pb.DataType_DATA_TYPE_CARD,
			},
			wantError: false,
		},
		{
			name: "get_note",
			setup: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					RetrieveSecret(mock.MatchedBy(func(s models.Secret) bool {
						note, ok := s.(*models.Note)
						if ok {
							note.Text = []byte("lorem ipsum")
							note.CreatedAt = testTime
							note.CreatedBy = "creator"
							return true
						}
						return false
					})).
					Return(nil)
			},
			request: &pb.GetRequest{
				Path: "/test/path",
				Type: pb.DataType_DATA_TYPE_NOTE,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault := mocksrv.NewVault(t)
			tt.setup(vault)

			server := grpc.NewGophkeeperServer(vault, nil, nil)
			resp, err := server.Get(context.Background(), tt.request)

			if tt.wantError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.GetData())
			}
		})
	}
}
