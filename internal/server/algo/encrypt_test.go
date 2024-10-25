package algo_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gophkeeper.com/internal/server/algo"
	"gophkeeper.com/internal/server/models"
	mocks "gophkeeper.com/mocks/internal_/server/service"
)

func TestEncryptionVisitor_VisitLogin(t *testing.T) {
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
					EncryptStream(mock.Anything, mock.Anything).
					Run(func(_ io.Reader, dst io.Writer) {
						_, _ = dst.Write([]byte("encryptedpassword"))
					}).
					Return([]byte("datakey"), []byte("encrypteddatakey"), nil)
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
					EncryptStream(mock.Anything, mock.Anything).
					Return(nil, nil, errors.New("encryption failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEncryptionService := mocks.NewEncryptionService(t)
			tt.setupMock(mockEncryptionService)

			visitor := algo.NewEncryptionVisitor(mockEncryptionService)
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

func TestEncryptionVisitor_VisitCard(t *testing.T) {
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
				Cvc:    []byte("123"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					EncryptStream(mock.Anything, mock.Anything).
					Run(func(_ io.Reader, dst io.Writer) {
						_, _ = dst.Write([]byte("encryptednumber"))
					}).
					Return([]byte("datakey"), []byte("encrypteddatakey"), nil).
					Once()

				m.EXPECT().
					EncryptStreamWithKey(mock.Anything, mock.Anything, []byte("datakey")).
					Run(func(_ io.Reader, dst io.Writer, _ []byte) {
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
				Cvc:    []byte("123"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					EncryptStream(mock.Anything, mock.Anything).
					Return(nil, nil, errors.New("number encryption failed"))
			},
			expectError: true,
		},
		{
			name: "cvc encryption failure",
			card: &models.Card{
				Number: []byte("4111111111111111"),
				Cvc:    []byte("123"),
			},
			setupMock: func(m *mocks.EncryptionService) {
				m.EXPECT().
					EncryptStream(mock.Anything, mock.Anything).
					Return([]byte("datakey"), []byte("encrypteddatakey"), nil)

				m.EXPECT().
					EncryptStreamWithKey(mock.Anything, mock.Anything, []byte("datakey")).
					Return(errors.New("cvc encryption failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.EncryptionService)
			tt.setupMock(mockService)

			visitor := algo.NewEncryptionVisitor(mockService)
			err := visitor.VisitCard(tt.card)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, []byte("encrypteddatakey"), tt.card.EncryptedDataKey)
				assert.Equal(t, tt.expectedNumber, tt.card.Number)
				assert.Equal(t, tt.expectedCvc, tt.card.Cvc)
			}

			mockService.AssertExpectations(t)
		})
	}
}
