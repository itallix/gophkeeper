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

type StorageDeleter struct {
	pool    *pgxpool.Pool
	context context.Context
}

func NewStorageDeleter(ctx context.Context, pool *pgxpool.Pool) *StorageDeleter {
	return &StorageDeleter{
		context: ctx,
		pool:    pool,
	}
}

func deleteSecret(ctx context.Context, tx pgx.Tx, secretPath string) error {
	deleteSQL := "DELETE FROM secrets WHERE path = $1 RETURNING secret_id"

	if _, err := tx.Exec(ctx, deleteSQL, secretPath); err != nil {
		return fmt.Errorf("failed to insert secret: %w", err)
	}

	return nil
}

func (s *StorageDeleter) VisitLogin(login *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[delete login]"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to begin transaction: %w", errPrefix, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err = deleteSecret(ctx, tx, login.Path); err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit transaction: %w", errPrefix, err)
	}

	logger.Log().Infof("Login with path=[%s] has been successfully deleted.", login.Path)

	return nil
}

func (s *StorageDeleter) VisitCard(card *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[delete card]"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to begin transaction: %w", errPrefix, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err = deleteSecret(ctx, tx, card.Path); err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit transaction: %w", errPrefix, err)
	}

	logger.Log().Infof("Card with path=[%s] has been successfully deleted.", card.Path)

	return nil
}

func (s *StorageDeleter) VisitNote(_ *models.Note) error {
	return nil
}

func (s *StorageDeleter) VisitBinary(_ *models.Binary) error {
	return nil
}

func (s *StorageDeleter) GetResult() any {
	return nil
}
