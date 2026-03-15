package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type AircraftRepo struct{ db *DB }

func (db *DB) Aircraft() repository.AircraftRepository { return &AircraftRepo{db: db} }

var _ repository.AircraftRepository = (*AircraftRepo)(nil)

func (r *AircraftRepo) GetByID(ctx context.Context, id string) (*domain.Aircraft, error) {
	var a domain.Aircraft
	err := r.db.Pool.QueryRow(ctx, `SELECT id::text, code, COALESCE(name,''), COALESCE(status,''), created_at, updated_at FROM aircraft WHERE id = $1::uuid`, id).Scan(
		&a.ID, &a.Code, &a.Name, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AircraftRepo) List(ctx context.Context) ([]*domain.Aircraft, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id::text, code, COALESCE(name,''), COALESCE(status,''), created_at, updated_at FROM aircraft ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Aircraft
	for rows.Next() {
		var a domain.Aircraft
		if err := rows.Scan(&a.ID, &a.Code, &a.Name, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *AircraftRepo) Create(ctx context.Context, a *domain.Aircraft) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO aircraft (code, name, status) VALUES ($1, $2, $3)
		RETURNING id::text, created_at, updated_at`, a.Code, a.Name, a.Status,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}
