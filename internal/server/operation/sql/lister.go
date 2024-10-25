package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophkeeper.com/internal/server/models"
)

type StorageLister struct {
	pool    *pgxpool.Pool
	context context.Context
	result  []string
}

func NewStorageLister(ctx context.Context, pool *pgxpool.Pool) *StorageLister {
	return &StorageLister{
		context: ctx,
		pool:    pool,
		result:  nil,
	}
}

func (s *StorageLister) VisitLogin(_ *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[list logins]"
	loginSQL := "SELECT path FROM logins l INNER JOIN secrets s ON l.secret_id = s.secret_id"
	rows, err := s.pool.Query(ctx, loginSQL)
	if err != nil {
		return fmt.Errorf("%s failed to query logins: %w", errPrefix, err)
	}
	defer rows.Close()

	var secrets []string
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return fmt.Errorf("%s failed to scan secret: %w", errPrefix, err)
		}
		secrets = append(secrets, s)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("%s error during iteration: %w", errPrefix, err)
	}

	s.result = secrets
	return nil
}

func (s *StorageLister) VisitCard(_ *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[list cards]"
	loginSQL := "SELECT path FROM cards l INNER JOIN secrets s ON l.secret_id = s.secret_id"
	rows, err := s.pool.Query(ctx, loginSQL)
	if err != nil {
		return fmt.Errorf("%s failed to query cards: %w", errPrefix, err)
	}
	defer rows.Close()

	var secrets []string
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return fmt.Errorf("%s failed to scan secret: %w", errPrefix, err)
		}
		secrets = append(secrets, s)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("%s error during iteration: %w", errPrefix, err)
	}

	s.result = secrets
	return nil
}

func (s *StorageLister) VisitNote(_ *models.Note) error {
	return nil
}

func (s *StorageLister) VisitBinary(_ *models.Binary) error {
	return nil
}

func (s *StorageLister) GetResult() any {
	return s.result
}
