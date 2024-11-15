package service_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itallix/gophkeeper/internal/server/service"
)

func TestNewRSAKMS(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name             string
		setupFunc        func() (string, string, error)
		expectedErrorMsg string
	}{
		{
			name: "valid_keys",
			setupFunc: func() (string, string, error) {
				masterKeyPath, encKeyPath, err := generateTestKeys(tmpDir)
				return masterKeyPath, encKeyPath, err
			},
			expectedErrorMsg: "",
		},
		{
			name: "missing_master_key",
			setupFunc: func() (string, string, error) {
				return filepath.Join(tmpDir, "nonexistent.pem"), filepath.Join(tmpDir, "enc.key"), nil
			},
			expectedErrorMsg: "failed to read master key file",
		},
		{
			name: "invalid_master_key_format",
			setupFunc: func() (string, string, error) {
				masterKeyPath := filepath.Join(tmpDir, "invalid.pem")
				encKeyPath := filepath.Join(tmpDir, "enc.key")
				err := os.WriteFile(masterKeyPath, []byte("invalid key"), 0600)
				return masterKeyPath, encKeyPath, err
			},
			expectedErrorMsg: "failed to decode PEM block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masterKeyPath, encKeyPath, err := tt.setupFunc()
			require.NoError(t, err, "setup failed")

			kms, err := service.NewRSAKMS(masterKeyPath, encKeyPath)
			if tt.expectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Nil(t, kms)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, kms)
			}
		})
	}
}

func TestRSAKMS_GenerateDataKey(t *testing.T) {
	kms := setupTestKMS(t)

	for i := 0; i < 5; i++ {
		plainKey, encKey, err := kms.GenerateDataKey()
		require.NoError(t, err)

		assert.Len(t, plainKey, service.DataKeyLength)
		assert.NotEmpty(t, encKey)
		assert.NotEqual(t, plainKey, encKey)

		decryptedKey, err := kms.DecryptDataKey(encKey)
		require.NoError(t, err)
		assert.Equal(t, plainKey, decryptedKey)
	}
}

func TestRSAKMS_DecryptDataKey(t *testing.T) {
	kms := setupTestKMS(t)

	tests := []struct {
		name          string
		encryptedKey  []byte
		expectedError string
	}{
		{
			name: "valid_key",
			encryptedKey: func() []byte {
				key := make([]byte, service.DataKeyLength)
				_, _ = rand.Read(key)
				encKey, _ := service.EncryptAES(key, kms.EncryptionKey)
				return encKey
			}(),
			expectedError: "",
		},
		{
			name:          "invalid_key",
			encryptedKey:  []byte("invalid encrypted key"),
			expectedError: "authentication failed",
		},
		{
			name:          "empty_key",
			encryptedKey:  []byte{},
			expectedError: "ciphertext too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptedKey, err := kms.DecryptDataKey(tt.encryptedKey)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, decryptedKey)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, decryptedKey)
				assert.Len(t, decryptedKey, service.DataKeyLength)
			}
		})
	}
}

// Helper functions

func setupTestKMS(t *testing.T) *service.RSAKMS {
	tmpDir := t.TempDir()
	masterKeyPath, encKeyPath, err := generateTestKeys(tmpDir)
	require.NoError(t, err)

	kms, err := service.NewRSAKMS(masterKeyPath, encKeyPath)
	require.NoError(t, err)
	return kms
}

func generateTestKeys(tmpDir string) (string, string, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Save private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	masterKeyPath := filepath.Join(tmpDir, "master.pem")
	masterKeyFile, err := os.Create(masterKeyPath)
	if err != nil {
		return "", "", err
	}
	defer masterKeyFile.Close()

	err = pem.Encode(masterKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err != nil {
		return "", "", err
	}

	// Generate and encrypt a random key
	encKey := make([]byte, 32)
	if _, err = rand.Read(encKey); err != nil {
		return "", "", err
	}

	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&privateKey.PublicKey,
		encKey,
		nil,
	)
	if err != nil {
		return "", "", err
	}

	encKeyPath := filepath.Join(tmpDir, "enc.key")
	if err = os.WriteFile(encKeyPath, encryptedKey, 0600); err != nil {
		return "", "", err
	}

	return masterKeyPath, encKeyPath, nil
}

// Test AES encryption/decryption functions directly.
func TestAESEncryptionDecryption(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	tests := []struct {
		name        string
		plaintext   []byte
		shouldError bool
	}{
		{
			name:        "normal_data",
			plaintext:   []byte("test data"),
			shouldError: false,
		},
		{
			name:        "empty_data",
			plaintext:   []byte(nil),
			shouldError: false,
		},
		{
			name:        "large_data",
			plaintext:   make([]byte, 1024),
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, encErr := service.EncryptAES(tt.plaintext, key)
			if tt.shouldError {
				assert.Error(t, encErr)
				return
			}
			require.NoError(t, encErr)
			assert.NotEqual(t, tt.plaintext, ciphertext)

			// Decrypt
			decrypted, decErr := service.DecryptAES(ciphertext, key)
			require.NoError(t, decErr)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}
