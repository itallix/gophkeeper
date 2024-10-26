package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
)

type StorageRetriever struct {
	pool    *pgxpool.Pool
	context context.Context
}

func NewStorageRetriever(ctx context.Context, pool *pgxpool.Pool) *StorageRetriever {
	return &StorageRetriever{
		context: ctx,
		pool:    pool,
	}
}

func (s *StorageRetriever) VisitLogin(login *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[retrieve login]"
	loginSQL := `
	SELECT encrypted_data_key, login, password FROM logins l 
	INNER JOIN secrets s ON l.secret_id = s.secret_id
	WHERE s.path = $1
	`

	err := s.pool.QueryRow(ctx, loginSQL, login.Path).
		Scan(&login.EncryptedDataKey, &login.Login, &login.Password)
	if err != nil {
		return fmt.Errorf("%s failed to query logins: %w", errPrefix, err)
	}

	return nil
}

func (s *StorageRetriever) VisitCard(card *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[retrieve card]"
	cardSQL := `
	SELECT encrypted_data_key, cardholder_name, number, expiry_month, expiry_year, cvc FROM cards c 
	INNER JOIN secrets s ON c.secret_id = s.secret_id
	WHERE s.path = $1
	`

	err := s.pool.QueryRow(ctx, cardSQL, card.Path).
		Scan(
			&card.EncryptedDataKey,
			&card.CardholderName,
			&card.Number,
			&card.ExpiryMonth,
			&card.ExpiryYear,
			&card.CVC,
		)
	if err != nil {
		return fmt.Errorf("%s failed to query logins: %w", errPrefix, err)
	}

	return nil
}

func (s *StorageRetriever) VisitNote(_ *models.Note) error {
	return nil
}

func (s *StorageRetriever) VisitBinary(_ *models.Binary) error {
	return nil
}

func (s *StorageRetriever) GetResult() any {
	return nil
}
