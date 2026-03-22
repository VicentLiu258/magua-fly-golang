package postgres

import (
	"context"
	"database/sql"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type AircraftAvailabilityRepo struct{ db *DB }

func (db *DB) AircraftAvailability() repository.AircraftAvailabilityRepository {
	return &AircraftAvailabilityRepo{db: db}
}

var _ repository.AircraftAvailabilityRepository = (*AircraftAvailabilityRepo)(nil)

func (r *AircraftAvailabilityRepo) ListByDate(ctx context.Context, date time.Time) ([]*domain.AircraftAvailability, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, aircraft_id::text, date, start_time, end_time, status,
		       maintenance_window_id::text, created_at, updated_at
		FROM aircraft_availability
		WHERE date = $1
		ORDER BY aircraft_id, start_time`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.AircraftAvailability
	for rows.Next() {
		var a domain.AircraftAvailability
		var mwID sql.NullString
		if err := rows.Scan(&a.ID, &a.AircraftID, &a.Date, &a.StartTime, &a.EndTime, &a.Status, &mwID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		if mwID.Valid {
			s := mwID.String
			a.MaintenanceWindowID = &s
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *AircraftAvailabilityRepo) Create(ctx context.Context, a *domain.AircraftAvailability) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO aircraft_availability (aircraft_id, date, start_time, end_time, status, maintenance_window_id)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at, updated_at`,
		a.AircraftID, a.Date, a.StartTime, a.EndTime, a.Status, nullUUID(a.MaintenanceWindowID),
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *AircraftAvailabilityRepo) DeleteByMaintenanceWindowID(ctx context.Context, maintenanceWindowID string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM aircraft_availability WHERE maintenance_window_id = $1::uuid`, maintenanceWindowID)
	return err
}


