// Package server provides secure storage and management of sensitive data through
// a visitor-based processing system.
package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/operation"
	"gophkeeper.com/internal/server/s3"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"
)

// Vault defines the interface for secure secret management operations.
// It provides methods for storing, retrieving, and managing different types of secrets
// while handling encryption and secure storage automatically.
type Vault interface {
	StoreSecret(secret models.Secret) error
	RetrieveSecret(secret models.Secret) error
	DeleteSecret(secret models.Secret) error
	ListSecrets(secret models.Secret) ([]string, error)
}

// VaultImpl implements the Vault interface using a combination of database storage
// for metadata and object storage for binary data. It provides secure secret management
// with encryption at rest.
type VaultImpl struct {
	ctx               context.Context
	pool              *pgxpool.Pool
	objectStorage     *s3.ObjectStorage
	encryptionService service.EncryptionService
}

// NewVault creates and initializes a new Vault instance with the provided dependencies.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of operations
//   - pool: PostgreSQL connection pool for database operations
//   - objectStorage: S3-compatible storage for binary data
//   - encryptionService: Service for encrypting and decrypting sensitive data
//
// Returns:
//   - *Vault: A new instance of Vault initialized with the provided dependencies
func NewVaultImpl(ctx context.Context, pool *pgxpool.Pool, objectStorage *s3.ObjectStorage,
	encryptionService service.EncryptionService) *VaultImpl {
	return &VaultImpl{
		ctx:               ctx,
		pool:              pool,
		objectStorage:     objectStorage,
		encryptionService: encryptionService,
	}
}

// StoreSecret securely stores a secret in the vault. The secret is validated,
// encrypted, and then stored using the appropriate storage mechanism based on its type.
//
// Parameters:
//   - secret: The secret to be stored, implementing the models.Secret interface
//
// Returns:
//   - error: nil if successful, otherwise an error describing what went wrong
func (v *VaultImpl) StoreSecret(secret models.Secret) error {
	op := operation.NewProcessorBuilder().
		WithValidation().
		WithEncryption(v.encryptionService).
		WithStorageCreator(v.ctx, v.pool, v.objectStorage).
		Build()

	if err := op.Process(secret); err != nil {
		return err
	}
	return nil
}

// RetrieveSecret fetches and decrypts a previously stored secret from the vault.
// The secret is retrieved from storage and decrypted using the encryption service.
//
// Parameters:
//   - secret: A secret object containing the necessary metadata for retrieval
//
// Returns:
//   - error: nil if successful, otherwise an error describing what went wrong
func (v *VaultImpl) RetrieveSecret(secret models.Secret) error {
	op := operation.NewProcessorBuilder().
		WithStorageRetriever(v.ctx, v.pool, v.objectStorage).
		WithDecryption(v.encryptionService).
		Build()

	if err := op.Process(secret); err != nil {
		return err
	}
	return nil
}

// DeleteSecret removes a secret from the vault, cleaning up both database
// and object storage records as appropriate.
//
// Parameters:
//   - secret: The secret to be deleted, containing necessary metadata
//
// Returns:
//   - error: nil if successful, otherwise an error describing what went wrong
func (v *VaultImpl) DeleteSecret(secret models.Secret) error {
	deleter := storage.NewDeleter(v.ctx, v.pool, v.objectStorage)

	if err := secret.Accept(deleter); err != nil {
		return err
	}
	return nil
}

// ListSecrets retrieves a list of secret identifiers of a specific type
// stored in the vault.
//
// Parameters:
//   - secret: A secret object indicating the type of secrets to list
//
// Returns:
//   - []string: A slice of secret identifiers
//   - error: nil if successful, otherwise an error describing what went wrong
func (v *VaultImpl) ListSecrets(secret models.Secret) ([]string, error) {
	lister := storage.NewLister(v.ctx, v.pool)

	if err := secret.Accept(lister); err != nil {
		return nil, err
	}

	return lister.GetResult().([]string), nil
}
