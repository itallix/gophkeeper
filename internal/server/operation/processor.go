package operation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/itallix/gophkeeper/internal/server/models"
	"github.com/itallix/gophkeeper/internal/server/s3"
	"github.com/itallix/gophkeeper/internal/server/service"
	"github.com/itallix/gophkeeper/internal/server/storage"
)

type SecretProcessor struct {
	visitors []models.SecretVisitor
}

func NewSecretProcessor(visitors ...models.SecretVisitor) *SecretProcessor {
	return &SecretProcessor{visitors: visitors}
}

func (p *SecretProcessor) Process(secret models.Secret) error {
	for _, visitor := range p.visitors {
		if err := secret.Accept(visitor); err != nil {
			return fmt.Errorf("processing error at %T: %w", visitor, err)
		}
	}
	return nil
}

type ProcessorBuilder struct {
	visitors []models.SecretVisitor
}

func NewProcessorBuilder(visitors ...models.SecretVisitor) *ProcessorBuilder {
	return &ProcessorBuilder{visitors: visitors}
}

func (b *ProcessorBuilder) WithValidation() *ProcessorBuilder {
	b.visitors = append(b.visitors, NewValidator())
	return b
}

func (b *ProcessorBuilder) WithEncryption(service service.EncryptionService) *ProcessorBuilder {
	b.visitors = append(b.visitors, NewEncryptor(service))
	return b
}

func (b *ProcessorBuilder) WithDecryption(service service.EncryptionService) *ProcessorBuilder {
	b.visitors = append(b.visitors, NewDecryptor(service))
	return b
}

func (b *ProcessorBuilder) WithStorageCreator(ctx context.Context, pool *pgxpool.Pool,
	objectStorage *s3.ObjectStorage) *ProcessorBuilder {
	b.visitors = append(b.visitors, storage.NewCreator(ctx, pool, objectStorage))
	return b
}

func (b *ProcessorBuilder) WithStorageDeleter(ctx context.Context, pool *pgxpool.Pool,
	objectStorage *s3.ObjectStorage) *ProcessorBuilder {
	b.visitors = append(b.visitors, storage.NewDeleter(ctx, pool, objectStorage))
	return b
}

func (b *ProcessorBuilder) WithStorageLister(ctx context.Context, pool *pgxpool.Pool) *ProcessorBuilder {
	b.visitors = append(b.visitors, storage.NewLister(ctx, pool))
	return b
}

func (b *ProcessorBuilder) WithStorageRetriever(ctx context.Context, pool *pgxpool.Pool,
	objectStorage *s3.ObjectStorage) *ProcessorBuilder {
	b.visitors = append(b.visitors, storage.NewRetriever(ctx, pool, objectStorage))
	return b
}

func (b *ProcessorBuilder) Build() *SecretProcessor {
	return NewSecretProcessor(b.visitors...)
}
