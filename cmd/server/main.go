package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/logger"
	"gophkeeper.com/internal/server"
	"gophkeeper.com/internal/server/models"
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
		ExpiryYear:     2027,
		CVC:            []byte("345"),
	}

	vault := server.NewVault(ctx, pool, encryptionService)

	if err = vault.StoreSecret("vitalii", "jwt", login); err != nil {
		log.Fatalf("Failed to process login: %s", err)
	}
	if err = vault.StoreSecret("vitalii", "jwt", card); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}

	retrieveLogin := &models.Login{
		SecretMetadata: models.SecretMetadata{
			Path: "login1",
		},
	}
	if err = vault.RetrieveSecret("vitalii", "jwt", retrieveLogin); err != nil {
		log.Fatalf("Failed to process login: %s", err)
	}
	logger.Log().Infof("[login=%s password=%s]", retrieveLogin.Login, string(retrieveLogin.Password))

	retrieveCard := models.NewCard(
		[]models.SecretOption{
			models.WithPath("card1"),
		},
		nil,
	)
	if err = vault.RetrieveSecret("vitalii", "jwt", retrieveCard); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}
	logger.Log().Infof("[number=%s cvc=%s]", string(retrieveCard.Number), string(retrieveCard.CVC))

	names, _ := vault.ListSecrets("vitalii", "jwt", retrieveCard)
	logger.Log().Infof("cards=%v", names)
}
