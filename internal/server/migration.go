package server

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	mp "github.com/golang-migrate/migrate/v4/database/postgres"
)

// ApplyMigrations applies database migrations from the specified file to a PostgreSQL database.
// It uses the golang-migrate library to handle the migration process.
//
// The function will attempt to run all pending migrations. If there are no pending migrations,
// it will return nil. Any other errors during the migration process will be returned.
//
// Parameters:
//   - dsn: Database connection string in PostgreSQL format
//     (e.g., "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
//   - filename: Path to the directory containing migration files (without the "file://" prefix)
//
// Returns:
//   - error: nil if migrations are successfully applied or if there are no changes to apply,
//     otherwise returns an error describing what went wrong during the migration process
//
// Example:
//
//	err := ApplyMigrations("postgres://user:pass@localhost:5432/mydb?sslmode=disable", "./migrations")
//	if err != nil {
//	    log.Fatal("failed to apply migrations:", err)
//	}
func ApplyMigrations(dsn, filename string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := mp.WithInstance(db, &mp.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+filename, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
