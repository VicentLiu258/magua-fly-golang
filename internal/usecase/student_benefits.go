package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/pkg/studentcontract"
	"magua-fly-golang/internal/repository"
)

// StudentBenefitsUseCase 后台学员权益概览与调整。
type StudentBenefitsUseCase struct {
	students  repository.StudentRepository
	contracts repository.ContractRepository
	ledger    repository.StudentBenefitLedgerRepository
	adjuster  repository.StudentBenefitsAdjuster
}

// NewStudentBenefitsUseCase constructs StudentBenefitsUseCase.
func NewStudentBenefitsUseCase(
	students repository.StudentRepository,
	contracts repository.ContractRepository,
	ledger repository.StudentBenefitLedgerRepository,
	adjuster repository.StudentBenefitsAdjuster,
) *StudentBenefitsUseCase {
	return &StudentBenefitsUseCase{
		students:  students,
		contracts: contracts,
		ledger:    ledger,
		adjuster:  adjuster,
	}
}

func benefitsSumContractMinutesNonCancelled(contracts []*domain.Contract) int {
	n := 0
	for _, c := range contracts {
		if strings.EqualFold(strings.TrimSpace(c.Status), "cancelled") {
			continue
		}
		if c.TotalMinutes != nil {
			n += *c.TotalMinutes
		}
	}
	return n
}

func totalMinutesForStudent(s *domain.Student, contracts []*domain.Contract) int {
	paid := benefitsSumContractMinutesNonCancelled(contracts)
	total := paid
	if s.ContractDurationMinutes != nil && *s.ContractDurationMinutes > 0 {
		total = *s.ContractDurationMinutes
	}
	return total
}

func hasActiveNonCancelledContract(contracts []*domain.Contract) bool {
	for _, c := range contracts {
		if !strings.EqualFold(strings.TrimSpace(c.Status), "cancelled") {
			return true
		}
	}
	return false
}

func maxTime(ts ...time.Time) time.Time {
	var m time.Time
	for _, t := range ts {
		if t.After(m) {
			m = t
		}
	}
	return m
}

func remainingFlightMinutes(total, flown, adjust int) int {
	r := total - flown + adjust
	if r < 0 {
		return 0
	}
	return r
}

// BenefitsOverviewData 与 /admin/students/benefits/overview 文档一致。
type BenefitsOverviewData struct {
	StudentID          string                        `json:"student_id"`
	ComputedAt         time.Time                     `json:"computed_at"`
	FlightTime         BenefitsFlightTimeBlock       `json:"flight_time"`
	GuestExperience    BenefitsCountBlock            `json:"guest_experience"`
	LearningPoints     BenefitsPointsBlock           `json:"learning_points"`
	BookingLimits      studentcontract.BookingLimits `json:"booking_limits"`
	Extras             studentcontract.Extras        `json:"extras"`
	HasActiveContract  bool                          `json:"has_active_contract"`
}

// BenefitsFlightTimeBlock 飞行时间（小时，1 位小数）。
type BenefitsFlightTimeBlock struct {
	Unit      string  `json:"unit"`
	Total     float64 `json:"total"`
	Used      float64 `json:"used"`
	Remaining float64 `json:"remaining"`
}

// BenefitsCountBlock 次数类权益。
type BenefitsCountBlock struct {
	Unit      string  `json:"unit"`
	Total     float64 `json:"total"`
	Used      float64 `json:"used"`
	Remaining float64 `json:"remaining"`
}

// BenefitsPointsBlock 积分。
type BenefitsPointsBlock struct {
	Unit      string  `json:"unit"`
	Total     float64 `json:"total"`
	Used      float64 `json:"used"`
	Remaining float64 `json:"remaining"`
}

