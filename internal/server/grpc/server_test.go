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

func TestRefreshToken(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.RefreshTokenRequest
		setupMock     func(*mocks.AuthenticationService)
		expectedResp  *pb.AuthResponse
		expectedError *status.Status
	}{
		{
			name: "successful_token_refresh",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMock: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					RefreshTokens("valid_refresh_token").
					Return(
						&service.TokenPair{
							AccessToken:  "new_access_token",
							RefreshToken: "new_refresh_token",
						},
						nil,
					)
				mas.EXPECT().
					ValidateAccessToken("new_access_token").
					Return("test_user", nil)
			},
			expectedResp: &pb.AuthResponse{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				UserId:       "test_user",
			},
			expectedError: nil,
		},
		{
			name: "invalid_refresh_token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "invalid_refresh_token",
			},
			setupMock: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					RefreshTokens("invalid_refresh_token").
					Return(nil, errors.New("token expired"))
			},
			expectedResp:  nil,
			expectedError: status.New(codes.Unauthenticated, "invalid refresh token"),
		},
		{
			name: "token_validation_failure",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMock: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					RefreshTokens("valid_refresh_token").
					Return(
						&service.TokenPair{
							AccessToken:  "new_access_token",
							RefreshToken: "new_refresh_token",
						},
						nil,
					)
				mas.EXPECT().
					ValidateAccessToken("new_access_token").
					Return("", errors.New("validation failed"))
			},
			expectedResp:  nil,
			expectedError: status.New(codes.Internal, "failed to validate new access token"),
		},
		{
			name: "empty_refresh_token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "",
			},
			setupMock: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					RefreshTokens("").
					Return(nil, errors.New("empty token"))
			},
			expectedResp:  nil,
			expectedError: status.New(codes.Unauthenticated, "invalid refresh token"),
		},
		{
			name: "malformed_token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "malformed.token.string",
			},
			setupMock: func(mas *mocks.AuthenticationService) {
				mas.EXPECT().
					RefreshTokens("malformed.token.string").
					Return(nil, errors.New("malformed token"))
			},
			expectedResp:  nil,
			expectedError: status.New(codes.Unauthenticated, "invalid refresh token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := mocks.NewAuthenticationService(t)
			if tt.setupMock != nil {
				tt.setupMock(authService)
			}

			server := grpc.NewGophkeeperServer(nil, authService, nil)
			resp, err := server.RefreshToken(context.Background(), tt.request)

			if tt.expectedError != nil {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError.Code(), st.Code())
				assert.Equal(t, tt.expectedError.Message(), st.Message())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedResp.GetAccessToken(), resp.GetAccessToken())
				assert.Equal(t, tt.expectedResp.GetRefreshToken(), resp.GetRefreshToken())
				assert.Equal(t, tt.expectedResp.GetUserId(), resp.GetUserId())
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
		{
			name: "create_binary",
			request: &pb.CreateRequest{
				Data: &pb.TypedData{
					Base: &pb.Metadata{Path: "/test/binary"},
					Type: pb.DataType_DATA_TYPE_BINARY,
				},
			},
			username:  "testuser",
			wantError: true,
			errorCode: codes.Internal,
		},
		{
			name: "unspecified",
			request: &pb.CreateRequest{
				Data: &pb.TypedData{
					Base: &pb.Metadata{Path: "/test/unspecified"},
					Type: pb.DataType_DATA_TYPE_UNSPECIFIED,
				},
			},
			username:  "testuser",
			wantError: true,
			errorCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault := mocksrv.NewVault(t)
			if tt.setup != nil {
				tt.setup(vault)
			}

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
		{
			name: "get_binary",
			request: &pb.GetRequest{
				Path: "/test/path",
				Type: pb.DataType_DATA_TYPE_BINARY,
			},
			wantError: true,
			errorCode: codes.Internal,
		},
		{
			name: "unspecified",
			request: &pb.GetRequest{
				Path: "/test/path",
				Type: pb.DataType_DATA_TYPE_UNSPECIFIED,
			},
			wantError: true,
			errorCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault := mocksrv.NewVault(t)
			if tt.setup != nil {
				tt.setup(vault)
			}

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

func TestDelete(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.DeleteRequest
		setupMock     func(*mocksrv.Vault)
		expectedMsg   string
		expectedError *status.Status
	}{
		{
			name: "delete_login_success",
			request: &pb.DeleteRequest{
				Path: "/test/login",
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					DeleteSecret(mock.MatchedBy(func(s models.Secret) bool {
						login, ok := s.(*models.Login)
						return ok && login.Path == "/test/login"
					})).
					Return(nil)
			},
			expectedMsg:   "secret with path=/test/login has been successfully deleted",
			expectedError: nil,
		},
		{
			name: "delete_card_success",
			request: &pb.DeleteRequest{
				Path: "/test/card",
				Type: pb.DataType_DATA_TYPE_CARD,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					DeleteSecret(mock.MatchedBy(func(s models.Secret) bool {
						card, ok := s.(*models.Card)
						return ok && card.Path == "/test/card"
					})).
					Return(nil)
			},
			expectedMsg:   "secret with path=/test/card has been successfully deleted",
			expectedError: nil,
		},
		{
			name: "delete_note_success",
			request: &pb.DeleteRequest{
				Path: "/test/note",
				Type: pb.DataType_DATA_TYPE_NOTE,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					DeleteSecret(mock.MatchedBy(func(s models.Secret) bool {
						note, ok := s.(*models.Note)
						return ok && note.Path == "/test/note"
					})).
					Return(nil)
			},
			expectedMsg:   "secret with path=/test/note has been successfully deleted",
			expectedError: nil,
		},
		{
			name: "delete_binary_success",
			request: &pb.DeleteRequest{
				Path: "/test/binary",
				Type: pb.DataType_DATA_TYPE_BINARY,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					DeleteSecret(mock.MatchedBy(func(s models.Secret) bool {
						binary, ok := s.(*models.Binary)
						return ok && binary.Path == "/test/binary"
					})).
					Return(nil)
			},
			expectedMsg:   "secret with path=/test/binary has been successfully deleted",
			expectedError: nil,
		},
		{
			name: "delete_with_vault_error",
			request: &pb.DeleteRequest{
				Path: "/test/error",
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					DeleteSecret(mock.Anything).
					Return(errors.New("vault error"))
			},
			expectedMsg:   "",
			expectedError: status.New(codes.Internal, "cannot perform the action vault error"),
		},
		{
			name: "delete_with_unspecified_type",
			request: &pb.DeleteRequest{
				Path: "/test/unspecified",
				Type: pb.DataType_DATA_TYPE_UNSPECIFIED,
			},
			expectedMsg:   "",
			expectedError: status.New(codes.Internal, "unspecified data type is not allowed"),
		},
		{
			name: "delete_with_invalid_type",
			request: &pb.DeleteRequest{
				Path: "/test/invalid",
				Type: pb.DataType(99), // Invalid type
			},
			expectedMsg:   "",
			expectedError: status.Newf(codes.Internal, "unknown data type: %v", pb.DataType(99)),
		},
		{
			name: "delete_with_empty_path",
			request: &pb.DeleteRequest{
				Path: "",
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					DeleteSecret(mock.MatchedBy(func(s models.Secret) bool {
						login, ok := s.(*models.Login)
						return ok && login.Path == ""
					})).
					Return(errors.New("invalid path"))
			},
			expectedMsg:   "",
			expectedError: status.New(codes.Internal, "cannot perform the action invalid path"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVault := mocksrv.NewVault(t)
			if tt.setupMock != nil {
				tt.setupMock(mockVault)
			}

			server := grpc.NewGophkeeperServer(mockVault, nil, nil)
			resp, err := server.Delete(context.Background(), tt.request)

			if tt.expectedError != nil {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError.Code(), st.Code())
				assert.Equal(t, tt.expectedError.Message(), st.Message())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedMsg, resp.GetMessage())
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.ListRequest
		setupMock     func(*mocksrv.Vault)
		expectedList  []string
		expectedError *status.Status
	}{
		{
			name: "list_logins_success",
			request: &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					ListSecrets(mock.Anything).
					Return([]string{"login1", "login2"}, nil)
			},
			expectedList:  []string{"login1", "login2"},
			expectedError: nil,
		},
		{
			name: "list_cards_success",
			request: &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_CARD,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					ListSecrets(mock.Anything).
					Return([]string{"card1", "card2"}, nil)
			},
			expectedList:  []string{"card1", "card2"},
			expectedError: nil,
		},
		{
			name: "list_notes_success",
			request: &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					ListSecrets(mock.Anything).
					Return([]string{"note1", "note2"}, nil)
			},
			expectedList:  []string{"note1", "note2"},
			expectedError: nil,
		},
		{
			name: "list_binary_success",
			request: &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_BINARY,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					ListSecrets(mock.Anything).
					Return([]string{"binary1", "binary2"}, nil)
			},
			expectedList:  []string{"binary1", "binary2"},
			expectedError: nil,
		},
		{
			name: "list_with_vault_error",
			request: &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_LOGIN,
			},
			setupMock: func(mv *mocksrv.Vault) {
				mv.EXPECT().
					ListSecrets(mock.Anything).
					Return(nil, errors.New("vault error"))
			},
			expectedList:  nil,
			expectedError: status.New(codes.Internal, "cannot perform the action vault error"),
		},
		{
			name: "list_with_unspecified_type",
			request: &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_UNSPECIFIED,
			},
			expectedList:  nil,
			expectedError: status.New(codes.Internal, "unspecified data type is not allowed"),
		},
		{
			name: "list_with_invalid_type",
			request: &pb.ListRequest{
				Type: pb.DataType(99), // Invalid type
			},
			expectedList:  nil,
			expectedError: status.Newf(codes.Internal, "unknown data type: %v", pb.DataType(99)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVault := mocksrv.NewVault(t)
			if tt.setupMock != nil {
				tt.setupMock(mockVault)
			}

			server := grpc.NewGophkeeperServer(mockVault, nil, nil)
			resp, err := server.List(context.Background(), tt.request)

			if tt.expectedError != nil {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError.Code(), st.Code())
				assert.Equal(t, tt.expectedError.Message(), st.Message())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedList, resp.GetSecrets())
			}
		})
	}
}
