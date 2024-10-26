package operation_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/operation"
)

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name     string
		login    *models.Login
		wantErr  bool
		errCount int // number of expected errors
	}{
		{
			name: "valid login and password",
			login: &models.Login{
				Login:    "johndoe",
				Password: []byte("securepass123"),
			},
			wantErr: false,
		},
		{
			name: "short login",
			login: &models.Login{
				Login:    "jo",
				Password: []byte("securepass123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "short password",
			login: &models.Login{
				Login:    "johndoe",
				Password: []byte("pass"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "both short login and password",
			login: &models.Login{
				Login:    "jo",
				Password: []byte("pass"),
			},
			wantErr:  true,
			errCount: 2,
		},
	}

	v := operation.NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.VisitLogin(tt.login)
			if tt.wantErr {
				require.Error(t, err)
				errs := strings.Split(err.Error(), "\n")
				assert.Equal(t, errs, tt.errCount)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVisitCard(t *testing.T) {
	currentYear := int16(time.Now().Year())
	currentMonth := int8(time.Now().Month())

	tests := []struct {
		name     string
		card     *models.Card
		wantErr  bool
		errCount int
	}{
		{
			name: "valid future date",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 2,
				CVC:         []byte("123"),
			},
			wantErr: false,
		},
		{
			name: "invalid month zero",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: 0,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "invalid month > 12",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: 13,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "past year",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: currentMonth,
				ExpiryYear:  currentYear - 1,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{

			name: "too far future",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: currentMonth,
				ExpiryYear:  currentYear + 21,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "expired card same year",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: currentMonth - 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "valid card",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: currentMonth + 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr: false,
		},
		{
			name: "short card number",
			card: &models.Card{
				Number:      []byte("123456789012"),
				ExpiryMonth: currentMonth + 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "long card number",
			card: &models.Card{
				Number:      []byte("12345678901234567890"),
				ExpiryMonth: currentMonth + 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "non-digit card number",
			card: &models.Card{
				Number:      []byte("4532015112a30366"),
				ExpiryMonth: currentMonth + 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("123"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "invalid CVC length",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: currentMonth + 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("12345"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "non-digit CVC",
			card: &models.Card{
				Number:      []byte("4532015112830366"),
				ExpiryMonth: currentMonth + 1,
				ExpiryYear:  currentYear,
				CVC:         []byte("12a"),
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "multiple validation errors",
			card: &models.Card{
				Number:      []byte("123"), // short number
				ExpiryMonth: 13,            // invalid month
				ExpiryYear:  currentYear,
				CVC:         []byte("12a"), // non-digit CVC
			},
			wantErr:  true,
			errCount: 3,
		},
	}

	v := operation.NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.VisitCard(tt.card)
			if tt.wantErr {
				require.Error(t, err)
				errs := strings.Split(err.Error(), "\n")
				assert.Len(t, errs, tt.errCount)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
