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

func TestDecryptor_VisitLogin(t *testing.T) {
	tests := []struct {
		name         string
		login        *models.Login
		setupMock    func(*mocks.EncryptionService)
		expectError  bool
		expectedPass []byte
	}{
		{
			name: "successful decryption",
			login: &models.Login{
				Password: []byte("encryptedpassword"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Run(func(_ io.Reader, dst io.Writer, _ []byte) {
						_, _ = dst.Write([]byte("decryptedpassword"))
					}).
					Return(nil).
					Once()
			},
			expectError:  false,
			expectedPass: []byte("decryptedpassword"),
		},
		{
			name: "decryption failure",
			login: &models.Login{
				Password: []byte("encryptedpassword"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Return(errors.New("decryption failed")).
					Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewEncryptionService(t)
			tt.setupMock(mockService)

			visitor := operation.NewDecryptor(mockService)
			err := visitor.VisitLogin(tt.login)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot decrypt password")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPass, tt.login.Password)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDecryptor_VisitCard(t *testing.T) {
	tests := []struct {
		name           string
		card           *models.Card
		setupMock      func(*mocks.EncryptionService)
		expectError    bool
		expectedNumber []byte
		expectedCvc    []byte
	}{
		{
			name: "successful card decryption",
			card: &models.Card{
				Number: []byte("encryptednumber"),
				CVC:    []byte("encryptedcvc"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				// First call for number decryption
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Run(func(_ io.Reader, dst io.Writer, _ []byte) {
						_, _ = dst.Write([]byte("decryptednumber"))
					}).
					Return(nil).Once()

				// Second call for CVC decryption
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Run(func(_ io.Reader, dst io.Writer, _ []byte) {
						_, _ = dst.Write([]byte("decryptedcvc"))
					}).
					Return(nil).Once()
			},
			expectError:    false,
			expectedNumber: []byte("decryptednumber"),
			expectedCvc:    []byte("decryptedcvc"),
		},
		{
			name: "card number decryption failure",
			card: &models.Card{
				Number: []byte("encryptednumber"),
				CVC:    []byte("encryptedcvc"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Return(errors.New("number decryption failed")).
					Once()
			},
			expectError: true,
		},
		{
			name: "cvc decryption failure",
			card: &models.Card{
				Number: []byte("encryptednumber"),
				CVC:    []byte("encryptedcvc"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				// Successful number decryption
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Run(func(_ io.Reader, dst io.Writer, _ []byte) {
						_, _ = dst.Write([]byte("decryptednumber"))
					}).
					Return(nil).Once()

				// Failed CVC decryption
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Return(errors.New("cvc decryption failed")).
					Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewEncryptionService(t)
			tt.setupMock(mockService)

			visitor := operation.NewDecryptor(mockService)
			err := visitor.VisitCard(tt.card)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedNumber, tt.card.Number)
				assert.Equal(t, tt.expectedCvc, tt.card.CVC)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDecryptor_VisitNote(t *testing.T) {
	tests := []struct {
		name         string
		note         *models.Note
		setupMock    func(*mocks.EncryptionService)
		expectError  bool
		expectedText []byte
	}{
		{
			name: "successful decryption",
			note: &models.Note{
				Text: []byte("encryptedtext"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Run(func(_ io.Reader, dst io.Writer, _ []byte) {
						_, _ = dst.Write([]byte("decryptedtext"))
					}).
					Return(nil).
					Once()
			},
			expectError:  false,
			expectedText: []byte("decryptedtext"),
		},
		{
			name: "decryption failure",
			note: &models.Note{
				Text: []byte("encryptedtext"),
				SecretMetadata: models.SecretMetadata{
					EncryptedDataKey: []byte("encrypteddatakey"),
				},
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					DecryptStream(mock.Anything, mock.Anything, []byte("encrypteddatakey")).
					Return(errors.New("decryption failed")).
					Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewEncryptionService(t)
			tt.setupMock(mockService)

			visitor := operation.NewDecryptor(mockService)
			err := visitor.VisitNote(tt.note)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot decrypt note")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedText, tt.note.Text)
			}

			mockService.AssertExpectations(t)
		})
	}
}
