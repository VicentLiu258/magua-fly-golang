package postgres

import (
	"context"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type InstructorScheduleRepo struct{ db *DB }

func (db *DB) InstructorSchedule() repository.InstructorScheduleRepository {
	return &InstructorScheduleRepo{db: db}
}

var _ repository.InstructorScheduleRepository = (*InstructorScheduleRepo)(nil)

func (r *InstructorScheduleRepo) ListByDate(ctx context.Context, date time.Time) ([]*domain.InstructorSchedule, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, instructor_id::text, date, start_time, end_time, created_at, updated_at
		FROM instructor_schedules
		WHERE date = $1
		ORDER BY instructor_id, start_time`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.InstructorSchedule
	for rows.Next() {
		var s domain.InstructorSchedule
		if err := rows.Scan(&s.ID, &s.InstructorID, &s.Date, &s.StartTime, &s.EndTime, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

