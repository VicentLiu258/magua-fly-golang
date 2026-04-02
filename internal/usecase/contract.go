package usecase

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

const (
	maxInstallmentPeriods = 24
	paymentSumEpsilon     = 0.01
)

// NormalizePaymentMode 空值视为 full；否则返回小写去空格后的字符串。
func NormalizePaymentMode(mode string) string {
	m := strings.ToLower(strings.TrimSpace(mode))
	if m == "" {
		return domain.PaymentModeFull
	}
	return m
}

// ValidateContractPayment 校验全款/分期与付款计划（分期时各期金额之和须与 price_total 一致，允许 0.01 误差）。
func ValidateContractPayment(priceTotal *float64, mode string, schedule []domain.PaymentInstallment) error {
	mode = NormalizePaymentMode(mode)
	if mode != domain.PaymentModeFull && mode != domain.PaymentModeInstallment {
		return ErrInvalidPaymentMode
	}
	if mode == domain.PaymentModeFull {
		if len(schedule) > 0 {
			return ErrInvalidPaymentSchedule
		}
		return nil
	}
	if priceTotal == nil {
		return ErrPriceTotalRequired
	}
	if len(schedule) < 2 || len(schedule) > maxInstallmentPeriods {
		return ErrInvalidPaymentSchedule
	}
	var sum float64
	for _, p := range schedule {
		if p.Amount <= 0 {
			return ErrInvalidPaymentSchedule
		}
		if strings.TrimSpace(p.DueDate) == "" {
			return ErrInvalidPaymentSchedule
		}
		if _, err := time.Parse("2006-01-02", p.DueDate); err != nil {
			return ErrInvalidPaymentSchedule
		}
		sum += p.Amount
	}
	if math.Abs(sum-*priceTotal) > paymentSumEpsilon {
		return ErrPaymentScheduleSumMismatch
	}
	return nil
}

// ContractUseCase 后台合同创建/更新与付款计划校验。
type ContractUseCase struct {
	contractRepo repository.ContractRepository
	studentRepo  repository.StudentRepository
}

// NewContractUseCase 创建 ContractUseCase。
func NewContractUseCase(contractRepo repository.ContractRepository, studentRepo repository.StudentRepository) *ContractUseCase {
	return &ContractUseCase{
		contractRepo: contractRepo,
		studentRepo:  studentRepo,
	}
}

// ContractCreateInput 创建合同入参。
type ContractCreateInput struct {
	StudentID       string
	ContractNo      string
	SignDate        *time.Time
	StartDate       *time.Time
	EndDate         *time.Time
	TotalMinutes    *int
	PriceTotal      *float64
	Currency        string
	Status          string
	Remark          string
	PaymentMode     string
	PaymentSchedule []domain.PaymentInstallment
}

// ContractUpdateInput 更新合同入参（指针字段为 nil 表示不修改）。
type ContractUpdateInput struct {
	ID              string
	ContractNo      *string
	SignDate        *time.Time
	StartDate       *time.Time
	EndDate         *time.Time
	TotalMinutes    *int
	PriceTotal      *float64
	Currency        *string
	Status          *string
	Remark          *string
	PaymentMode     *string
	PaymentSchedule *[]domain.PaymentInstallment
}

// Create 创建合同。
func (uc *ContractUseCase) Create(ctx context.Context, in *ContractCreateInput) (*domain.Contract, error) {
	if uc == nil || in == nil || strings.TrimSpace(in.StudentID) == "" {
		return nil, ErrInvalidArgument
	}
	_, err := uc.studentRepo.GetByID(ctx, in.StudentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	mode := NormalizePaymentMode(in.PaymentMode)
	if mode != domain.PaymentModeFull && mode != domain.PaymentModeInstallment {
		return nil, ErrInvalidPaymentMode
	}
	sched := in.PaymentSchedule
	if mode == domain.PaymentModeFull {
		sched = nil
	}
	if err := ValidateContractPayment(in.PriceTotal, mode, sched); err != nil {
		return nil, err
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	c := &domain.Contract{
		StudentID:       in.StudentID,
		ContractNo:      in.ContractNo,
		SignDate:        in.SignDate,
		StartDate:       in.StartDate,
		EndDate:         in.EndDate,
		TotalMinutes:    in.TotalMinutes,
		PriceTotal:      in.PriceTotal,
		Currency:        in.Currency,
		Status:          status,
		Remark:          in.Remark,
		PaymentMode:     mode,
		PaymentSchedule: sched,
		MongoFileIDs:    nil,
	}
	if mode == domain.PaymentModeFull {
		c.PaymentSchedule = []domain.PaymentInstallment{}
	}
	if err := uc.contractRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return uc.contractRepo.GetByID(ctx, c.ID)
}

// Update 更新合同。
func (uc *ContractUseCase) Update(ctx context.Context, in *ContractUpdateInput) (*domain.Contract, error) {
	if uc == nil || in == nil || strings.TrimSpace(in.ID) == "" {
		return nil, ErrInvalidArgument
	}
	existing, err := uc.contractRepo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if in.ContractNo != nil {
		existing.ContractNo = *in.ContractNo
	}
	if in.SignDate != nil {
		existing.SignDate = in.SignDate
	}
	if in.StartDate != nil {
		existing.StartDate = in.StartDate
	}
	if in.EndDate != nil {
		existing.EndDate = in.EndDate
	}
	if in.TotalMinutes != nil {
		existing.TotalMinutes = in.TotalMinutes
	}
	if in.PriceTotal != nil {
		existing.PriceTotal = in.PriceTotal
	}
	if in.Currency != nil {
		existing.Currency = *in.Currency
	}
	if in.Status != nil {
		existing.Status = *in.Status
	}
	if in.Remark != nil {
		existing.Remark = *in.Remark
	}
	if in.PaymentMode != nil {
		m := NormalizePaymentMode(*in.PaymentMode)
		if m != domain.PaymentModeFull && m != domain.PaymentModeInstallment {
			return nil, ErrInvalidPaymentMode
		}
		existing.PaymentMode = m
	}
	if in.PaymentSchedule != nil {
		existing.PaymentSchedule = *in.PaymentSchedule
	}
	if existing.PaymentMode == "" {
		existing.PaymentMode = domain.PaymentModeFull
	}
	if existing.PaymentMode == domain.PaymentModeFull {
		existing.PaymentSchedule = []domain.PaymentInstallment{}
	}
	sched := existing.PaymentSchedule
	if existing.PaymentMode == domain.PaymentModeFull {
		sched = nil
	}
	if err := ValidateContractPayment(existing.PriceTotal, existing.PaymentMode, sched); err != nil {
		return nil, err
	}
	if err := uc.contractRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return uc.contractRepo.GetByID(ctx, existing.ID)
}

// GetByID 按 ID 读取合同。
func (uc *ContractUseCase) GetByID(ctx context.Context, id string) (*domain.Contract, error) {
	if uc == nil || strings.TrimSpace(id) == "" {
		return nil, ErrInvalidArgument
	}
	c, err := uc.contractRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return c, nil
}
