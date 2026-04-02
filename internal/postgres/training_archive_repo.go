package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingArchiveRepo struct{ db *DB }

func (db *DB) TrainingArchive() repository.TrainingArchiveRepository { return &TrainingArchiveRepo{db: db} }

var _ repository.TrainingArchiveRepository = (*TrainingArchiveRepo)(nil)

var trainingArchiveColumns = `id::text, training_record_id::text, student_id::text, COALESCE(student_no,''), instructor_id::text, training_date, COALESCE(stage,''),
	planned_duration_minutes, total_experience_minutes, remaining_minutes,
	engine_on_time, engine_off_time, flight_duration_minutes,
	COALESCE(progress_note,''), COALESCE(instructor_review,''), COALESCE(student_self_review,''), COALESCE(task_assignment,''),
	student_signed_at, responsible_instructor_signed_at, chief_instructor_signed_at,
	responsible_instructor_id::text, chief_instructor_id::text, created_at, updated_at`

func (r *TrainingArchiveRepo) GetByID(ctx context.Context, id string) (*domain.TrainingArchive, error) {
	var a domain.TrainingArchive
	err := r.db.Pool.QueryRow(ctx, `SELECT `+trainingArchiveColumns+` FROM training_archives WHERE id = $1::uuid`, id).Scan(
		&a.ID, &a.TrainingRecordID, &a.StudentID, &a.StudentNo, &a.InstructorID, &a.TrainingDate, &a.Stage,
		&a.PlannedDurationMinutes, &a.TotalExperienceMinutes, &a.RemainingMinutes,
		&a.EngineOnTime, &a.EngineOffTime, &a.FlightDurationMinutes,
		&a.ProgressNote, &a.InstructorReview, &a.StudentSelfReview, &a.TaskAssignment,
		&a.StudentSignedAt, &a.ResponsibleInstructorSignedAt, &a.ChiefInstructorSignedAt,
		&a.ResponsibleInstructorID, &a.ChiefInstructorID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *TrainingArchiveRepo) GetByTrainingRecordID(ctx context.Context, recordID string) (*domain.TrainingArchive, error) {
	var a domain.TrainingArchive
	err := r.db.Pool.QueryRow(ctx, `SELECT `+trainingArchiveColumns+` FROM training_archives WHERE training_record_id = $1::uuid`, recordID).Scan(
		&a.ID, &a.TrainingRecordID, &a.StudentID, &a.StudentNo, &a.InstructorID, &a.TrainingDate, &a.Stage,
		&a.PlannedDurationMinutes, &a.TotalExperienceMinutes, &a.RemainingMinutes,
		&a.EngineOnTime, &a.EngineOffTime, &a.FlightDurationMinutes,
		&a.ProgressNote, &a.InstructorReview, &a.StudentSelfReview, &a.TaskAssignment,
		&a.StudentSignedAt, &a.ResponsibleInstructorSignedAt, &a.ChiefInstructorSignedAt,
		&a.ResponsibleInstructorID, &a.ChiefInstructorID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *TrainingArchiveRepo) ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingArchive, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT `+trainingArchiveColumns+` FROM training_archives WHERE student_id = $1::uuid ORDER BY training_date DESC LIMIT $2 OFFSET $3`, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingArchive
	for rows.Next() {
		var a domain.TrainingArchive
		if err := rows.Scan(
			&a.ID, &a.TrainingRecordID, &a.StudentID, &a.StudentNo, &a.InstructorID, &a.TrainingDate, &a.Stage,
			&a.PlannedDurationMinutes, &a.TotalExperienceMinutes, &a.RemainingMinutes,
			&a.EngineOnTime, &a.EngineOffTime, &a.FlightDurationMinutes,
			&a.ProgressNote, &a.InstructorReview, &a.StudentSelfReview, &a.TaskAssignment,
			&a.StudentSignedAt, &a.ResponsibleInstructorSignedAt, &a.ChiefInstructorSignedAt,
			&a.ResponsibleInstructorID, &a.ChiefInstructorID, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *TrainingArchiveRepo) Create(ctx context.Context, a *domain.TrainingArchive) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_archives (training_record_id, student_id, student_no, instructor_id, training_date, stage,
			planned_duration_minutes, total_experience_minutes, remaining_minutes,
			engine_on_time, engine_off_time, flight_duration_minutes,
			progress_note, instructor_review, student_self_review, task_assignment,
			student_signed_at, responsible_instructor_signed_at, chief_instructor_signed_at,
			responsible_instructor_id, chief_instructor_id)
		VALUES ($1::uuid, $2::uuid, NULLIF(TRIM($3),''), $4::uuid, $5, NULLIF($6,''), $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20::uuid, $21::uuid)
		RETURNING id::text, created_at, updated_at`,
		a.TrainingRecordID, a.StudentID, a.StudentNo, a.InstructorID, a.TrainingDate, a.Stage,
		a.PlannedDurationMinutes, a.TotalExperienceMinutes, a.RemainingMinutes,
		a.EngineOnTime, a.EngineOffTime, a.FlightDurationMinutes,
		a.ProgressNote, a.InstructorReview, a.StudentSelfReview, a.TaskAssignment,
		a.StudentSignedAt, a.ResponsibleInstructorSignedAt, a.ChiefInstructorSignedAt,
		nullUUID(a.ResponsibleInstructorID), nullUUID(a.ChiefInstructorID),
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *TrainingArchiveRepo) Update(ctx context.Context, a *domain.TrainingArchive) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE training_archives SET student_no=NULLIF(TRIM($2),''), stage=NULLIF($3,''), planned_duration_minutes=$4, total_experience_minutes=$5, remaining_minutes=$6,
			engine_on_time=$7, engine_off_time=$8, flight_duration_minutes=$9,
			progress_note=$10, instructor_review=$11, student_self_review=$12, task_assignment=$13,
			student_signed_at=$14, responsible_instructor_signed_at=$15, chief_instructor_signed_at=$16,
			responsible_instructor_id=$17::uuid, chief_instructor_id=$18::uuid, updated_at=NOW()
		WHERE id=$1::uuid`,
		a.ID, a.StudentNo, a.Stage, a.PlannedDurationMinutes, a.TotalExperienceMinutes, a.RemainingMinutes,
		a.EngineOnTime, a.EngineOffTime, a.FlightDurationMinutes,
		a.ProgressNote, a.InstructorReview, a.StudentSelfReview, a.TaskAssignment,
		a.StudentSignedAt, a.ResponsibleInstructorSignedAt, a.ChiefInstructorSignedAt,
		nullUUID(a.ResponsibleInstructorID), nullUUID(a.ChiefInstructorID),
	)
	return err
}
