package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"magua-fly-golang/internal/repository"
)

type personnelRepo struct {
	db *DB
}

func (db *DB) Personnel() repository.PersonnelRepository { return &personnelRepo{db: db} }

var _ repository.PersonnelRepository = (*personnelRepo)(nil)

func (r *personnelRepo) SoftDeleteStudent(ctx context.Context, studentID string) error {
	return withTx(ctx, r.db.Pool, func(tx pgx.Tx) error {
		var userID sql.NullString
		err := tx.QueryRow(ctx, `
			SELECT user_id::text FROM students
			WHERE id = $1::uuid AND deleted_at IS NULL`, studentID).Scan(&userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return repository.ErrNotFound
			}
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE students SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid AND deleted_at IS NULL`, studentID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return repository.ErrNotFound
		}
		if userID.Valid && userID.String != "" {
			if _, err := tx.Exec(ctx, `
				UPDATE users SET deleted_at = NOW(), updated_at = NOW()
				WHERE id = $1::uuid AND deleted_at IS NULL`, userID.String); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *personnelRepo) SoftDeleteInstructor(ctx context.Context, instructorID string) error {
	return withTx(ctx, r.db.Pool, func(tx pgx.Tx) error {
		var userID string
		err := tx.QueryRow(ctx, `
			SELECT user_id::text FROM instructors
			WHERE id = $1::uuid AND deleted_at IS NULL`, instructorID).Scan(&userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return repository.ErrNotFound
			}
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE instructors SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid AND deleted_at IS NULL`, instructorID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return repository.ErrNotFound
		}
		if _, err := tx.Exec(ctx, `
			UPDATE users SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid AND deleted_at IS NULL`, userID); err != nil {
			return err
		}
		return nil
	})
}

func (r *personnelRepo) SoftDeleteStaffUser(ctx context.Context, userID string) error {
	return withTx(ctx, r.db.Pool, func(tx pgx.Tx) error {
		var roleStr string
		err := tx.QueryRow(ctx, `
			SELECT role::text FROM users
			WHERE id = $1::uuid AND deleted_at IS NULL`, userID).Scan(&roleStr)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return repository.ErrNotFound
			}
			return err
		}
		if roleStr != "staff" && roleStr != "maintenance" {
			return repository.ErrPersonnelStaffRole
		}
		tag, err := tx.Exec(ctx, `
			UPDATE users SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1::uuid AND deleted_at IS NULL`, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return repository.ErrNotFound
		}
		return nil
	})
}

func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
