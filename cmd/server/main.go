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
	login := models.NewLogin(
		[]models.SecretOption{
			models.WithPath("login1"),
			models.WithCustomMetadata(map[string]string{"attr1": "foo", "attr2": "boo"}),
			models.WithCreatedBy("vitalii"),
			models.WithModifiedBy("vitalii"),
		},
		[]models.LoginOption{
			models.WithLogin("vitalii"),
			models.WithPassword("geheim"),
		},
	)
	card := models.NewCard(
		[]models.SecretOption{
			models.WithPath("card1"),
			models.WithCustomMetadata(map[string]string{"attr1": "foo", "attr2": "boo"}),
			models.WithCreatedBy("vitalii"),
			models.WithModifiedBy("vitalii"),
		},
		[]models.CardOption{
			models.WithCardHolder("Vitalii Karniushin"),
			models.WithCardNumber("2345548223450943"),
			models.WithExpiry(8, 2027),
			models.WithCVC("345"),
		},
	)
	note := models.NewNote(
		[]models.SecretOption{
			models.WithPath("note1"),
			models.WithCustomMetadata(map[string]string{"attr1": "foo", "attr2": "boo"}),
			models.WithCreatedBy("vitalii"),
			models.WithModifiedBy("vitalii"),
		},
		[]models.NoteOption{
			models.WithText(`
				Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt 
				ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco 
				laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in 
				voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat 
				non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`),
		},
	)

	vault := server.NewVault(ctx, pool, encryptionService)

	if err = vault.StoreSecret("vitalii", "jwt", login); err != nil {
		log.Fatalf("Failed to process login: %s", err)
	}
	if err = vault.StoreSecret("vitalii", "jwt", card); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}
	if err = vault.StoreSecret("vitalii", "jwt", note); err != nil {
		log.Fatalf("Failed to process note: %s", err)
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

	retrieveNote := models.NewNote(
		[]models.SecretOption{
			models.WithPath("note1"),
		},
		nil,
	)
	if err = vault.RetrieveSecret("vitalii", "jwt", retrieveNote); err != nil {
		log.Fatalf("Failed to process card: %s", err)
	}
	logger.Log().Infof("[text=%s]", string(retrieveNote.Text))

	names, _ := vault.ListSecrets("vitalii", "jwt", retrieveCard)
	logger.Log().Infof("cards=%v", names)
}
