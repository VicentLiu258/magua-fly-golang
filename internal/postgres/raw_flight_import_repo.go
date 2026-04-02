package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type RawFlightImportRepo struct{ db *DB }

func (db *DB) RawFlightImport() repository.RawFlightImportRepository {
	return &RawFlightImportRepo{db: db}
}

var _ repository.RawFlightImportRepository = (*RawFlightImportRepo)(nil)

func (r *RawFlightImportRepo) Create(ctx context.Context, row *domain.RawFlightImport) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO raw_flight_imports (batch_id, engine_on_time, engine_off_time, flight_duration_minutes, is_local, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at`,
		row.BatchID, row.EngineOnTime, row.EngineOffTime, row.FlightDurationMinutes, row.IsLocal, row.RawData,
	).Scan(&row.ID, &row.CreatedAt)
}

func (r *RawFlightImportRepo) ListPending(ctx context.Context, limit, offset int) ([]*domain.RawFlightImport, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, batch_id, engine_on_time, engine_off_time, flight_duration_minutes, is_local, raw_data, training_record_id::text, created_at
		FROM raw_flight_imports WHERE training_record_id IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.RawFlightImport
	for rows.Next() {
		var row domain.RawFlightImport
		if err := rows.Scan(&row.ID, &row.BatchID, &row.EngineOnTime, &row.EngineOffTime, &row.FlightDurationMinutes, &row.IsLocal, &row.RawData, &row.TrainingRecordID, &row.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &row)
	}
	return list, rows.Err()
}

func (r *RawFlightImportRepo) GetByID(ctx context.Context, id string) (*domain.RawFlightImport, error) {
	var row domain.RawFlightImport
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, batch_id, engine_on_time, engine_off_time, flight_duration_minutes, is_local, raw_data, training_record_id::text, created_at
		FROM raw_flight_imports WHERE id = $1::uuid`, id).Scan(
		&row.ID, &row.BatchID, &row.EngineOnTime, &row.EngineOffTime, &row.FlightDurationMinutes, &row.IsLocal, &row.RawData, &row.TrainingRecordID, &row.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RawFlightImportRepo) SetTrainingRecordID(ctx context.Context, id string, recordID string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE raw_flight_imports SET training_record_id = $2::uuid WHERE id = $1::uuid`, id, recordID)
	return err
}

func (r *RawFlightImportRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM raw_flight_imports WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
