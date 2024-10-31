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

type Vault struct {
	ctx               context.Context
	pool              *pgxpool.Pool
	objectStorage     *s3.ObjectStorage
	encryptionService service.EncryptionService
}

func NewVault(ctx context.Context, pool *pgxpool.Pool, objectStorage *s3.ObjectStorage,
	encryptionService service.EncryptionService) *Vault {
	return &Vault{
		ctx:               ctx,
		pool:              pool,
		objectStorage:     objectStorage,
		encryptionService: encryptionService,
	}
}

func (v *Vault) StoreSecret(secret models.Secret) error {
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

func (v *Vault) RetrieveSecret(secret models.Secret) error {
	op := operation.NewProcessorBuilder().
		WithStorageRetriever(v.ctx, v.pool).
		WithDecryption(v.encryptionService).
		Build()

	if err := op.Process(secret); err != nil {
		return err
	}
	return nil
}

func (v *Vault) DeleteSecret(secret models.Secret) error {
	deleter := storage.NewDeleter(v.ctx, v.pool)

	if err := secret.Accept(deleter); err != nil {
		return err
	}
	return nil
}

func (v *Vault) ListSecrets(secret models.Secret) ([]string, error) {
	lister := storage.NewLister(v.ctx, v.pool)

	if err := secret.Accept(lister); err != nil {
		return nil, err
	}

	return lister.GetResult().([]string), nil
}
