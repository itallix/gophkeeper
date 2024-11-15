package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/itallix/gophkeeper/internal/server/models"
)

type Lister struct {
	pool    *pgxpool.Pool
	context context.Context
	result  []string
}

func NewLister(ctx context.Context, pool *pgxpool.Pool) *Lister {
	return &Lister{
		context: ctx,
		pool:    pool,
		result:  nil,
	}
}

func listSecrets(ctx context.Context, pool *pgxpool.Pool, query, errMsgPrexix string) ([]string, error) {
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s failed to query logins: %w", errMsgPrexix, err)
	}
	defer rows.Close()

	var secrets []string
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, fmt.Errorf("%s failed to scan secret: %w", errMsgPrexix, err)
		}
		secrets = append(secrets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s error during iteration: %w", errMsgPrexix, err)
	}

	return secrets, nil
}

func (s *Lister) VisitLogin(_ *models.Login) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	selectSQL := "SELECT path FROM logins l INNER JOIN secrets s ON l.secret_id = s.secret_id"
	secrets, err := listSecrets(ctx, s.pool, selectSQL, "[LIST LOGINS]")
	if err != nil {
		return err
	}

	s.result = secrets
	return nil
}

func (s *Lister) VisitCard(_ *models.Card) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	selectSQL := "SELECT path FROM cards l INNER JOIN secrets s ON l.secret_id = s.secret_id"
	secrets, err := listSecrets(ctx, s.pool, selectSQL, "[LIST CARDS]")
	if err != nil {
		return err
	}

	s.result = secrets
	return nil
}

func (s *Lister) VisitNote(_ *models.Note) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	selectSQL := "SELECT path FROM notes n INNER JOIN secrets s ON n.secret_id = s.secret_id"
	secrets, err := listSecrets(ctx, s.pool, selectSQL, "[LIST NOTES]")
	if err != nil {
		return err
	}

	s.result = secrets
	return nil
}

func (s *Lister) VisitBinary(_ *models.Binary) error {
	ctx, cancel := context.WithTimeout(s.context, TimeoutInSeconds*time.Second)
	defer cancel()

	selectSQL := "SELECT path FROM binaries b INNER JOIN secrets s ON b.secret_id = s.secret_id"
	secrets, err := listSecrets(ctx, s.pool, selectSQL, "[LIST BINARIES]")
	if err != nil {
		return err
	}

	s.result = secrets
	return nil
}

func (s *Lister) GetResult() any {
	return s.result
}
