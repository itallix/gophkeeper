package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "gophkeeper.com/mocks/pkg/generated/api/proto/v1"
	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func TestLoginCommands(t *testing.T) {
	// Save original client and restore after tests
	originalClient := client
	defer func() { client = originalClient }()

	mockClient := mocks.NewGophkeeperServiceClient(t)
	client = mockClient

	t.Run("list logins", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewLoginCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().List(mock.Anything, &pb.ListRequest{
			Type: pb.DataType_DATA_TYPE_LOGIN,
		}).Return(&pb.ListResponse{
			Secrets: []string{"login1", "login2"},
		}, nil)

		cmd.SetArgs([]string{"list"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "login1")
		assert.Contains(t, buf.String(), "login2")
	})

	t.Run("get login", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewLoginCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().Get(mock.Anything, &pb.GetRequest{
			Type: pb.DataType_DATA_TYPE_LOGIN,
			Path: "test-login",
		}).Return(&pb.GetResponse{
			Data: &pb.TypedData{
				Base: &pb.Metadata{
					Path:      "test-login",
					CreatedAt: "2024-01-01",
					CreatedBy: "user",
					Metadata:  "test-meta",
				},
				Data: &pb.TypedData_Login{
					Login: &pb.LoginData{
						Login:    "mark",
						Password: "secret",
					},
				},
			},
		}, nil)

		cmd.SetArgs([]string{"get", "-p", "test-login"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "mark")
		assert.Contains(t, buf.String(), "secret")
	})

	t.Run("create login", func(t *testing.T) {
		input := "mark\nsecret\nsecret\n"
		cmd := NewLoginCmd()
		cmd.SetIn(strings.NewReader(input))
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		expectedReq := &pb.CreateRequest{
			Data: &pb.TypedData{
				Type: pb.DataType_DATA_TYPE_LOGIN,
				Base: &pb.Metadata{
					Path: "new-login",
				},
				Data: &pb.TypedData_Login{
					Login: &pb.LoginData{
						Login:    "mark",
						Password: "secret",
					},
				},
			},
		}

		mockClient.EXPECT().Create(mock.Anything, expectedReq).Return(&pb.CreateResponse{
			Message: "Login created successfully",
		}, nil)

		cmd.SetArgs([]string{"create", "-p", "new-login"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Login created successfully")
	})

	t.Run("delete login", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewLoginCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().Delete(mock.Anything, &pb.DeleteRequest{
			Type: pb.DataType_DATA_TYPE_LOGIN,
			Path: "test-login",
		}).Return(&pb.DeleteResponse{
			Message: "Login deleted successfully",
		}, nil)

		cmd.SetArgs([]string{"delete", "-p", "test-login"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Login deleted successfully")
	})
}
