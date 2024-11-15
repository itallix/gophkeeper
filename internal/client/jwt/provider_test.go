package jwt_test

import (
	"io/fs"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itallix/gophkeeper/internal/client/jwt"
)

func TestNewToken(t *testing.T) {
	accessToken := "access123"
	refreshToken := "refresh123"

	token := jwt.NewToken(accessToken, refreshToken)

	assert.Equal(t, accessToken, token.AccessToken)
	assert.Equal(t, refreshToken, token.RefreshToken)
}

func TestSaveAndLoadToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "token_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	tokenProvider := jwt.NewTokenProvider(tmpFile.Name())

	accessToken := "test_access_token"
	refreshToken := "test_refresh_token"
	tokenData := jwt.NewToken(accessToken, refreshToken)

	err = tokenProvider.SaveToken(tokenData)
	require.NoError(t, err)

	info, err := os.Stat(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, fs.FileMode(0600), info.Mode().Perm())

	loadedToken, err := tokenProvider.LoadToken()
	require.NoError(t, err)

	if !reflect.DeepEqual(tokenData, loadedToken) {
		t.Errorf("expected loaded token %+v, got %+v", tokenData, loadedToken)
	}
}

func TestSaveTokenError(t *testing.T) {
	tokenProvider := jwt.NewTokenProvider("/invalid_path/token.json")

	tokenData := jwt.NewToken("access_token", "refresh_token")
	err := tokenProvider.SaveToken(tokenData)
	require.Error(t, err)
}

func TestLoadTokenError(t *testing.T) {
	tokenProvider := jwt.NewTokenProvider("non_existent_file.json")

	_, err := tokenProvider.LoadToken()
	require.Error(t, err)
}

func TestLoadTokenInvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "token_invalid_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	invalidJSON := []byte(`{invalid_json}`)
	err = os.WriteFile(tmpFile.Name(), invalidJSON, 0600)
	require.NoError(t, err)

	tokenProvider := jwt.NewTokenProvider(tmpFile.Name())
	_, err = tokenProvider.LoadToken()
	require.Error(t, err)
}
