package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/pkg/studentcontract"
	"magua-fly-golang/internal/repository"
)

type studentBenefitsAdjuster struct{ db *DB }

func (db *DB) StudentBenefitsAdjuster() repository.StudentBenefitsAdjuster {
	return &studentBenefitsAdjuster{db: db}
}

var _ repository.StudentBenefitsAdjuster = (*studentBenefitsAdjuster)(nil)

func (a *studentBenefitsAdjuster) Adjust(ctx context.Context, op *domain.BenefitAdjustOp) error {
	if op == nil || op.StudentID == "" {
		return errors.New("invalid adjust op")
	}
	if op.Amount <= 0 || math.IsNaN(op.Amount) {
		return errors.New("invalid amount")
	}
	sr := &StudentRepo{db: a.db}
	return a.db.WithTx(ctx, func(tx pgx.Tx) error {
		s, err := sr.getByIDForUpdateTx(ctx, tx, op.StudentID)
		if err != nil {
			return err
		}
		newCI, ledgerAmount, err := applyBenefitAdjustToContractInfo(s.ContractInfo, op)
		if err != nil {
			if errors.Is(err, repository.ErrBenefitDebitExceeds) {
				return repository.ErrBenefitDebitExceeds
			}
			return err
		}
		s.ContractInfo = newCI
		if err := sr.updateTx(ctx, tx, s); err != nil {
			return err
		}
		return insertBenefitLedgerRow(ctx, tx, op, ledgerAmount)
	})
}

func insertBenefitLedgerRow(ctx context.Context, tx pgx.Tx, op *domain.BenefitAdjustOp, amount float64) error {
	var opUID interface{}
	if op.OperatorUserID != nil && *op.OperatorUserID != "" {
		opUID = *op.OperatorUserID
	}
	var idem interface{}
	if op.IdempotencyKey != "" {
		idem = op.IdempotencyKey
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO student_benefit_ledger (
			student_id, benefit_kind, direction, amount, reason,
			operator_user_id, operator_name, idempotency_key
		) VALUES (
			$1::uuid, $2, $3, $4, $5,
			$6::uuid, $7, $8
		)`,
		op.StudentID,
		op.Kind,
		op.Direction,
		amount,
		op.Reason,
		opUID,
		op.OperatorName,
		idem,
	)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return repository.ErrBenefitIdempotencyConflict
		}
		return err
	}
	return nil
}

func applyBenefitAdjustToContractInfo(raw []byte, op *domain.BenefitAdjustOp) ([]byte, float64, error) {
	snap := studentcontract.ParseBenefitsContractSnapshot(raw)
	patch := make(map[string]interface{})

	switch op.Kind {
	case domain.BenefitKindFlightTime:
		delta := int(math.Round(op.Amount))
		if delta <= 0 {
			return nil, 0, errors.New("invalid amount")
		}
		total := op.PrecomputedTotalMinutes
		flown := op.PrecomputedFlownMinutes
		adjust := snap.FlightTimeAdjustMinutes
		rem := total - flown + adjust
		if op.Direction == domain.BenefitDirectionCredit {
			adjust += delta
		} else {
			if rem < delta {
				return nil, 0, repository.ErrBenefitDebitExceeds
			}
			adjust -= delta
		}
		patch["flight_time_adjust_minutes"] = adjust
		out, err := studentcontract.MergeContractInfoMap(raw, patch)
		return out, float64(delta), err

	case domain.BenefitKindLearningPoints:
		n := int(math.Round(op.Amount))
		if n <= 0 || float64(n) != math.Round(op.Amount) {
			return nil, 0, errors.New("learning_points amount must be a positive integer")
		}
		lp := snap.LearningPoints
		if op.Direction == domain.BenefitDirectionCredit {
			lp.Total += n
		} else {
			if lp.Remaining < n {
				return nil, 0, repository.ErrBenefitDebitExceeds
			}
			lp.Used += n
		}
		lp.Remaining = lp.Total - lp.Used
		if lp.Remaining < 0 {
			lp.Remaining = 0
		}
		patch["learning_points"] = studentcontract.PatchLearningPointsObject(lp)
		out, err := studentcontract.MergeContractInfoMap(raw, patch)
		return out, float64(n), err

	case domain.BenefitKindGuestExperience:
		n := int(math.Round(op.Amount))
		if n <= 0 || float64(n) != math.Round(op.Amount) {
			return nil, 0, errors.New("guest_experience amount must be a positive integer")
		}
		ge := snap.GuestExperience
		if op.Direction == domain.BenefitDirectionCredit {
			ge.Quota += n
		} else {
			if ge.Remaining < n {
				return nil, 0, repository.ErrBenefitDebitExceeds
			}
			ge.Used += n
		}
		ge.Remaining = ge.Quota - ge.Used
		if ge.Remaining < 0 {
			ge.Remaining = 0
		}
		patch["guest_experience"] = studentcontract.PatchGuestExperienceObject(ge)
		out, err := studentcontract.MergeContractInfoMap(raw, patch)
		return out, float64(n), err

	default:
		return nil, 0, fmt.Errorf("unknown benefit_kind %q", op.Kind)
	}
}
