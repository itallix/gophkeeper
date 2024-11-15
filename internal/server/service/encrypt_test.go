package service_test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/itallix/gophkeeper/internal/server/service"
	mocks "github.com/itallix/gophkeeper/mocks/internal_/server/service"
)

func TestNewStandardEncryptionService(t *testing.T) {
	mockKMS := mocks.NewKMS(t)
	service := service.NewStandardEncryptionService(mockKMS)

	assert.NotNil(t, service)
}

func TestStandardEncryptionService_Encrypt(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		setupMock     func(*mocks.KMS)
		expectedError bool
	}{
		{
			name:  "successful_encryption",
			input: []byte("test data"),
			setupMock: func(m *mocks.KMS) {
				dataKey := make([]byte, 32)
				encDataKey := []byte("encrypted-key")
				_, _ = rand.Read(dataKey)
				m.EXPECT().
					GenerateDataKey().
					Return(dataKey, encDataKey, nil)
			},
			expectedError: false,
		},
		{
			name:  "kms_error",
			input: []byte("test data"),
			setupMock: func(m *mocks.KMS) {
				m.EXPECT().
					GenerateDataKey().
					Return([]byte{}, []byte{}, errors.New("kms error"))
			},
			expectedError: true,
		},
		{
			name:  "empty_input",
			input: []byte{},
			setupMock: func(m *mocks.KMS) {
				dataKey := make([]byte, 32)
				encDataKey := []byte("encrypted-key")
				_, _ = rand.Read(dataKey)
				m.EXPECT().
					GenerateDataKey().
					Return(dataKey, encDataKey, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKMS := mocks.NewKMS(t)
			tt.setupMock(mockKMS)

			encService := service.NewStandardEncryptionService(mockKMS)
			var dst bytes.Buffer

			encryptedKey, err := encService.Encrypt(tt.input, &dst)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, encryptedKey)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, encryptedKey)
				assert.NotEmpty(t, dst.Bytes())
				assert.NotEqual(t, tt.input, dst.Bytes())
			}
		})
	}
}

func TestStandardEncryptionService_EncryptWithKey(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		encryptedKey  []byte
		setupMock     func(*mocks.KMS)
		expectedError bool
	}{
		{
			name:         "successful_encryption",
			input:        []byte("test data"),
			encryptedKey: []byte("encrypted-key"),
			setupMock: func(m *mocks.KMS) {
				dataKey := make([]byte, 32)
				_, _ = rand.Read(dataKey)
				m.EXPECT().
					DecryptDataKey(mock.Anything).
					Return(dataKey, nil)
			},
			expectedError: false,
		},
		{
			name:         "kms_decryption_error",
			input:        []byte("test data"),
			encryptedKey: []byte("encrypted-key"),
			setupMock: func(m *mocks.KMS) {
				m.EXPECT().
					DecryptDataKey(mock.Anything).
					Return(nil, errors.New("decryption error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKMS := mocks.NewKMS(t)
			tt.setupMock(mockKMS)

			service := service.NewStandardEncryptionService(mockKMS)
			var dst bytes.Buffer

			err := service.EncryptWithKey(tt.input, &dst, tt.encryptedKey)

			if tt.expectedError {
				require.Error(t, err)
				assert.Empty(t, dst.Bytes())
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, dst.Bytes())
				assert.NotEqual(t, tt.input, dst.Bytes())
			}
		})
	}
}

func TestStandardEncryptionService_Decrypt(t *testing.T) {
	// Helper function to create encrypted data
	createEncryptedData := func(data []byte, key []byte) []byte {
		block, _ := aes.NewCipher(key)
		gcm, _ := cipher.NewGCM(block)
		nonce := make([]byte, gcm.NonceSize())
		_, _ = rand.Read(nonce)
		return gcm.Seal(nonce, nonce, data, nil)
	}

	tests := []struct {
		name          string
		setupData     func() ([]byte, []byte, []byte) // returns input, dataKey, encryptedKey
		setupMock     func(*mocks.KMS, []byte)
		expectedError bool
	}{
		{
			name: "successful_decryption",
			setupData: func() ([]byte, []byte, []byte) {
				originalData := []byte("test data")
				dataKey := make([]byte, 32)
				_, _ = rand.Read(dataKey)
				encryptedData := createEncryptedData(originalData, dataKey)
				return encryptedData, dataKey, []byte("encrypted-key")
			},
			setupMock: func(m *mocks.KMS, dataKey []byte) {
				m.EXPECT().
					DecryptDataKey(mock.Anything).
					Return(dataKey, nil)
			},
			expectedError: false,
		},
		{
			name: "kms_decryption_error",
			setupData: func() ([]byte, []byte, []byte) {
				return []byte("invalid data"), nil, []byte("encrypted-key")
			},
			setupMock: func(m *mocks.KMS, _ []byte) {
				m.EXPECT().
					DecryptDataKey(mock.Anything).
					Return(nil, errors.New("decryption error"))
			},
			expectedError: true,
		},
		{
			name: "invalid_ciphertext",
			setupData: func() ([]byte, []byte, []byte) {
				dataKey := make([]byte, 32)
				_, _ = rand.Read(dataKey)
				return []byte("invalid ciphertext"), dataKey, []byte("encrypted-key")
			},
			setupMock: func(m *mocks.KMS, dataKey []byte) {
				m.EXPECT().
					DecryptDataKey(mock.Anything).
					Return(dataKey, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKMS := mocks.NewKMS(t)
			encryptedData, dataKey, encryptedKey := tt.setupData()
			tt.setupMock(mockKMS, dataKey)

			service := service.NewStandardEncryptionService(mockKMS)
			var dst bytes.Buffer

			err := service.Decrypt(encryptedData, &dst, encryptedKey)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, dst.Bytes())
			}

			mockKMS.AssertExpectations(t)
		})
	}
}
