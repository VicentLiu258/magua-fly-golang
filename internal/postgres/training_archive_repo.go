package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingArchiveRepo struct{ db *DB }

func (db *DB) TrainingArchive() repository.TrainingArchiveRepository { return &TrainingArchiveRepo{db: db} }

var _ repository.TrainingArchiveRepository = (*TrainingArchiveRepo)(nil)

func (r *TrainingArchiveRepo) GetByID(ctx context.Context, id int64) (*domain.TrainingArchive, error) {
	var a domain.TrainingArchive
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, training_record_id, student_id, instructor_id, training_date, COALESCE(stage,''),
		       planned_duration_minutes, total_experience_minutes, remaining_minutes,
		       engine_on_time, engine_off_time, flight_duration_minutes,
		       COALESCE(progress_note,''), COALESCE(instructor_review,''), COALESCE(student_self_review,''), COALESCE(task_assignment,''),
		       student_signed_at, responsible_instructor_signed_at, chief_instructor_signed_at, created_at, updated_at
		FROM training_archives WHERE id = $1`, id).Scan(
		&a.ID, &a.TrainingRecordID, &a.StudentID, &a.InstructorID, &a.TrainingDate, &a.Stage,
		&a.PlannedDurationMinutes, &a.TotalExperienceMinutes, &a.RemainingMinutes,
		&a.EngineOnTime, &a.EngineOffTime, &a.FlightDurationMinutes,
		&a.ProgressNote, &a.InstructorReview, &a.StudentSelfReview, &a.TaskAssignment,
		&a.StudentSignedAt, &a.ResponsibleInstructorSignedAt, &a.ChiefInstructorSignedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *TrainingArchiveRepo) GetByTrainingRecordID(ctx context.Context, recordID int64) (*domain.TrainingArchive, error) {
	var a domain.TrainingArchive
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, training_record_id, student_id, instructor_id, training_date, COALESCE(stage,''),
		       planned_duration_minutes, total_experience_minutes, remaining_minutes,
		       engine_on_time, engine_off_time, flight_duration_minutes,
		       COALESCE(progress_note,''), COALESCE(instructor_review,''), COALESCE(student_self_review,''), COALESCE(task_assignment,''),
		       student_signed_at, responsible_instructor_signed_at, chief_instructor_signed_at, created_at, updated_at
		FROM training_archives WHERE training_record_id = $1`, recordID).Scan(
		&a.ID, &a.TrainingRecordID, &a.StudentID, &a.InstructorID, &a.TrainingDate, &a.Stage,
		&a.PlannedDurationMinutes, &a.TotalExperienceMinutes, &a.RemainingMinutes,
		&a.EngineOnTime, &a.EngineOffTime, &a.FlightDurationMinutes,
		&a.ProgressNote, &a.InstructorReview, &a.StudentSelfReview, &a.TaskAssignment,
		&a.StudentSignedAt, &a.ResponsibleInstructorSignedAt, &a.ChiefInstructorSignedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *TrainingArchiveRepo) Create(ctx context.Context, a *domain.TrainingArchive) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_archives (training_record_id, student_id, instructor_id, training_date, stage,
			planned_duration_minutes, total_experience_minutes, remaining_minutes,
			engine_on_time, engine_off_time, flight_duration_minutes,
			progress_note, instructor_review, student_self_review, task_assignment,
			student_signed_at, responsible_instructor_signed_at, chief_instructor_signed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at`,
		a.TrainingRecordID, a.StudentID, a.InstructorID, a.TrainingDate, a.Stage,
		a.PlannedDurationMinutes, a.TotalExperienceMinutes, a.RemainingMinutes,
		a.EngineOnTime, a.EngineOffTime, a.FlightDurationMinutes,
		a.ProgressNote, a.InstructorReview, a.StudentSelfReview, a.TaskAssignment,
		a.StudentSignedAt, a.ResponsibleInstructorSignedAt, a.ChiefInstructorSignedAt,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *TrainingArchiveRepo) Update(ctx context.Context, a *domain.TrainingArchive) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE training_archives SET stage=$2, planned_duration_minutes=$3, total_experience_minutes=$4, remaining_minutes=$5,
			engine_on_time=$6, engine_off_time=$7, flight_duration_minutes=$8,
			progress_note=$9, instructor_review=$10, student_self_review=$11, task_assignment=$12,
			student_signed_at=$13, responsible_instructor_signed_at=$14, chief_instructor_signed_at=$15, updated_at=NOW()
		WHERE id=$1`,
		a.ID, a.Stage, a.PlannedDurationMinutes, a.TotalExperienceMinutes, a.RemainingMinutes,
		a.EngineOnTime, a.EngineOffTime, a.FlightDurationMinutes,
		a.ProgressNote, a.InstructorReview, a.StudentSelfReview, a.TaskAssignment,
		a.StudentSignedAt, a.ResponsibleInstructorSignedAt, a.ChiefInstructorSignedAt,
	)
	return err
}
