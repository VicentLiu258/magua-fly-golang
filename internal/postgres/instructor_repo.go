package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type InstructorRepo struct{ db *DB }

func (db *DB) Instructor() repository.InstructorRepository { return &InstructorRepo{db: db} }

var _ repository.InstructorRepository = (*InstructorRepo)(nil)

func (r *InstructorRepo) GetByID(ctx context.Context, id string) (*domain.Instructor, error) {
	var i domain.Instructor
	err := r.db.Pool.QueryRow(ctx, `SELECT id::text, user_id::text, name, COALESCE(display_color,''), created_at, updated_at FROM instructors WHERE id = $1::uuid AND deleted_at IS NULL`, id).Scan(
		&i.ID, &i.UserID, &i.Name, &i.DisplayColor, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *InstructorRepo) GetByUserID(ctx context.Context, userID string) (*domain.Instructor, error) {
	var i domain.Instructor
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, name, COALESCE(display_color,''), created_at, updated_at
		FROM instructors WHERE user_id = $1::uuid AND deleted_at IS NULL`, userID).Scan(
		&i.ID, &i.UserID, &i.Name, &i.DisplayColor, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

func (r *InstructorRepo) List(ctx context.Context) ([]*domain.Instructor, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id::text, user_id::text, name, COALESCE(display_color,''), created_at, updated_at FROM instructors WHERE deleted_at IS NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Instructor
	for rows.Next() {
		var i domain.Instructor
		if err := rows.Scan(&i.ID, &i.UserID, &i.Name, &i.DisplayColor, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &i)
	}
	return list, rows.Err()
}

func (r *InstructorRepo) Create(ctx context.Context, i *domain.Instructor) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO instructors (user_id, name, display_color) VALUES ($1::uuid, $2, $3)
		RETURNING id::text, created_at, updated_at`, i.UserID, i.Name, i.DisplayColor,
	).Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
}

func (r *InstructorRepo) Update(ctx context.Context, i *domain.Instructor) error {
	tag, err := r.db.Pool.Exec(ctx, `
		UPDATE instructors
		SET name = $2,
		    display_color = NULLIF(TRIM($3), ''),
		    updated_at = NOW()
		WHERE id = $1::uuid AND deleted_at IS NULL`,
		i.ID, i.Name, i.DisplayColor,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
