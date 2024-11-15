package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/itallix/gophkeeper/internal/common/logger"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		pool: pool,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, login, passwordHash string) error {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	errPrefix := "[CREATE USER]"

	insertSQL := `
	INSERT INTO users(
		login,
		password_hash
	) VALUES($1, $2)`

	_, err := r.pool.Exec(c, insertSQL, login, passwordHash)
	if err != nil {
		return fmt.Errorf("%s failed to insert users: %w", errPrefix, err)
	}

	logger.Log().Infof("User with login=[%s] has been successfully created.", login)

	return nil
}

func (r *UserRepo) GetPasswordHash(ctx context.Context, login string) (string, error) {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	var hash string
	selectSQL := "SELECT password_hash FROM users WHERE login = $1"

	if err := r.pool.QueryRow(c, selectSQL, login).Scan(&hash); err != nil {
		return "", fmt.Errorf("failed to get user password hash: %w", err)
	}

	return hash, nil
}

func (r *UserRepo) Exists(ctx context.Context, login string) (bool, error) {
	c, cancel := context.WithTimeout(ctx, TimeoutInSeconds*time.Second)
	defer cancel()

	var count int
	selectSQL := "SELECT COUNT(*) FROM users WHERE login = $1"

	if err := r.pool.QueryRow(c, selectSQL, login).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	exists := count > 0
	logger.Log().Infof("User existence check for login=[%s]: exists=%v", login, exists)

	return exists, nil
}
