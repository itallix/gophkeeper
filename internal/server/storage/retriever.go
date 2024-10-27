package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
)

type Retriever struct {
	pool    *pgxpool.Pool
	context context.Context
}

func NewRetriever(ctx context.Context, pool *pgxpool.Pool) *Retriever {
	return &Retriever{
		context: ctx,
		pool:    pool,
	}
}

func (s *Retriever) VisitLogin(login *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[RETRIEVE LOGIN]"
	selectSQL := `
	SELECT encrypted_data_key, created_at, created_by, login, password FROM logins l 
	INNER JOIN secrets s ON l.secret_id = s.secret_id
	WHERE s.path = $1
	`

	err := s.pool.QueryRow(ctx, selectSQL, login.Path).
		Scan(
			&login.EncryptedDataKey, 
			&login.CreatedAt,
			&login.CreatedBy,
			&login.Login, 
			&login.Password,
		)
	if err != nil {
		return fmt.Errorf("%s failed to query logins: %w", errPrefix, err)
	}

	return nil
}

func (s *Retriever) VisitCard(card *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[RETRIEVE CARD]"
	selectSQL := `
	SELECT encrypted_data_key, created_at, created_by, cardholder_name, number, expiry_month, expiry_year, cvc 
	FROM cards c 
	INNER JOIN secrets s ON c.secret_id = s.secret_id
	WHERE s.path = $1
	`

	err := s.pool.QueryRow(ctx, selectSQL, card.Path).
		Scan(
			&card.EncryptedDataKey,
			&card.CreatedAt,
			&card.CreatedBy,
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

func (s *Retriever) VisitNote(note *models.Note) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[RETRIEVE NOTE]"
	selectSQL := `
	SELECT encrypted_data_key, created_at, created_by, text FROM notes n 
	INNER JOIN secrets s ON n.secret_id = s.secret_id
	WHERE s.path = $1
	`

	err := s.pool.QueryRow(ctx, selectSQL, note.Path).
		Scan(&note.EncryptedDataKey, &note.CreatedAt, &note.CreatedBy, &note.Text)
	if err != nil {
		return fmt.Errorf("%s failed to query notes: %w", errPrefix, err)
	}

	return nil
}

func (s *Retriever) VisitBinary(_ *models.Binary) error {
	return nil
}

func (s *Retriever) GetResult() any {
	return nil
}
