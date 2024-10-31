package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/s3"
	"gophkeeper.com/pkg/logger"
)

type Deleter struct {
	pool          *pgxpool.Pool
	context       context.Context
	objectStorage *s3.ObjectStorage
}

func NewDeleter(ctx context.Context, pool *pgxpool.Pool, objectStorage *s3.ObjectStorage) *Deleter {
	return &Deleter{
		context:       ctx,
		pool:          pool,
		objectStorage: objectStorage,
	}
}

func deleteSecret(ctx context.Context, pool *pgxpool.Pool, secretPath string) error {
	deleteSQL := "DELETE FROM secrets WHERE path = $1 RETURNING secret_id"

	if _, err := pool.Exec(ctx, deleteSQL, secretPath); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

func (s *Deleter) VisitLogin(login *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	if err := deleteSecret(ctx, s.pool, login.Path); err != nil {
		return fmt.Errorf("[DELETE LOGIN]: %w", err)
	}

	logger.Log().Infof("Login with path=[%s] has been successfully deleted.", login.Path)

	return nil
}

func (s *Deleter) VisitCard(card *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	if err := deleteSecret(ctx, s.pool, card.Path); err != nil {
		return fmt.Errorf("[DELETE CARD]: %w", err)
	}

	logger.Log().Infof("Card with path=[%s] has been successfully deleted.", card.Path)

	return nil
}

func (s *Deleter) VisitNote(note *models.Note) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	if err := deleteSecret(ctx, s.pool, note.Path); err != nil {
		return fmt.Errorf("[DELETE NOTE]: %w", err)
	}

	logger.Log().Infof("Note with path=[%s] has been successfully deleted.", note.Path)

	return nil
}

func (s *Deleter) VisitBinary(binary *models.Binary) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	if err := s.objectStorage.DeleteChunks(ctx, BucketBinaries, binary.Path); err != nil {
		return err
	}

	if err := deleteSecret(ctx, s.pool, binary.Path); err != nil {
		return fmt.Errorf("[DELETE BINARY]: %w", err)
	}

	logger.Log().Infof("Binary [%s] has been successfully deleted.", binary.Path)

	return nil
}

func (s *Deleter) GetResult() any {
	return nil
}
