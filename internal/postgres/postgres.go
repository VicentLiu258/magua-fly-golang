package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps pgx pool for dependency injection.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new PostgreSQL connection pool.
func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &DB{Pool: pool}, nil
}

// Close closes the connection pool.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Health checks if the database is reachable.
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
