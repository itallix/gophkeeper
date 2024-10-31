package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/common/logger"
	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/s3"
)

type Retriever struct {
	pool    *pgxpool.Pool
	context context.Context
	objectStorage *s3.ObjectStorage
}

func NewRetriever(ctx context.Context, pool *pgxpool.Pool, objectStorage *s3.ObjectStorage) *Retriever {
	return &Retriever{
		context: ctx,
		pool:    pool,
		objectStorage: objectStorage,
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

func (s *Retriever) VisitBinary(binary *models.Binary) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	if binary.Chunks == 0 {
		errPrefix := "[RETRIEVE BINARY]"
		selectSQL := `
		SELECT encrypted_data_key, created_at, created_by, chunks, hash FROM binaries b
		INNER JOIN secrets s ON b.secret_id = s.secret_id
		WHERE s.path = $1
		`

		err := s.pool.QueryRow(ctx, selectSQL, binary.Path).
			Scan(
				&binary.EncryptedDataKey, 
				&binary.CreatedAt, 
				&binary.CreatedBy, 
				&binary.Chunks,
				&binary.Hash,
			)
		if err != nil {
			return fmt.Errorf("%s failed to query binaries: %w", errPrefix, err)
		}
	} else {
		chunkName := fmt.Sprintf("%s/%d", binary.Path, binary.ChunkID)
		reader, size, err := s.objectStorage.GetObject(ctx, BucketBinaries, chunkName)
		if err != nil {
			return fmt.Errorf("error getting chunk data from storage: %w", err)
		}
		defer func(){
			_ = reader.Close()
		}()
		data, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("error reading chunk data to buffer: %w", err)
		}
		binary.Data = data
		logger.Log().Infof("Binary chunk with size=%d & name=%s has been successfully loaded.", size, chunkName)
	}
	return nil
}

func (s *Retriever) GetResult() any {
	return nil
}
