package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingPlanRepo struct{ db *DB }

func (db *DB) TrainingPlan() repository.TrainingPlanRepository { return &TrainingPlanRepo{db: db} }

var _ repository.TrainingPlanRepository = (*TrainingPlanRepo)(nil)

func (r *TrainingPlanRepo) Create(ctx context.Context, p *domain.TrainingPlan) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_plans (student_id, instructor_id, aircraft_id, plan_date, start_time, end_time, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		p.StudentID, p.InstructorID, p.AircraftID, p.PlanDate, p.StartTime, p.EndTime, p.Source,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *TrainingPlanRepo) ListByDate(ctx context.Context, from, to string) ([]*domain.TrainingPlan, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, student_id, instructor_id, aircraft_id, plan_date, start_time, end_time, source, created_at, updated_at
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
