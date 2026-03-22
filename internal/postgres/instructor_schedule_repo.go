package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type InstructorScheduleRepo struct{ db *DB }

func (db *DB) InstructorSchedule() repository.InstructorScheduleRepository {
	return &InstructorScheduleRepo{db: db}
}

var _ repository.InstructorScheduleRepository = (*InstructorScheduleRepo)(nil)

func scanInstructorSchedule(row pgx.Row) (*domain.InstructorSchedule, error) {
	var s domain.InstructorSchedule
	var dutyWindows []byte
	err := row.Scan(&s.ID, &s.InstructorID, &s.Date, &s.DutyStatus, &s.Remark, &dutyWindows, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.DutyWindowsJSON = dutyWindows
	return &s, nil
}

func (r *InstructorScheduleRepo) ListByDate(ctx context.Context, date time.Time) ([]*domain.InstructorSchedule, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, instructor_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
		FROM instructor_schedules
		WHERE date = $1
		ORDER BY instructor_id`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.InstructorSchedule
	for rows.Next() {
		s, err := scanInstructorSchedule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *InstructorScheduleRepo) ListByDateRange(ctx context.Context, from, to time.Time, instructorIDs []string) ([]*domain.InstructorSchedule, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

	var (
		rows pgx.Rows
		err  error
	)
	if len(instructorIDs) == 0 {
		rows, err = r.db.Pool.Query(ctx, `
			SELECT id::text, instructor_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
			FROM instructor_schedules
			WHERE date >= $1 AND date <= $2
			ORDER BY date, instructor_id`, fromD, toD)
	} else {
		rows, err = r.db.Pool.Query(ctx, `
			SELECT id::text, instructor_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
			FROM instructor_schedules
			WHERE date >= $1 AND date <= $2 AND instructor_id = ANY($3::uuid[])
			ORDER BY date, instructor_id`, fromD, toD, instructorIDs)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.InstructorSchedule
	for rows.Next() {
		s, err := scanInstructorSchedule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *InstructorScheduleRepo) GetByInstructorAndDate(ctx context.Context, instructorID string, date time.Time) (*domain.InstructorSchedule, error) {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	s, err := scanInstructorSchedule(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, instructor_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
		FROM instructor_schedules
		WHERE instructor_id = $1::uuid AND date = $2`, instructorID, day))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *InstructorScheduleRepo) UpsertDay(ctx context.Context, s *domain.InstructorSchedule) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO instructor_schedules (instructor_id, date, duty_status, remark, duty_windows, updated_at)
		VALUES ($1::uuid, $2::date, $3::instructor_duty_status, $4, $5, NOW())
		ON CONFLICT (instructor_id, date) DO UPDATE SET
			duty_status = EXCLUDED.duty_status,
			remark = EXCLUDED.remark,
			duty_windows = EXCLUDED.duty_windows,
			updated_at = NOW()
		RETURNING id::text, created_at, updated_at`,
		s.InstructorID, s.Date, s.DutyStatus, nullStr(s.Remark), s.DutyWindowsJSON,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}
