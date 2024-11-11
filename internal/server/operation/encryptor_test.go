package operation_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/operation"
	mocks "gophkeeper.com/mocks/internal_/server/service"
)

func TestEncryptor_VisitLogin(t *testing.T) {
	tests := []struct {
		name        string
		login       *models.Login
		setupMock   func(*mocks.EncryptionService)
		expectError bool
	}{
		{
			name: "successful encryption",
			login: &models.Login{
				Password: []byte("mysecretpassword"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Run(func(_ []byte, dst io.Writer) {
						_, _ = dst.Write([]byte("encryptedpassword"))
					}).
					Return([]byte("encrypteddatakey"), nil)
			},
			expectError: false,
		},
		{
			name: "encryption failure",
			login: &models.Login{
				Password: []byte("mysecretpassword"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Return(nil, errors.New("encryption failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEncryptionService := mocks.NewEncryptionService(t)
			tt.setupMock(mockEncryptionService)

			visitor := operation.NewEncryptor(mockEncryptionService)
			err := visitor.VisitLogin(tt.login)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte("encrypteddatakey"), tt.login.EncryptedDataKey)
				assert.Equal(t, []byte("encryptedpassword"), tt.login.Password)
			}
		})
	}
}

func TestEncryptor_VisitCard(t *testing.T) {
	tests := []struct {
		name           string
		card           *models.Card
		setupMock      func(*mocks.EncryptionService)
		expectError    bool
		expectedNumber []byte
		expectedCvc    []byte
	}{
		{
			name: "successful card encryption",
			card: &models.Card{
				Number: []byte("4111111111111111"),
				CVC:    []byte("123"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Run(func(_ []byte, dst io.Writer) {
						_, _ = dst.Write([]byte("encryptednumber"))
					}).
					Return([]byte("encrypteddatakey"), nil).
					Once()

				m.EXPECT().
					EncryptWithKey(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Run(func(_ []byte, dst io.Writer, _ []byte) {
						_, _ = dst.Write([]byte("encryptedcvc"))
					}).
					Return(nil).
					Once()
			},
			expectError:    false,
			expectedNumber: []byte("encryptednumber"),
			expectedCvc:    []byte("encryptedcvc"),
		},
		{
			name: "card number encryption failure",
			card: &models.Card{
				Number: []byte("4111111111111111"),
				CVC:    []byte("123"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Return(nil, errors.New("number encryption failed"))
			},
			expectError: true,
		},
		{
			name: "cvc encryption failure",
			card: &models.Card{
				Number: []byte("4111111111111111"),
				CVC:    []byte("123"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Return([]byte("encrypteddatakey"), nil)

				m.EXPECT().
					EncryptWithKey(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Return(errors.New("cvc encryption failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.EncryptionService)
			tt.setupMock(mockService)

			visitor := operation.NewEncryptor(mockService)
			err := visitor.VisitCard(tt.card)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte("encrypteddatakey"), tt.card.EncryptedDataKey)
				assert.Equal(t, tt.expectedNumber, tt.card.Number)
				assert.Equal(t, tt.expectedCvc, tt.card.CVC)
			}
		})
	}
}

func TestEncryptor_VisitNote(t *testing.T) {
	tests := []struct {
		name        string
		note        *models.Note
		setupMock   func(*mocks.EncryptionService)
		expectError bool
	}{
		{
			name: "successful encryption",
			note: &models.Note{
				Text: []byte("mysecrettext"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Run(func(_ []byte, dst io.Writer) {
						_, _ = dst.Write([]byte("encryptedtext"))
					}).
					Return([]byte("encrypteddatakey"), nil)
			},
			expectError: false,
		},
		{
			name: "encryption failure",
			note: &models.Note{
				Text: []byte("mysecrettext"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Return(nil, errors.New("encryption failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEncryptionService := mocks.NewEncryptionService(t)
			tt.setupMock(mockEncryptionService)

			visitor := operation.NewEncryptor(mockEncryptionService)
			err := visitor.VisitNote(tt.note)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte("encrypteddatakey"), tt.note.EncryptedDataKey)
				assert.Equal(t, []byte("encryptedtext"), tt.note.Text)
			}
		})
	}
}

func TestEncryptor_VisitBinary(t *testing.T) {
	tests := []struct {
		name        string
		binary      *models.Binary
		setupMock   func(*mocks.EncryptionService)
		expectError bool
	}{
		{
			name: "successful encryption",
			binary: &models.Binary{
				Data: []byte("mysecretdata"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Run(func(_ []byte, dst io.Writer) {
						_, _ = dst.Write([]byte("encrypted"))
					}).
					Return([]byte("encrypteddatakey"), nil)
			},
			expectError: false,
		},
		{
			name: "encryption failure",
			binary: &models.Binary{
				Data: []byte("mysecretdata"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					Encrypt(mock.Anything, mock.Anything).
					Return(nil, errors.New("encryption failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEncryptionService := mocks.NewEncryptionService(t)
			tt.setupMock(mockEncryptionService)

			visitor := operation.NewEncryptor(mockEncryptionService)
			err := visitor.VisitBinary(tt.binary)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte("encrypteddatakey"), tt.binary.EncryptedDataKey)
				assert.Equal(t, []byte("encrypted"), tt.binary.Data)
			}
		})
	}
}