// Overview 聚合权益 Tab 数据。
func (u *StudentBenefitsUseCase) Overview(ctx context.Context, studentID string) (*BenefitsOverviewData, error) {
	s, err := u.students.GetByID(ctx, studentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	contracts, err := u.contracts.ListByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	totalMin := totalMinutesForStudent(s, contracts)
	flownMin := s.FlownMinutes
	snap := studentcontract.ParseBenefitsContractSnapshot(s.ContractInfo)
	remMin := remainingFlightMinutes(totalMin, flownMin, snap.FlightTimeAdjustMinutes)

	var ledgerMax *time.Time
	if u.ledger != nil {
		var err error
		ledgerMax, err = u.ledger.MaxCreatedAt(ctx, studentID)
		if err != nil {
			return nil, err
		}
	}
	computed := time.Now()
	computed = maxTime(computed, s.UpdatedAt)
	for _, c := range contracts {
		if c.UpdatedAt.After(computed) {
			computed = c.UpdatedAt
		}
	}
	if ledgerMax != nil && ledgerMax.After(computed) {
		computed = *ledgerMax
	}

	out := &BenefitsOverviewData{
		StudentID:         s.ID,
		ComputedAt:        computed,
		HasActiveContract: hasActiveNonCancelledContract(contracts),
		BookingLimits:     snap.BookingLimits,
		Extras:            snap.Extras,
		FlightTime: BenefitsFlightTimeBlock{
			Unit:      "hour",
			Total:     studentcontract.RoundHours1(totalMin),
			Used:      studentcontract.RoundHours1(flownMin),
			Remaining: studentcontract.RoundHours1(remMin),
		},
		GuestExperience: BenefitsCountBlock{
			Unit:      "times",
			Total:     float64(snap.GuestExperience.Quota),
			Used:      float64(snap.GuestExperience.Used),
			Remaining: float64(snap.GuestExperience.Remaining),
		},
		LearningPoints: BenefitsPointsBlock{
			Unit:      "points",
			Total:     float64(snap.LearningPoints.Total),
			Used:      float64(snap.LearningPoints.Used),
			Remaining: float64(snap.LearningPoints.Remaining),
		},
	}
	return out, nil
}

// BenefitAdjustRequest 调整请求体。
type BenefitAdjustRequest struct {
	StudentID      string
	BenefitKind    string
	Direction      string
	Amount         float64
	Reason         string
	IdempotencyKey string
}

// BenefitAdjustResult 成功时返回该类权益快照（字段名与 overview 子块一致）。
type BenefitAdjustResult struct {
	BenefitKind       string                  `json:"benefit_kind"`
	FlightTime        *BenefitsFlightTimeBlock `json:"flight_time,omitempty"`
	GuestExperience   *BenefitsCountBlock      `json:"guest_experience,omitempty"`
	LearningPoints    *BenefitsPointsBlock     `json:"learning_points,omitempty"`
}

// Adjust 手工增减权益；operatorUserID/operatorName 来自 JWT 与 users 表。
func (u *StudentBenefitsUseCase) Adjust(ctx context.Context, req *BenefitAdjustRequest, operatorUserID *string, operatorName string) (*BenefitAdjustResult, error) {
	if req == nil {
		return nil, ErrInvalidArgument
	}
	req.BenefitKind = strings.TrimSpace(req.BenefitKind)
	req.Direction = strings.TrimSpace(req.Direction)
	if req.StudentID == "" || req.BenefitKind == "" || req.Direction == "" {
		return nil, ErrInvalidArgument
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidArgument
	}
	if req.Direction != domain.BenefitDirectionCredit && req.Direction != domain.BenefitDirectionDebit {
		return nil, ErrInvalidArgument
	}
	switch req.BenefitKind {
	case domain.BenefitKindFlightTime, domain.BenefitKindGuestExperience, domain.BenefitKindLearningPoints:
	default:
		return nil, ErrInvalidArgument
	}

	s, err := u.students.GetByID(ctx, req.StudentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.IdempotencyKey != "" && u.ledger != nil {
		if row, err := u.ledger.GetByIdempotencyKey(ctx, req.IdempotencyKey); err != nil {
			return nil, err
		} else if row != nil {
			return u.buildAdjustResult(ctx, req.StudentID, req.BenefitKind)
		}
	}

	contracts, err := u.contracts.ListByStudent(ctx, req.StudentID)
	if err != nil {
		return nil, err
	}
	totalMin := totalMinutesForStudent(s, contracts)
	flownMin := s.FlownMinutes

	op := &domain.BenefitAdjustOp{
		StudentID:               req.StudentID,
		Kind:                    req.BenefitKind,
		Direction:               req.Direction,
		Amount:                  req.Amount,
		Reason:                  req.Reason,
		IdempotencyKey:          req.IdempotencyKey,
		OperatorUserID:          operatorUserID,
		OperatorName:            operatorName,
		PrecomputedTotalMinutes: totalMin,
		PrecomputedFlownMinutes: flownMin,
	}

	if err := u.adjuster.Adjust(ctx, op); err != nil {
		if errors.Is(err, repository.ErrBenefitDebitExceeds) {
			return nil, ErrBenefitAdjustExceeds
		}
		if errors.Is(err, repository.ErrBenefitIdempotencyConflict) {
			return nil, repository.ErrBenefitIdempotencyConflict
		}
		return nil, err
	}
	return u.buildAdjustResult(ctx, req.StudentID, req.BenefitKind)
}

func (u *StudentBenefitsUseCase) buildAdjustResult(ctx context.Context, studentID, kind string) (*BenefitAdjustResult, error) {
	ov, err := u.Overview(ctx, studentID)
	if err != nil {
		return nil, err
	}
	out := &BenefitAdjustResult{BenefitKind: kind}
	switch kind {
	case domain.BenefitKindFlightTime:
		out.FlightTime = &ov.FlightTime
	case domain.BenefitKindGuestExperience:
		out.GuestExperience = &ov.GuestExperience
	case domain.BenefitKindLearningPoints:
		out.LearningPoints = &ov.LearningPoints
	default:
		return nil, ErrInvalidArgument
	}
	return out, nil
}

// BenefitLogItem 变更历史一行。
type BenefitLogItem struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Direction string    `json:"direction"`
	Amount    float64   `json:"amount"`
	Reason    string    `json:"reason"`
	Operator  string    `json:"operator"`
	CreatedAt time.Time `json:"created_at"`
}

// BenefitLogsData 分页结果。
type BenefitLogsData struct {
	Items    []BenefitLogItem `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// Logs 分页查询流水。
func (u *StudentBenefitsUseCase) Logs(ctx context.Context, studentID, benefitKind string, page, pageSize int) (*BenefitLogsData, error) {
	if _, err := u.students.GetByID(ctx, studentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	rows, total, err := u.ledger.ListByStudentKind(ctx, studentID, benefitKind, pageSize, offset)
	if err != nil {
		return nil, err
	}
	items := make([]BenefitLogItem, 0, len(rows))
	for _, r := range rows {
		loc := time.Local
		t := r.CreatedAt.In(loc)
		items = append(items, BenefitLogItem{
			ID:        r.ID,
			Date:      t.Format("2006-01-02"),
			Direction: r.Direction,
			Amount:    r.Amount,
			Reason:    r.Reason,
			Operator:  r.OperatorName,
			CreatedAt: r.CreatedAt,
		})
	}
	return &BenefitLogsData{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// SaveExtrasRequest 仅更新 booking_limits / extras。
type SaveExtrasRequest struct {
	StudentID      string
	BookingLimits  *studentcontract.BookingLimits
	Extras         *studentcontract.Extras
}

// SaveExtras 合并写入 contract_info，不改动额度与流水。
func (u *StudentBenefitsUseCase) SaveExtras(ctx context.Context, req *SaveExtrasRequest) error {
	if req == nil || req.StudentID == "" {
		return ErrInvalidArgument
	}
	s, err := u.students.GetByID(ctx, req.StudentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	patch := make(map[string]interface{})
	if req.BookingLimits != nil {
		bl := *req.BookingLimits
		studentcontract.NormalizeBookingLimitsForSave(&bl)
		patch["booking_limits"] = bl
	}
	if req.Extras != nil {
		patch["extras"] = *req.Extras
	}
	if len(patch) == 0 {
		return nil
	}
	newCI, err := studentcontract.MergeContractInfoMap(s.ContractInfo, patch)
	if err != nil {
		return ErrInvalidContractInfo
	}
	s.ContractInfo = newCI
	return u.students.Update(ctx, s)
}
