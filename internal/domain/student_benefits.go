package domain

import "time"

// 权益流水种类（student_benefit_ledger.benefit_kind）。
const (
	BenefitKindFlightTime       = "flight_time"
	BenefitKindGuestExperience  = "guest_experience"
	BenefitKindLearningPoints   = "learning_points"
)

// 调整方向。
const (
	BenefitDirectionCredit = "credit"
	BenefitDirectionDebit  = "debit"
)

// StudentBenefitLedger 学员权益手工调整流水。
type StudentBenefitLedger struct {
	ID               string
	StudentID        string
	BenefitKind      string
	Direction        string
	Amount           float64
	Reason           string
	OperatorUserID   *string
	OperatorName     string
	CreatedAt        time.Time
	IdempotencyKey   string
}

// BenefitAdjustOp 手工增减权益（飞行时间为分钟；亲友/积分为整数额度）。
type BenefitAdjustOp struct {
	StudentID      string
	Kind           string
	Direction      string
	Amount         float64
	Reason         string
	IdempotencyKey string
	OperatorUserID *string
	OperatorName   string
	// PrecomputedTotalMinutes / PrecomputedFlownMinutes 仅 Kind=flight_time 时使用（与训练汇总 total / 台账 flown 一致）。
	PrecomputedTotalMinutes int
	PrecomputedFlownMinutes int
}
