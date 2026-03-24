package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SyllabusOutlineRepo struct{ db *DB }

func (db *DB) SyllabusOutline() repository.SyllabusOutlineRepository {
	return &SyllabusOutlineRepo{db: db}
}

var _ repository.SyllabusOutlineRepository = (*SyllabusOutlineRepo)(nil)

func (r *SyllabusOutlineRepo) GetByID(ctx context.Context, id string) (*domain.SyllabusOutline, error) {
	var o domain.SyllabusOutline
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, category, name, sort_order, created_at, updated_at
		FROM syllabus_outlines WHERE id = $1::uuid`, id,
	).Scan(&o.ID, &o.Category, &o.Name, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *SyllabusOutlineRepo) List(ctx context.Context) ([]*domain.SyllabusOutline, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, category, name, sort_order, created_at, updated_at
		FROM syllabus_outlines
		ORDER BY sort_order, category, name, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.SyllabusOutline
	for rows.Next() {
		var o domain.SyllabusOutline
		if err := rows.Scan(&o.ID, &o.Category, &o.Name, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &o)
	}
	return list, rows.Err()
}

func (r *SyllabusOutlineRepo) Create(ctx context.Context, o *domain.SyllabusOutline) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO syllabus_outlines (category, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id::text, created_at, updated_at`,
		o.Category, o.Name, o.SortOrder,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}
