package operation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/operation/sql"
	"gophkeeper.com/internal/server/service"
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

// func (b *ProcessorBuilder) WithValidation() *ProcessorBuilder {
//     b.visitors = append(b.visitors, NewValidator(ctx))
//     return b
// }

func (b *ProcessorBuilder) WithEncryption(service service.EncryptionService) *ProcessorBuilder {
	b.visitors = append(b.visitors, NewEncryptor(service))
	return b
}

func (b *ProcessorBuilder) WithDecryption(service service.EncryptionService) *ProcessorBuilder {
	b.visitors = append(b.visitors, NewDecryptor(service))
	return b
}

func (b *ProcessorBuilder) WithStorageCreator(ctx context.Context, pool *pgxpool.Pool) *ProcessorBuilder {
	b.visitors = append(b.visitors, sql.NewStorageCreator(ctx, pool))
	return b
}

func (b *ProcessorBuilder) WithStorageDeleter(ctx context.Context, pool *pgxpool.Pool) *ProcessorBuilder {
	b.visitors = append(b.visitors, sql.NewStorageDeleter(ctx, pool))
	return b
}

func (b *ProcessorBuilder) WithStorageLister(ctx context.Context, pool *pgxpool.Pool) *ProcessorBuilder {
	b.visitors = append(b.visitors, sql.NewStorageLister(ctx, pool))
	return b
}

func (b *ProcessorBuilder) WithStorageRetriever(ctx context.Context, pool *pgxpool.Pool) *ProcessorBuilder {
	b.visitors = append(b.visitors, sql.NewStorageRetriever(ctx, pool))
	return b
}

func (b *ProcessorBuilder) Build() *SecretProcessor {
	return NewSecretProcessor(b.visitors...)
}
