package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingRecordRepo struct{ db *DB }

func (db *DB) TrainingRecord() repository.TrainingRecordRepository { return &TrainingRecordRepo{db: db} }

var _ repository.TrainingRecordRepository = (*TrainingRecordRepo)(nil)

var trainingRecordColumns = `id::text, student_id::text, instructor_id::text, aircraft_id::text, training_date, engine_on_time, engine_off_time,
	flight_duration_minutes, is_local, import_source, COALESCE(raw_import_id::text,''), COALESCE(status::text,'normal'),
	planned_duration_minutes, theory_duration_minutes, sim_duration_minutes, COALESCE(syllabus_outline,''), COALESCE(archive_display_no,''),
	COALESCE(departure,''), COALESCE(destination,''), COALESCE(flight_type,''), COALESCE(task_type,''), landings,
	COALESCE(training_plan_id::text,''),
	created_at, updated_at`

func (r *TrainingRecordRepo) GetByID(ctx context.Context, id string) (*domain.TrainingRecord, error) {
	var rec domain.TrainingRecord
	err := r.db.Pool.QueryRow(ctx, `SELECT `+trainingRecordColumns+` FROM training_records WHERE id = $1::uuid`, id).Scan(
		&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
		&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
		&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
		&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *TrainingRecordRepo) FindByTrainingPlanID(ctx context.Context, trainingPlanID string) (*domain.TrainingRecord, error) {
	var rec domain.TrainingRecord
	err := r.db.Pool.QueryRow(ctx, `SELECT `+trainingRecordColumns+` FROM training_records WHERE training_plan_id = $1::uuid`, trainingPlanID).Scan(
		&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
		&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
		&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
		&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *TrainingRecordRepo) ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingRecord, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT `+trainingRecordColumns+` FROM training_records WHERE student_id = $1::uuid ORDER BY training_date DESC LIMIT $2 OFFSET $3`, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingRecord
	for rows.Next() {
		var rec domain.TrainingRecord
		if err := rows.Scan(
			&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
			&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
			&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
			&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &rec)
	}
	return list, rows.Err()
}

func (r *TrainingRecordRepo) List(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit, offset int) ([]*domain.TrainingRecord, int, error) {
	var conds []string
	var args []interface{}
	argNum := 1
	if filter != nil {
		if filter.StudentID != nil {
			conds = append(conds, fmt.Sprintf("student_id = $%d::uuid", argNum))
			args = append(args, *filter.StudentID)
			argNum++
		}
		if filter.InstructorID != nil {
			conds = append(conds, fmt.Sprintf("instructor_id = $%d::uuid", argNum))
			args = append(args, *filter.InstructorID)
			argNum++
		}
		if filter.DateFrom != nil {
			conds = append(conds, fmt.Sprintf("training_date >= $%d", argNum))
			args = append(args, *filter.DateFrom)
			argNum++
		}
		if filter.DateTo != nil {
			conds = append(conds, fmt.Sprintf("training_date <= $%d", argNum))
			args = append(args, *filter.DateTo)
			argNum++
		}
		if filter.Status != nil {
			conds = append(conds, fmt.Sprintf("status = $%d", argNum))
			args = append(args, string(*filter.Status))
			argNum++
		}
		if filter.ImportSource != nil {
			conds = append(conds, fmt.Sprintf("import_source = $%d", argNum))
			args = append(args, string(*filter.ImportSource))
			argNum++
		}
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	order := " ORDER BY training_date DESC"
	if sort != nil && sort.Field != "" {
		order = " ORDER BY " + sort.Field
		if !sort.Asc {
			order += " DESC"
		} else {
			order += " ASC"
		}
	}
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM training_records`+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitOff := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
	rows, err := r.db.Pool.Query(ctx, `SELECT `+trainingRecordColumns+` FROM training_records`+where+order+limitOff, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.TrainingRecord
	for rows.Next() {
		var rec domain.TrainingRecord
		if err := rows.Scan(
			&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
			&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
			&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
			&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *TrainingRecordRepo) Create(ctx context.Context, rec *domain.TrainingRecord) error {
	status := rec.Status
	if status == "" {
		status = domain.RecordStatusNormal
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_records (student_id, instructor_id, aircraft_id, training_date, engine_on_time, engine_off_time,
			flight_duration_minutes, is_local, import_source, raw_import_id, status,
			planned_duration_minutes, theory_duration_minutes, sim_duration_minutes, syllabus_outline, archive_display_no,
			departure, destination, flight_type, task_type, landings, training_plan_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10::uuid, $11::training_record_status, $12, $13, $14, NULLIF($15,''), NULLIF($16,''), NULLIF($17,''), NULLIF($18,''), NULLIF($19,''), NULLIF($20,''), $21, $22::uuid)
		RETURNING id::text, created_at, updated_at`,
		rec.StudentID, rec.InstructorID, rec.AircraftID, rec.TrainingDate, rec.EngineOnTime, rec.EngineOffTime,
		rec.FlightDurationMinutes, rec.IsLocal, rec.ImportSource, nullEmpty(rec.RawImportID), status,
		rec.PlannedDurationMinutes, rec.TheoryDurationMinutes, rec.SimDurationMinutes, rec.SyllabusOutline, rec.ArchiveDisplayNo,
		rec.Departure, rec.Destination, rec.FlightType, rec.TaskType, rec.Landings, nullEmpty(rec.TrainingPlanID),
	).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt)
}

func nullEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *TrainingRecordRepo) Update(ctx context.Context, rec *domain.TrainingRecord) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE training_records SET student_id=$2::uuid, instructor_id=$3::uuid, aircraft_id=$4::uuid, training_date=$5,
			engine_on_time=$6, engine_off_time=$7, flight_duration_minutes=$8, is_local=$9, import_source=$10, raw_import_id=$11::uuid,
			status=$12::training_record_status, planned_duration_minutes=$13, theory_duration_minutes=$14, sim_duration_minutes=$15,
			syllabus_outline=NULLIF($16,''), archive_display_no=NULLIF($17,''),
			departure=NULLIF($18,''), destination=NULLIF($19,''), flight_type=NULLIF($20,''), task_type=NULLIF($21,''), landings=$22,
			training_plan_id=$23::uuid, updated_at=NOW()
		WHERE id=$1::uuid`,
		rec.ID, rec.StudentID, rec.InstructorID, rec.AircraftID, rec.TrainingDate, rec.EngineOnTime, rec.EngineOffTime,
		rec.FlightDurationMinutes, rec.IsLocal, rec.ImportSource, nullEmpty(rec.RawImportID),
		rec.Status, rec.PlannedDurationMinutes, rec.TheoryDurationMinutes, rec.SimDurationMinutes, rec.SyllabusOutline, rec.ArchiveDisplayNo,
		rec.Departure, rec.Destination, rec.FlightType, rec.TaskType, rec.Landings, nullEmpty(rec.TrainingPlanID),
	)
	return err
}

func (r *TrainingRecordRepo) UpdateStatus(ctx context.Context, id string, status domain.TrainingRecordStatus) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE training_records SET status = $2::training_record_status, updated_at = NOW() WHERE id = $1::uuid`, id, string(status))
	return err
}
