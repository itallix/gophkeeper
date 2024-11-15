package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/itallix/gophkeeper/internal/client/jwt"
	mocks "github.com/itallix/gophkeeper/mocks/pkg/generated/api/proto/v1"
	pb "github.com/itallix/gophkeeper/pkg/generated/api/proto/v1"
)

func TestUserCommands(t *testing.T) {
	originalClient := client
	originalTokenProvider := tokenProvider
	defer func() {
		client = originalClient
		tokenProvider = originalTokenProvider
	}()

	mockClient := mocks.NewGophkeeperServiceClient(t)
	client = mockClient
	tmp, err := os.CreateTemp("", "token")
	config.TokenFile = tmp.Name()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	tokenProvider = jwt.NewTokenProvider(tmp.Name())

	t.Run("register a new user", func(t *testing.T) {
		input := "secret\nsecret\n"
		buf := new(bytes.Buffer)
		cmd := NewUserCmd()
		cmd.SetIn(strings.NewReader(input))
		cmd.SetOut(buf)

		mockClient.EXPECT().Register(mock.Anything, &pb.RegisterRequest{
			Login:    "mark",
			Password: "secret",
		}).Return(&pb.AuthResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}, nil)

		cmd.SetArgs([]string{"register", "-l", "mark"})
		err = cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "User with login=mark successfully registered")
		_, err = tmp.Seek(0, io.SeekStart)
		require.NoError(t, err)
		tokens, readErr := io.ReadAll(tmp)
		require.NoError(t, readErr)
		want := "{\"access_token\":\"access_token\",\"refresh_token\":\"refresh_token\"}"
		assert.Equal(t, want, string(tokens))
	})

	t.Run("login as a user", func(t *testing.T) {
		input := "secret\n"
		buf := new(bytes.Buffer)
		cmd := NewUserCmd()
		cmd.SetIn(strings.NewReader(input))
		cmd.SetOut(buf)

		mockClient.EXPECT().Login(mock.Anything, &pb.LoginRequest{
			Login:    "mark",
			Password: "secret",
		}).Return(&pb.AuthResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}, nil)

		cmd.SetArgs([]string{"auth", "-l", "mark"})
		err = cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully logged in as mark")
		_, err = tmp.Seek(0, io.SeekStart)
		require.NoError(t, err)
		tokens, readErr := io.ReadAll(tmp)
		require.NoError(t, readErr)
		want := "{\"access_token\":\"access_token\",\"refresh_token\":\"refresh_token\"}"
		assert.Equal(t, want, string(tokens))
	})

	t.Run("logout", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewUserCmd()
		cmd.SetOut(buf)

		cmd.SetArgs([]string{"logout"})
		err = cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Successfully logged out")
		assert.NoFileExists(t, tmp.Name())
	})
}
