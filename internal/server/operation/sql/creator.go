package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/logger"
	"gophkeeper.com/internal/server/models"
)

type StorageCreator struct {
	pool    *pgxpool.Pool
	context context.Context
}

func NewStorageCreator(ctx context.Context, pool *pgxpool.Pool) *StorageCreator {
	return &StorageCreator{
		context: ctx,
		pool:    pool,
	}
}

func createSecret(ctx context.Context, tx pgx.Tx, secret models.SecretMetadata) (int64, error) {
	insertSQL := `
	INSERT INTO secrets (
		path,
		created_at,
		modified_at,
		custom_metadata,
		encrypted_data_key,
		created_by,
		modified_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING secret_id`

	var secretID int64
	if err := tx.QueryRow(ctx, insertSQL,
		secret.Path,
		secret.CreatedAt,
		secret.ModifiedAt,
		secret.CustomMeta,
		secret.EncryptedDataKey,
		secret.CreatedBy,
		secret.ModifiedBy,
	).Scan(&secretID); err != nil {
		return 0, fmt.Errorf("failed to insert secret: %w", err)
	}

	return secretID, nil
}

func (s *StorageCreator) VisitLogin(login *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[create login]"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to begin transaction: %w", errPrefix, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	secretID, err := createSecret(ctx, tx, login.SecretMetadata)
	if err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	insertSQL := `
        INSERT INTO logins (
            secret_id,
            login,
            password
        ) VALUES ($1, $2, $3)
		RETURNING login_id`

	var loginID int64
	if err = tx.QueryRow(ctx, insertSQL,
		secretID,
		login.Login,
		login.Password,
	).Scan(&loginID); err != nil {
		return fmt.Errorf("%s failed to insert login: %w", errPrefix, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit transaction: %w", errPrefix, err)
	}

	login.SecretID = secretID
	login.LoginID = loginID

	logger.Log().Infof("Login with path=[%s] has been successfully created.", login.Path)

	return nil
}

func (s *StorageCreator) VisitCard(card *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[create card]"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to begin transaction: %w", errPrefix, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	secretID, err := createSecret(ctx, tx, card.SecretMetadata)
	if err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	insertSQL := `
        INSERT INTO cards (
            secret_id,
            cardholder_name,
            number,
			expiry_month,
			expiry_year,
			cvc
        ) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING card_id`

	var cardID int64
	if err = tx.QueryRow(ctx, insertSQL,
		secretID,
		card.CardholderName,
		card.Number,
		card.ExpiryMonth,
		card.ExpiryYear,
		card.CVC,
	).Scan(&cardID); err != nil {
		return fmt.Errorf("%s failed to insert card: %w", errPrefix, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit transaction: %w", errPrefix, err)
	}

	// Update the login ID after successful insert
	card.SecretID = secretID
	card.CardID = cardID

	logger.Log().Infof("Card with path=[%s] has been successfully created.", card.Path)

	return nil
}

func (s *StorageCreator) VisitNote(_ *models.Note) error {
	return nil
}

func (s *StorageCreator) VisitBinary(_ *models.Binary) error {
	return nil
}

func (s *StorageCreator) GetResult() any {
	return nil
}
