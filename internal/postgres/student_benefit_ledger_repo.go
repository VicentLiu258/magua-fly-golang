package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type studentBenefitLedgerRepo struct{ db *DB }

func (db *DB) StudentBenefitLedger() repository.StudentBenefitLedgerRepository {
	return &studentBenefitLedgerRepo{db: db}
}

var _ repository.StudentBenefitLedgerRepository = (*studentBenefitLedgerRepo)(nil)

func (r *studentBenefitLedgerRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.StudentBenefitLedger, error) {
	if key == "" {
		return nil, nil
	}
	var row domain.StudentBenefitLedger
	var opUID *string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text,
		       student_id::text,
		       benefit_kind,
		       direction,
		       amount::float8,
		       COALESCE(reason,''),
		       operator_user_id::text,
		       COALESCE(operator_name,''),
		       created_at,
		       COALESCE(idempotency_key,'')
		FROM student_benefit_ledger
		WHERE idempotency_key = $1
		LIMIT 1`, key,
	).Scan(
		&row.ID,
		&row.StudentID,
		&row.BenefitKind,
		&row.Direction,
		&row.Amount,
		&row.Reason,
		&opUID,
		&row.OperatorName,
		&row.CreatedAt,
		&row.IdempotencyKey,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	row.OperatorUserID = opUID
	return &row, nil
}

func (r *studentBenefitLedgerRepo) ListByStudentKind(ctx context.Context, studentID, benefitKind string, limit, offset int) ([]*domain.StudentBenefitLedger, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM student_benefit_ledger
		WHERE student_id = $1::uuid AND benefit_kind = $2`, studentID, benefitKind,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text,
		       student_id::text,
		       benefit_kind,
		       direction,
		       amount::float8,
		       COALESCE(reason,''),
		       operator_user_id::text,
		       COALESCE(operator_name,''),
		       created_at,
		       COALESCE(idempotency_key,'')
		FROM student_benefit_ledger
		WHERE student_id = $1::uuid AND benefit_kind = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, studentID, benefitKind, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.StudentBenefitLedger
	for rows.Next() {
		var row domain.StudentBenefitLedger
		var opUID *string
		if err := rows.Scan(
			&row.ID,
			&row.StudentID,
			&row.BenefitKind,
			&row.Direction,
			&row.Amount,
			&row.Reason,
			&opUID,
			&row.OperatorName,
			&row.CreatedAt,
			&row.IdempotencyKey,
		); err != nil {
			return nil, 0, err
		}
		row.OperatorUserID = opUID
		list = append(list, &row)
	}
	return list, total, rows.Err()
}

func (r *studentBenefitLedgerRepo) MaxCreatedAt(ctx context.Context, studentID string) (*time.Time, error) {
	var nt sql.NullTime
	err := r.db.Pool.QueryRow(ctx, `
		SELECT MAX(created_at) FROM student_benefit_ledger WHERE student_id = $1::uuid`, studentID,
	).Scan(&nt)
	if err != nil {
		return nil, err
	}
	if !nt.Valid {
		return nil, nil
	}
	t := nt.Time
	return &t, nil
}
