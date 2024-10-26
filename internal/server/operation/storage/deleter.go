package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/logger"
	"gophkeeper.com/internal/server/models"
)

type Deleter struct {
	pool    *pgxpool.Pool
	context context.Context
}

func NewDeleter(ctx context.Context, pool *pgxpool.Pool) *Deleter {
	return &Deleter{
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

func (s *Deleter) VisitLogin(login *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[DELETE LOGIN]"
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

func (s *Deleter) VisitCard(card *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[DELETE CARD]"
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

func (s *Deleter) VisitNote(note *models.Note) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[DELETE NOTE]"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to begin transaction: %w", errPrefix, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err = deleteSecret(ctx, tx, note.Path); err != nil {
		return fmt.Errorf("%s: %w", errPrefix, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s failed to commit transaction: %w", errPrefix, err)
	}

	logger.Log().Infof("Note with path=[%s] has been successfully deleted.", note.Path)

	return nil
}

func (s *Deleter) VisitBinary(_ *models.Binary) error {
	return nil
}

func (s *Deleter) GetResult() any {
	return nil
}
