package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type BookingRepo struct{ db *DB }

func (db *DB) Booking() repository.BookingRepository { return &BookingRepo{db: db} }

var _ repository.BookingRepository = (*BookingRepo)(nil)

func (r *BookingRepo) GetByID(ctx context.Context, id int64) (*domain.Booking, error) {
	var b domain.Booking
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, student_id, instructor_id, aircraft_id, plan_date, start_time, end_time,
		       status, COALESCE(reject_reason,''), reviewed_at, reviewer_id, created_at, updated_at
		FROM bookings WHERE id = $1`, id).Scan(
		&b.ID, &b.StudentID, &b.InstructorID, &b.AircraftID, &b.PlanDate, &b.StartTime, &b.EndTime,
		&b.Status, &b.RejectReason, &b.ReviewedAt, &b.ReviewerID, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookingRepo) ListByStudent(ctx context.Context, studentID int64, status *domain.BookingStatus) ([]*domain.Booking, error) {
	query := `
		SELECT id, student_id, instructor_id, aircraft_id, plan_date, start_time, end_time,
		       status, COALESCE(reject_reason,''), reviewed_at, reviewer_id, created_at, updated_at
		FROM bookings WHERE student_id = $1`
	args := []interface{}{studentID}
	if status != nil {
		query += ` AND status = $2 ORDER BY plan_date DESC, start_time`
		args = append(args, *status)
	} else {
		query += ` ORDER BY plan_date DESC, start_time`
	}
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.StudentID, &b.InstructorID, &b.AircraftID, &b.PlanDate, &b.StartTime, &b.EndTime,
			&b.Status, &b.RejectReason, &b.ReviewedAt, &b.ReviewerID, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &b)
	}
	return list, rows.Err()
}

func (r *BookingRepo) ListPending(ctx context.Context) ([]*domain.Booking, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, student_id, instructor_id, aircraft_id, plan_date, start_time, end_time,
		       status, COALESCE(reject_reason,''), reviewed_at, reviewer_id, created_at, updated_at
		FROM bookings WHERE status = 'pending' ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.StudentID, &b.InstructorID, &b.AircraftID, &b.PlanDate, &b.StartTime, &b.EndTime,
			&b.Status, &b.RejectReason, &b.ReviewedAt, &b.ReviewerID, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &b)
	}
	return list, rows.Err()
}

func (r *BookingRepo) Create(ctx context.Context, b *domain.Booking) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO bookings (student_id, instructor_id, aircraft_id, plan_date, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		b.StudentID, b.InstructorID, b.AircraftID, b.PlanDate, b.StartTime, b.EndTime, b.Status,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id int64, status domain.BookingStatus, rejectReason string, reviewerID int64) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE bookings SET status=$2, reject_reason=$3, reviewed_at=NOW(), reviewer_id=$4, updated_at=NOW() WHERE id=$1`,
		id, status, rejectReason, reviewerID,
	)
	return err
}
