package postgres

import (
	"context"
	"encoding/json"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type ContractRepo struct{ db *DB }

func (db *DB) Contract() repository.ContractRepository { return &ContractRepo{db: db} }

var _ repository.ContractRepository = (*ContractRepo)(nil)

func contractPaymentScheduleJSON(c *domain.Contract) ([]byte, error) {
	if c.PaymentSchedule == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(c.PaymentSchedule)
}

func scanContractPaymentSchedule(raw []byte, c *domain.Contract) error {
	if len(raw) == 0 {
		c.PaymentSchedule = nil
		return nil
	}
	var sched []domain.PaymentInstallment
	if err := json.Unmarshal(raw, &sched); err != nil {
		return err
	}
	c.PaymentSchedule = sched
	return nil
}

func (r *ContractRepo) GetByID(ctx context.Context, id string) (*domain.Contract, error) {
	var schedBytes []byte
	var c domain.Contract
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text,
		       student_id::text,
		       contract_no,
		       sign_date,
		       start_date,
		       end_date,
		       total_minutes,
		       price_total,
		       currency,
		       status,
		       payment_mode,
		       payment_schedule,
		       mongo_file_ids,
		       remark,
		       created_at,
		       updated_at
		FROM contracts
		WHERE id = $1::uuid`,
		id,
	).Scan(
		&c.ID,
		&c.StudentID,
		&c.ContractNo,
		&c.SignDate,
		&c.StartDate,
		&c.EndDate,
		&c.TotalMinutes,
		&c.PriceTotal,
		&c.Currency,
		&c.Status,
		&c.PaymentMode,
		&schedBytes,
		&c.MongoFileIDs,
		&c.Remark,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := scanContractPaymentSchedule(schedBytes, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContractRepo) ListByStudent(ctx context.Context, studentID string) ([]*domain.Contract, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text,
		       student_id::text,
		       contract_no,
		       sign_date,
		       start_date,
		       end_date,
		       total_minutes,
		       price_total,
		       currency,
		       status,
		       payment_mode,
		       payment_schedule,
		       mongo_file_ids,
		       remark,
		       created_at,
		       updated_at
		FROM contracts
		WHERE student_id = $1::uuid
		ORDER BY created_at DESC`,
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Contract
	for rows.Next() {
		var c domain.Contract
		var schedBytes []byte
		if err := rows.Scan(
			&c.ID,
			&c.StudentID,
			&c.ContractNo,
			&c.SignDate,
			&c.StartDate,
			&c.EndDate,
			&c.TotalMinutes,
			&c.PriceTotal,
			&c.Currency,
			&c.Status,
			&c.PaymentMode,
			&schedBytes,
			&c.MongoFileIDs,
			&c.Remark,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err := scanContractPaymentSchedule(schedBytes, &c); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (r *ContractRepo) Create(ctx context.Context, c *domain.Contract) error {
	schedJSON, err := contractPaymentScheduleJSON(c)
	if err != nil {
		return err
	}
	if c.PaymentMode == "" {
		c.PaymentMode = domain.PaymentModeFull
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO contracts (
			student_id,
			contract_no,
			sign_date,
			start_date,
			end_date,
			total_minutes,
			price_total,
			currency,
			status,
			payment_mode,
			payment_schedule,
			mongo_file_ids,
			remark
		) VALUES (
			$1::uuid,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11::jsonb,
			$12,
			$13
		)
		RETURNING id::text, created_at, updated_at`,
		c.StudentID,
		c.ContractNo,
		c.SignDate,
		c.StartDate,
		c.EndDate,
		c.TotalMinutes,
		c.PriceTotal,
		c.Currency,
		c.Status,
		c.PaymentMode,
		schedJSON,
		c.MongoFileIDs,
		c.Remark,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ContractRepo) Update(ctx context.Context, c *domain.Contract) error {
	schedJSON, err := contractPaymentScheduleJSON(c)
	if err != nil {
		return err
	}
	if c.PaymentMode == "" {
		c.PaymentMode = domain.PaymentModeFull
	}
	_, err = r.db.Pool.Exec(ctx, `
		UPDATE contracts
		SET contract_no = $2,
		    sign_date = $3,
		    start_date = $4,
		    end_date = $5,
		    total_minutes = $6,
		    price_total = $7,
		    currency = $8,
		    status = $9,
		    payment_mode = $10,
		    payment_schedule = $11::jsonb,
		    mongo_file_ids = $12,
		    remark = $13,
		    updated_at = NOW()
		WHERE id = $1::uuid`,
		c.ID,
		c.ContractNo,
		c.SignDate,
		c.StartDate,
		c.EndDate,
		c.TotalMinutes,
		c.PriceTotal,
		c.Currency,
		c.Status,
		c.PaymentMode,
		schedJSON,
		c.MongoFileIDs,
		c.Remark,
	)
	return err
}
