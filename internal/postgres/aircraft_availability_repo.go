package postgres

import (
	"context"
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
		SELECT id, aircraft_id, date, start_time, end_time, status, created_at, updated_at
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
		if err := rows.Scan(&a.ID, &a.AircraftID, &a.Date, &a.StartTime, &a.EndTime, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

