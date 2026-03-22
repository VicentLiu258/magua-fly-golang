package postgres

import (
	"context"
	"database/sql"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type AircraftRepo struct{ db *DB }

func (db *DB) Aircraft() repository.AircraftRepository { return &AircraftRepo{db: db} }

var _ repository.AircraftRepository = (*AircraftRepo)(nil)

func (r *AircraftRepo) GetByID(ctx context.Context, id string) (*domain.Aircraft, error) {
	var a domain.Aircraft
	var total, last sql.NullFloat64
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, code, COALESCE(name,''), COALESCE(status,''),
		       total_airframe_hours, last_inspection_hours, COALESCE(inspection_interval_hours, 50),
		       created_at, updated_at
		FROM aircraft WHERE id = $1::uuid`, id).Scan(
		&a.ID, &a.Code, &a.Name, &a.Status,
		&total, &last, &a.InspectionIntervalHours,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if total.Valid {
		v := total.Float64
		a.TotalAirframeHours = &v
	}
	if last.Valid {
		v := last.Float64
		a.LastInspectionHours = &v
	}
	return &a, nil
}

func (r *AircraftRepo) List(ctx context.Context) ([]*domain.Aircraft, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, code, COALESCE(name,''), COALESCE(status,''),
		       total_airframe_hours, last_inspection_hours, COALESCE(inspection_interval_hours, 50),
		       created_at, updated_at
		FROM aircraft ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Aircraft
	for rows.Next() {
		var a domain.Aircraft
		var total, last sql.NullFloat64
		if err := rows.Scan(&a.ID, &a.Code, &a.Name, &a.Status, &total, &last, &a.InspectionIntervalHours, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		if total.Valid {
			v := total.Float64
			a.TotalAirframeHours = &v
		}
		if last.Valid {
			v := last.Float64
			a.LastInspectionHours = &v
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

func (r *AircraftRepo) UpdateAirworthiness(ctx context.Context, id string, totalHours, lastInspectionHours *float64, inspectionIntervalHours *int) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE aircraft SET
			total_airframe_hours = COALESCE($2, total_airframe_hours),
			last_inspection_hours = COALESCE($3, last_inspection_hours),
			inspection_interval_hours = COALESCE($4, inspection_interval_hours),
			updated_at = NOW()
		WHERE id = $1::uuid`,
		id, totalHours, lastInspectionHours, inspectionIntervalHours,
	)
	return err
}
