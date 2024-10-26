package models

import "time"

type VaultItemType string

const (
	CardType   VaultItemType = "card"
	LoginType  VaultItemType = "login"
	NoteType   VaultItemType = "note"
	BinaryType VaultItemType = "binary"
)

type SecretOptions struct {
	Path             string
	CreatedAt        time.Time
	ModifiedAt       time.Time
	EncryptedDataKey []byte
	CustomMetadata   map[string]string
	CreatedBy        string
	ModifiedBy       string
}

type SecretOption func(*SecretOptions)

func WithPath(path string) SecretOption {
	return func(o *SecretOptions) {
		o.Path = path
	}
}

func WithEncryptedDataKey(key []byte) SecretOption {
	return func(o *SecretOptions) {
		o.EncryptedDataKey = key
	}
}

func WithCustomMetadata(metadata map[string]string) SecretOption {
	return func(o *SecretOptions) {
		o.CustomMetadata = metadata
	}
}

func WithCreatedAt(t time.Time) SecretOption {
	return func(o *SecretOptions) {
		o.CreatedAt = t
	}
}

// Login-specific options.
type LoginOptions struct {
	SecretOptions
	Login    string
	Password []byte
	URL      string
}

type LoginOption func(*LoginOptions)

func WithLogin(login string) LoginOption {
	return func(o *LoginOptions) {
		o.Login = login
	}
}

func WithPassword(password []byte) LoginOption {
	return func(o *LoginOptions) {
		o.Password = password
	}
}

// Card-specific options.
type CardOptions struct {
	SecretOptions
	Number      []byte
	CVC         []byte
	ExpiryMonth int8
	ExpiryYear  int16
	CardHolder  string
}

type CardOption func(*CardOptions)

func WithCardNumber(number []byte) CardOption {
	return func(o *CardOptions) {
		o.Number = number
	}
}

func WithCVC(cvc []byte) CardOption {
	return func(o *CardOptions) {
		o.CVC = cvc
	}
}

func WithExpiry(month int8, year int16) CardOption {
	return func(o *CardOptions) {
		o.ExpiryMonth = month
		o.ExpiryYear = year
	}
}

func WithCardHolder(name string) CardOption {
	return func(o *CardOptions) {
		o.CardHolder = name
	}
}

// Factory functions.
func NewLogin(commonOpts []SecretOption, loginOpts []LoginOption) *Login {
	// Initialize with defaults
	options := &LoginOptions{
		SecretOptions: SecretOptions{
			CreatedAt:      time.Now(),
			ModifiedAt:     time.Now(),
			CustomMetadata: make(map[string]string),
		},
	}

	// Apply common options
	for _, opt := range commonOpts {
		opt(&options.SecretOptions)
	}

	// Apply login-specific options
	for _, opt := range loginOpts {
		opt(options)
	}

	// Create login
	return &Login{
		SecretMetadata: SecretMetadata{
			Path:             options.Path,
			CreatedAt:        options.CreatedAt,
			ModifiedAt:       options.ModifiedAt,
			EncryptedDataKey: options.EncryptedDataKey,
			CustomMeta:       options.CustomMetadata,
		},
		Login:    options.Login,
		Password: options.Password,
	}
}

func NewCard(commonOpts []SecretOption, cardOpts []CardOption) *Card {
	options := &CardOptions{
		SecretOptions: SecretOptions{
			CreatedAt:      time.Now(),
			ModifiedAt:     time.Now(),
			CustomMetadata: make(map[string]string),
		},
	}

	for _, opt := range commonOpts {
		opt(&options.SecretOptions)
	}

	for _, opt := range cardOpts {
		opt(options)
	}

	return &Card{
		SecretMetadata: SecretMetadata{
			Path:             options.Path,
			CreatedAt:        options.CreatedAt,
			ModifiedAt:       options.ModifiedAt,
			EncryptedDataKey: options.EncryptedDataKey,
			CustomMeta:       options.CustomMetadata,
		},
		Number:         options.Number,
		CVC:            options.CVC,
		ExpiryMonth:    options.ExpiryMonth,
		ExpiryYear:     options.ExpiryYear,
		CardholderName: options.CardHolder,
	}
}

func NewVaultItem(vaultType VaultItemType, path string) Secret {
	switch vaultType {
	case LoginType:
		return NewLogin(
			[]SecretOption{
				WithPath(path),
			},
			[]LoginOption{},
		)
	case CardType:
		return NewCard(
			[]SecretOption{
				WithPath(path),
			},
			[]CardOption{},
		)
	}
	return nil
}
