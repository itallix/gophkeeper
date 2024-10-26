package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/operation"
	"gophkeeper.com/internal/server/operation/storage"
	"gophkeeper.com/internal/server/service"
)

type Vault struct {
	ctx           context.Context
	pool          *pgxpool.Pool
	encryptionSrv service.EncryptionService
}

func NewVault(ctx context.Context, pool *pgxpool.Pool, encryptionSrv service.EncryptionService) *Vault {
	return &Vault{
		ctx:           ctx,
		pool:          pool,
		encryptionSrv: encryptionSrv,
	}
}

func (v *Vault) StoreSecret(user, token string, secret models.Secret) error {
	op := operation.NewProcessorBuilder().
		WithValidation().
		WithEncryption(v.encryptionSrv).
		WithStorageCreator(v.ctx, v.pool).
		Build()

	if err := op.Process(secret); err != nil {
		return err
	}
	return nil
}

func (v *Vault) RetrieveSecret(user, token string, secret models.Secret) error {
	op := operation.NewProcessorBuilder().
		WithStorageRetriever(v.ctx, v.pool).
		WithDecryption(v.encryptionSrv).
		Build()

	if err := op.Process(secret); err != nil {
		return err
	}
	return nil
}

func (v *Vault) DeleteSecret(user, token string, secret models.Secret) error {
	deleter := storage.NewDeleter(v.ctx, v.pool)

	if err := secret.Accept(deleter); err != nil {
		return err
	}
	return nil
}

func (v *Vault) ListSecrets(user, token string, secret models.Secret) ([]string, error) {
	lister := storage.NewLister(v.ctx, v.pool)

	if err := secret.Accept(lister); err != nil {
		return nil, err
	}

	return lister.GetResult().([]string), nil
}

// // Updated StoreFile method in Vault
// func (v *Vault) StoreFile(username, token, path string, data io.Reader, metadata models.SecretMetadata) error {
// 	// ... (previous authentication and access control checks remain)

// 	// Create a pipe for streaming encryption
// 	pr, pw := io.Pipe()
// 	var encryptedDataKey []byte
// 	var encryptionErr error

// 	go func() {
// 		defer pw.Close()
// 		encryptedDataKey, encryptionErr = v.encryptionService.EncryptStream(data, pw)
// 	}()

// 	// Store the encrypted file
// 	err := v.secretStore.StoreFile(path, pr, metadata)
// 	if err != nil {
// 		return err
// 	}

// 	if encryptionErr != nil {
// 		return encryptionErr
// 	}

// 	// Update metadata with encrypted data key
// 	metadata.EncryptedDataKey = encryptedDataKey
// 	err = v.secretStore.UpdateMetadata(path, metadata)
// 	if err != nil {
// 		return err
// 	}

// 	return v.auditLogService.Log(username, "store_file", path, true)
// }

// // Updated RetrieveFile method in Vault
// func (v *Vault) RetrieveFile(username, token, path string, version int) (io.ReadCloser, SecretMetadata, error) {
// 	// ... (previous authentication and access control checks remain)

// 	// Retrieve the encrypted file and metadata
// 	encryptedReader, metadata, err := v.secretStore.RetrieveFile(path, version)
// 	if err != nil {
// 		return nil, models.SecretMetadata{}, err
// 	}

// 	// Create a pipe for streaming decryption
// 	pr, pw := io.Pipe()
// 	go func() {
// 		defer pw.Close()
// 		err := v.encryptionService.Decrypt(encryptedReader, pw, metadata.EncryptedDataKey)
// 		if err != nil {
// 			pw.CloseWithError(err)
// 		}
// 	}()

// 	v.auditLogService.Log(username, "retrieve_file", fmt.Sprintf("%s:v%d", path, version), true)
// 	return pr, metadata, nil
// }
