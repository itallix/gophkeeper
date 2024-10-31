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

func WithCreatedBy(createdBy string) SecretOption {
	return func(o *SecretOptions) {
		o.CreatedBy = createdBy
	}
}

func WithModifiedBy(modifiedBy string) SecretOption {
	return func(o *SecretOptions) {
		o.ModifiedBy = modifiedBy
	}
}

// Login-specific options.
type LoginOptions struct {
	Login    string
	Password string
	URL      string

	SecretOptions
}

type LoginOption func(*LoginOptions)

func WithLogin(login string) LoginOption {
	return func(o *LoginOptions) {
		o.Login = login
	}
}

func WithPassword(password string) LoginOption {
	return func(o *LoginOptions) {
		o.Password = password
	}
}

// Card-specific options.
type CardOptions struct {
	Number      string
	CVC         string
	ExpiryMonth int8
	ExpiryYear  int16
	CardHolder  string

	SecretOptions
}

type CardOption func(*CardOptions)

func WithCardNumber(number string) CardOption {
	return func(o *CardOptions) {
		o.Number = number
	}
}

func WithCVC(cvc string) CardOption {
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

// Note-specific options.
type NoteOptions struct {
	Text string

	SecretOptions
}

type NoteOption func(*NoteOptions)

func WithText(text string) NoteOption {
	return func(o *NoteOptions) {
		o.Text = text
	}
}

// Binary-specific options.
type BinaryOptions struct {
	ChunkID int64
	Chunks  int16
	Hash    string
	Data    []byte

	SecretOptions
}

type BinaryOption func(*BinaryOptions)

func WithChunkID(chunkID int64) BinaryOption {
	return func(o *BinaryOptions) {
		o.ChunkID = chunkID
	}
}

func WithChunks(chunks int16) BinaryOption {
	return func(o *BinaryOptions) {
		o.Chunks = chunks
	}
}

func WithHash(hash string) BinaryOption {
	return func(o *BinaryOptions) {
		o.Hash = hash
	}
}

func WithData(data []byte) BinaryOption {
	return func(o *BinaryOptions) {
		o.Data = data
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
			CreatedBy:        options.CreatedBy,
			ModifiedBy:       options.ModifiedBy,
		},
		Login:    options.Login,
		Password: []byte(options.Password),
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
			CreatedBy:        options.CreatedBy,
			ModifiedBy:       options.ModifiedBy,
		},
		Number:         []byte(options.Number),
		CVC:            []byte(options.CVC),
		ExpiryMonth:    options.ExpiryMonth,
		ExpiryYear:     options.ExpiryYear,
		CardholderName: options.CardHolder,
	}
}

func NewNote(commonOpts []SecretOption, noteOpts []NoteOption) *Note {
	// Initialize with defaults
	options := &NoteOptions{
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

	// Apply note-specific options
	for _, opt := range noteOpts {
		opt(options)
	}

	// Create note
	return &Note{
		SecretMetadata: SecretMetadata{
			Path:             options.Path,
			CreatedAt:        options.CreatedAt,
			ModifiedAt:       options.ModifiedAt,
			EncryptedDataKey: options.EncryptedDataKey,
			CustomMeta:       options.CustomMetadata,
			CreatedBy:        options.CreatedBy,
			ModifiedBy:       options.ModifiedBy,
		},
		Text: []byte(options.Text),
	}
}

func NewBinary(commonOpts []SecretOption, binaryOpts []BinaryOption) *Binary {
	// Initialize with defaults
	options := &BinaryOptions{
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

	// Apply binary-specific options
	for _, opt := range binaryOpts {
		opt(options)
	}

	// Create note
	return &Binary{
		SecretMetadata: SecretMetadata{
			Path:             options.Path,
			CreatedAt:        options.CreatedAt,
			ModifiedAt:       options.ModifiedAt,
			EncryptedDataKey: options.EncryptedDataKey,
			CustomMeta:       options.CustomMetadata,
			CreatedBy:        options.CreatedBy,
			ModifiedBy:       options.ModifiedBy,
		},
		ChunkID: options.ChunkID,
		Chunks:  options.Chunks,
		Hash:    options.Hash,
		Data:    options.Data,
	}
}
