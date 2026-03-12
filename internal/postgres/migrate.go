package postgres

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// RunMigrations runs SQL migrations from the given path (e.g. "file://migrations").
func RunMigrations(dsn, migratePath string) error {
	// pgx stdlib registers driver "pgx" so we can use the same DSN as pgxpool.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres.WithInstance: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migratePath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	return nil
}
