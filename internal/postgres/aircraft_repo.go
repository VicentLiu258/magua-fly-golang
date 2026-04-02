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

const aircraftSelectBase = `
		SELECT id::text, code, COALESCE(name,''), COALESCE(status,''),
		       COALESCE(avatar_gridfs_id,''),
		       total_airframe_hours, last_inspection_hours, COALESCE(inspection_interval_hours, 50),
		       created_at, updated_at
		FROM aircraft `

func scanAircraftRow(
	scanner interface {
		Scan(dest ...any) error
	},
) (*domain.Aircraft, error) {
	var a domain.Aircraft
	var total, last sql.NullFloat64
	var avatar string
	err := scanner.Scan(
		&a.ID, &a.Code, &a.Name, &a.Status, &avatar,
		&total, &last, &a.InspectionIntervalHours,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.AvatarGridFSID = avatar
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

func (r *AircraftRepo) GetByID(ctx context.Context, id string) (*domain.Aircraft, error) {
	return scanAircraftRow(
		r.db.Pool.QueryRow(ctx, aircraftSelectBase+` WHERE id = $1::uuid`, id),
	)
}

func (r *AircraftRepo) GetByCode(ctx context.Context, code string) (*domain.Aircraft, error) {
	return scanAircraftRow(
		r.db.Pool.QueryRow(ctx, aircraftSelectBase+` WHERE code = $1`, code),
	)
}

func (r *AircraftRepo) ListByCodeOrNameNormalized(ctx context.Context, codeOrName string) ([]*domain.Aircraft, error) {
	rows, err := r.db.Pool.Query(ctx, aircraftSelectBase+`
		WHERE lower(trim(code)) = lower(trim($1))
		   OR lower(trim(COALESCE(name,''))) = lower(trim($1))`, codeOrName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Aircraft
	for rows.Next() {
		a, err := scanAircraftRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *AircraftRepo) List(ctx context.Context) ([]*domain.Aircraft, error) {
	rows, err := r.db.Pool.Query(ctx, aircraftSelectBase+` ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Aircraft
	for rows.Next() {
		a, err := scanAircraftRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *AircraftRepo) Create(ctx context.Context, a *domain.Aircraft) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO aircraft (code, name, status) VALUES ($1, $2, $3)
		RETURNING id::text, created_at, updated_at`, a.Code, a.Name, a.Status,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *AircraftRepo) Update(ctx context.Context, id string, name, status *string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE aircraft SET
			name = COALESCE($2, name),
			status = COALESCE($3, status),
			updated_at = NOW()
		WHERE id = $1::uuid`,
		id, name, status,
	)
	return err
}

func (r *AircraftRepo) SetAvatarGridFSID(ctx context.Context, id string, gridFSID string) error {
	var v interface{}
	if gridFSID == "" {
		v = nil
	} else {
		v = gridFSID
	}
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE aircraft SET avatar_gridfs_id = $2, updated_at = NOW() WHERE id = $1::uuid`,
		id, v,
	)
	return err
}

func (r *AircraftRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM aircraft WHERE id = $1::uuid`, id)
	return err
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
