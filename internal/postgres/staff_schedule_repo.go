package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type StaffScheduleRepo struct{ db *DB }

func (db *DB) StaffSchedule() repository.StaffScheduleRepository {
	return &StaffScheduleRepo{db: db}
}

var _ repository.StaffScheduleRepository = (*StaffScheduleRepo)(nil)

func scanStaffSchedule(row pgx.Row) (*domain.StaffSchedule, error) {
	var s domain.StaffSchedule
	var dutyWindows []byte
	err := row.Scan(&s.ID, &s.UserID, &s.Date, &s.DutyStatus, &s.Remark, &dutyWindows, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.DutyWindowsJSON = dutyWindows
	return &s, nil
}

func (r *StaffScheduleRepo) ListByDate(ctx context.Context, date time.Time) ([]*domain.StaffSchedule, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, user_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
		FROM staff_schedules
		WHERE date = $1
		ORDER BY user_id`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.StaffSchedule
	for rows.Next() {
		s, err := scanStaffSchedule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *StaffScheduleRepo) ListByDateRange(ctx context.Context, from, to time.Time, userIDs []string) ([]*domain.StaffSchedule, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	var (
		rows pgx.Rows
		err  error
	)
	if len(userIDs) == 0 {
		rows, err = r.db.Pool.Query(ctx, `
			SELECT id::text, user_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
			FROM staff_schedules
			WHERE date >= $1 AND date <= $2
			ORDER BY date, user_id`, fromD, toD)
	} else {
		rows, err = r.db.Pool.Query(ctx, `
			SELECT id::text, user_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
			FROM staff_schedules
			WHERE date >= $1 AND date <= $2 AND user_id = ANY($3::uuid[])
			ORDER BY date, user_id`, fromD, toD, userIDs)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.StaffSchedule
	for rows.Next() {
		s, err := scanStaffSchedule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *StaffScheduleRepo) GetByUserAndDate(ctx context.Context, userID string, date time.Time) (*domain.StaffSchedule, error) {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	s, err := scanStaffSchedule(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, date, duty_status::text, COALESCE(remark,''), duty_windows, created_at, updated_at
		FROM staff_schedules
		WHERE user_id = $1::uuid AND date = $2`, userID, day))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *StaffScheduleRepo) UpsertDay(ctx context.Context, s *domain.StaffSchedule) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO staff_schedules (user_id, date, duty_status, remark, duty_windows, updated_at)
		VALUES ($1::uuid, $2::date, $3::instructor_duty_status, $4, $5, NOW())
		ON CONFLICT (user_id, date) DO UPDATE SET
			duty_status = EXCLUDED.duty_status,
			remark = EXCLUDED.remark,
			duty_windows = EXCLUDED.duty_windows,
			updated_at = NOW()
		RETURNING id::text, created_at, updated_at`,
		s.UserID, s.Date, s.DutyStatus, nullStr(s.Remark), s.DutyWindowsJSON,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}
