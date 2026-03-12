package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingRecordRepo struct{ db *DB }

func (db *DB) TrainingRecord() repository.TrainingRecordRepository { return &TrainingRecordRepo{db: db} }

var _ repository.TrainingRecordRepository = (*TrainingRecordRepo)(nil)

func (r *TrainingRecordRepo) GetByID(ctx context.Context, id int64) (*domain.TrainingRecord, error) {
	var rec domain.TrainingRecord
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, student_id, instructor_id, aircraft_id, training_date, engine_on_time, engine_off_time,
		       flight_duration_minutes, is_local, import_source, COALESCE(raw_import_id,''), created_at, updated_at
		FROM training_records WHERE id = $1`, id).Scan(
		&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
		&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *TrainingRecordRepo) ListByStudent(ctx context.Context, studentID int64, limit, offset int) ([]*domain.TrainingRecord, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, student_id, instructor_id, aircraft_id, training_date, engine_on_time, engine_off_time,
		       flight_duration_minutes, is_local, import_source, COALESCE(raw_import_id,''), created_at, updated_at
		FROM training_records WHERE student_id = $1 ORDER BY training_date DESC LIMIT $2 OFFSET $3`, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingRecord
	for rows.Next() {
		var rec domain.TrainingRecord
		if err := rows.Scan(&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
			&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &rec)
	}
	return list, rows.Err()
}

func (r *TrainingRecordRepo) Create(ctx context.Context, rec *domain.TrainingRecord) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_records (student_id, instructor_id, aircraft_id, training_date, engine_on_time, engine_off_time,
			flight_duration_minutes, is_local, import_source, raw_import_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`,
		rec.StudentID, rec.InstructorID, rec.AircraftID, rec.TrainingDate, rec.EngineOnTime, rec.EngineOffTime,
		rec.FlightDurationMinutes, rec.IsLocal, rec.ImportSource, rec.RawImportID,
	).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt)
}

func (r *TrainingRecordRepo) Update(ctx context.Context, rec *domain.TrainingRecord) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE training_records SET student_id=$2, instructor_id=$3, aircraft_id=$4, training_date=$5,
			engine_on_time=$6, engine_off_time=$7, flight_duration_minutes=$8, is_local=$9, import_source=$10, raw_import_id=$11, updated_at=NOW()
		WHERE id=$1`,
		rec.ID, rec.StudentID, rec.InstructorID, rec.AircraftID, rec.TrainingDate, rec.EngineOnTime, rec.EngineOffTime,
		rec.FlightDurationMinutes, rec.IsLocal, rec.ImportSource, rec.RawImportID,
	)
	return err
}
