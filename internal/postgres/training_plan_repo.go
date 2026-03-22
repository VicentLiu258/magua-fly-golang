package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingPlanRepo struct{ db *DB }

func (db *DB) TrainingPlan() repository.TrainingPlanRepository { return &TrainingPlanRepo{db: db} }

var _ repository.TrainingPlanRepository = (*TrainingPlanRepo)(nil)

func (r *TrainingPlanRepo) Create(ctx context.Context, p *domain.TrainingPlan) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_plans (student_id, instructor_id, aircraft_id, plan_date, start_time, end_time, source)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7)
		RETURNING id::text, created_at, updated_at`,
		p.StudentID, p.InstructorID, p.AircraftID, p.PlanDate, p.StartTime, p.EndTime, p.Source,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *TrainingPlanRepo) DeleteMatchingBookingFields(ctx context.Context, studentID, instructorID, aircraftID string, planDate time.Time, startTime, endTime time.Time) (deletedID *string, err error) {
	var id string
	err = r.db.Pool.QueryRow(ctx, `
		DELETE FROM training_plans
		WHERE student_id = $1::uuid AND instructor_id = $2::uuid AND aircraft_id = $3::uuid
		  AND plan_date = $4::date AND start_time = $5 AND end_time = $6
		RETURNING id::text`,
		studentID, instructorID, aircraftID, planDate, startTime, endTime,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *TrainingPlanRepo) ListByDate(ctx context.Context, from, to string) ([]*domain.TrainingPlan, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, student_id::text, instructor_id::text, aircraft_id::text, plan_date, start_time, end_time, source, created_at, updated_at
		FROM training_plans WHERE plan_date >= $1 AND plan_date <= $2 ORDER BY plan_date, start_time`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingPlan
	for rows.Next() {
		var p domain.TrainingPlan
		if err := rows.Scan(&p.ID, &p.StudentID, &p.InstructorID, &p.AircraftID, &p.PlanDate, &p.StartTime, &p.EndTime, &p.Source, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}
