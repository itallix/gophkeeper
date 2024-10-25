package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/logger"
	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/operation"
	"gophkeeper.com/internal/server/operation/sql"
	"gophkeeper.com/internal/server/service"
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}

	ctx := context.Background()
	dsn := "postgres://postgres:P@ssw0rd@localhost/gophkeeper?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to initialize connection pool: %s", err)
	}

	kms, err := service.NewRSAKMS()
	if err != nil {
		log.Fatalf("Failed to initialize kms: %s", err)
	}
	encryptionService := service.NewStreamEncryptionService(kms)
	login := &models.Login{
		SecretMetadata: models.SecretMetadata{
			Path:       "login1",
			CustomMeta: map[string]string{"attr1": "foo", "attr2": "boo"},
			CreatedBy:  "vitalii",
			ModifiedBy: "vitalii",
		},
		Login:    "vitalii",
		Password: []byte("geheim"),
	}
	card := &models.Card{
		SecretMetadata: models.SecretMetadata{
			Path:       "card1",
			CustomMeta: map[string]string{"attr1": "foo", "attr2": "boo"},
			CreatedBy:  "vitalii",
			ModifiedBy: "vitalii",
		},
		CardholderName: "Vitalii Karniushin",
		Number:         []byte("2345548223450943"),
		ExpiryMonth:    8,
		ExpiryYear:     27,
		Cvc:            []byte("345"),
	}
	processor := operation.NewProcessorBuilder().
		WithEncryption(encryptionService).
		WithStorageCreator(ctx, pool).
		Build()
	if err = processor.Process(login); err != nil {
		log.Fatalf("Failed to process login: %s", err)
	}
	if err = processor.Process(card); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}

	retrieveLogin := &models.Login{
		SecretMetadata: models.SecretMetadata{
			Path: "login1",
		},
	}
	decrypt := operation.NewProcessorBuilder().
		WithStorageRetriever(ctx, pool).
		WithDecryption(encryptionService).
		Build()
	if err = decrypt.Process(retrieveLogin); err != nil {
		log.Fatalf("Failed to process login: %s", err)
	}
	logger.Log().Infof("[login=%s password=%s]", retrieveLogin.Login, string(retrieveLogin.Password))
	retrieveCard := &models.Card{
		SecretMetadata: models.SecretMetadata{
			Path: "card1",
		},
	}
	if err = decrypt.Process(retrieveCard); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}
	logger.Log().Infof("[number=%s cvc=%s]", string(retrieveCard.Number), string(retrieveCard.Cvc))

	if err = sql.NewStorageDeleter(ctx, pool).VisitLogin(login); err != nil {
		log.Fatalf("Failed to process login: %s", err)
	}
	if err = sql.NewStorageDeleter(ctx, pool).VisitCard(card); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}

	// if err != nil {
	// 	return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	// }
}
