package postgres

import (
	"context"
	"fmt"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type AircraftMaintenanceRepo struct{ db *DB }

func (db *DB) AircraftMaintenance() repository.AircraftMaintenanceRepository {
	return &AircraftMaintenanceRepo{db: db}
}

var _ repository.AircraftMaintenanceRepository = (*AircraftMaintenanceRepo)(nil)

func (r *AircraftMaintenanceRepo) Create(ctx context.Context, w *domain.MaintenanceWindow) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO aircraft_maintenance_windows (
			aircraft_id, kind, scheduled_start_at, scheduled_end_at, status, hangar_location, notes
		) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, created_at, updated_at`,
		w.AircraftID, w.Kind, w.ScheduledStartAt, w.ScheduledEndAt, w.Status, nullStr(w.HangarLocation), nullStr(w.Notes),
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *AircraftMaintenanceRepo) GetByID(ctx context.Context, id string) (*domain.MaintenanceWindow, error) {
	var w domain.MaintenanceWindow
	var hangar, notes *string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, aircraft_id::text, kind, scheduled_start_at, scheduled_end_at, status::text,
		       hangar_location, notes, created_at, updated_at
		FROM aircraft_maintenance_windows WHERE id = $1::uuid`, id,
	).Scan(&w.ID, &w.AircraftID, &w.Kind, &w.ScheduledStartAt, &w.ScheduledEndAt, &w.Status, &hangar, &notes, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if hangar != nil {
		w.HangarLocation = *hangar
	}
	if notes != nil {
		w.Notes = *notes
	}
	return &w, nil
}

func (r *AircraftMaintenanceRepo) Update(ctx context.Context, w *domain.MaintenanceWindow) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE aircraft_maintenance_windows SET
			kind = $2,
			scheduled_start_at = $3,
			scheduled_end_at = $4,
			status = $5::maintenance_window_status,
			hangar_location = $6,
			notes = $7,
			updated_at = NOW()
		WHERE id = $1::uuid`,
		w.ID, w.Kind, w.ScheduledStartAt, w.ScheduledEndAt, w.Status, nullStr(w.HangarLocation), nullStr(w.Notes),
	)
	return err
}

func (r *AircraftMaintenanceRepo) ListByAircraft(ctx context.Context, aircraftID string, from, to *time.Time, status *domain.MaintenanceWindowStatus) ([]*domain.MaintenanceWindow, error) {
	q := `
		SELECT id::text, aircraft_id::text, kind, scheduled_start_at, scheduled_end_at, status::text,
		       hangar_location, notes, created_at, updated_at
		FROM aircraft_maintenance_windows
		WHERE 1=1`
	var args []interface{}
	idx := 1
	if aircraftID != "" {
		q += fmt.Sprintf(` AND aircraft_id = $%d::uuid`, idx)
		args = append(args, aircraftID)
		idx++
	}
	if from != nil {
		q += fmt.Sprintf(` AND scheduled_end_at >= $%d`, idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		q += fmt.Sprintf(` AND scheduled_start_at <= $%d`, idx)
		args = append(args, *to)
		idx++
	}
	if status != nil {
		q += fmt.Sprintf(` AND status = $%d::maintenance_window_status`, idx)
		args = append(args, string(*status))
	}
	q += ` ORDER BY scheduled_start_at`
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.MaintenanceWindow
	for rows.Next() {
		var w domain.MaintenanceWindow
		var hangar, notes *string
		if err := rows.Scan(&w.ID, &w.AircraftID, &w.Kind, &w.ScheduledStartAt, &w.ScheduledEndAt, &w.Status, &hangar, &notes, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		if hangar != nil {
			w.HangarLocation = *hangar
		}
		if notes != nil {
			w.Notes = *notes
		}
		list = append(list, &w)
	}
	return list, rows.Err()
}
