package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type InstructorRepo struct{ db *DB }

func (db *DB) Instructor() repository.InstructorRepository { return &InstructorRepo{db: db} }

var _ repository.InstructorRepository = (*InstructorRepo)(nil)

func (r *InstructorRepo) GetByID(ctx context.Context, id int64) (*domain.Instructor, error) {
	var i domain.Instructor
	err := r.db.Pool.QueryRow(ctx, `SELECT id, user_id, name, COALESCE(display_color,''), created_at, updated_at FROM instructors WHERE id = $1`, id).Scan(
		&i.ID, &i.UserID, &i.Name, &i.DisplayColor, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *InstructorRepo) List(ctx context.Context) ([]*domain.Instructor, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id, user_id, name, COALESCE(display_color,''), created_at, updated_at FROM instructors ORDER BY id`)
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
		INSERT INTO instructors (user_id, name, display_color) VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`, i.UserID, i.Name, i.DisplayColor,
	).Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
}
