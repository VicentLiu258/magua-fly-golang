package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// StudentTrainingSummaryUseCase 学员训练信息汇总（读）与概要字段写入。
type StudentTrainingSummaryUseCase struct {
	students    repository.StudentRepository
	contracts   repository.ContractRepository
	records     repository.TrainingRecordRepository
	instructors repository.InstructorRepository
}

// NewStudentTrainingSummaryUseCase constructs StudentTrainingSummaryUseCase.
func NewStudentTrainingSummaryUseCase(
	students repository.StudentRepository,
	contracts repository.ContractRepository,
	records repository.TrainingRecordRepository,
	instructors repository.InstructorRepository,
) *StudentTrainingSummaryUseCase {
	return &StudentTrainingSummaryUseCase{
		students:    students,
		contracts:   contracts,
		records:     records,
		instructors: instructors,
	}
}

// StudentTrainingSummary 学员训练信息汇总。
type StudentTrainingSummary struct {
	StudentID               string `json:"student_id"`
	Name                    string `json:"name"`
	StudentNo               string `json:"student_no"`
	FlightTrainingStatus    string `json:"flight_training_status"`
	TrainingCourse          string `json:"training_course,omitempty"`
	ResponsibleInstructorID *string `json:"responsible_instructor_id,omitempty"`
	ResponsibleInstructor   *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"responsible_instructor,omitempty"`
	HasNightFlight    bool `json:"has_night_flight"`
	HasCrossCountry   bool `json:"has_cross_country"`
	Duration          StudentTrainingDurationSummary `json:"duration"`
	Flights           StudentFlightCountSummary      `json:"flights"`
}

// StudentTrainingDurationSummary 时长统计（分钟）。
type StudentTrainingDurationSummary struct {
	// TotalMinutes 总时间：优先 `students.contract_duration_minutes`；若未设则为未取消合同的 `total_minutes` 合计。
	TotalMinutes int `json:"total_minutes"`
	// PaidMinutes 已购/已付款时长：状态非 `cancelled` 的合同的 `total_minutes` 合计。
	PaidMinutes int `json:"paid_minutes"`
	// FlownMinutes 已飞时间：由 `training_records` 中 `status=normal` 的 `flight_duration_minutes` 汇总。
	FlownMinutes int `json:"flown_minutes"`
	// LedgerFlownMinutes 台账已飞：`students.flown_minutes`（可能与记录汇总不一致，以业务录入为准）。
	LedgerFlownMinutes int `json:"ledger_flown_minutes"`
}

// StudentFlightCountSummary 架次统计。
type StudentFlightCountSummary struct {
	// NormalRecordCount 有效训练记录条数（`status=normal`）。
	NormalRecordCount int `json:"normal_record_count"`
}

func sumContractMinutesNonCancelled(contracts []*domain.Contract) int {
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

// Build 汇总指定学员的训练信息。
func (u *StudentTrainingSummaryUseCase) Build(ctx context.Context, studentID string) (*StudentTrainingSummary, error) {
	s, err := u.students.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return u.buildFromStudent(ctx, s)
}

func (u *StudentTrainingSummaryUseCase) buildFromStudent(ctx context.Context, s *domain.Student) (*StudentTrainingSummary, error) {
	out := &StudentTrainingSummary{
		StudentID:               s.ID,
		Name:                    s.Name,
		StudentNo:               s.StudentNo,
		FlightTrainingStatus:    string(s.FlightTrainingStatus),
		TrainingCourse:          s.TrainingCourse,
		ResponsibleInstructorID: s.ResponsibleInstructorID,
	}
	if s.ResponsibleInstructorID != nil && strings.TrimSpace(*s.ResponsibleInstructorID) != "" {
		if ins, err := u.instructors.GetByID(ctx, *s.ResponsibleInstructorID); err == nil && ins != nil {
			out.ResponsibleInstructor = &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{ID: ins.ID, Name: ins.Name}
		}
	}
	flownAgg, nCount, night, xc, err := u.records.AggregateStudentStats(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	out.HasNightFlight = night
	out.HasCrossCountry = xc
	out.Flights.NormalRecordCount = nCount

	contracts, err := u.contracts.ListByStudent(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	paidSum := sumContractMinutesNonCancelled(contracts)
	totalMin := paidSum
	if s.ContractDurationMinutes != nil && *s.ContractDurationMinutes > 0 {
		totalMin = *s.ContractDurationMinutes
	}
	out.Duration = StudentTrainingDurationSummary{
		TotalMinutes:       totalMin,
		PaidMinutes:        paidSum,
		FlownMinutes:       flownAgg,
		LedgerFlownMinutes: s.FlownMinutes,
	}
	return out, nil
}

// SaveStudentTrainingProfileRequest 更新学员训练概要字段（nil 表示不修改）。
type SaveStudentTrainingProfileRequest struct {
	TrainingCourse          *string `json:"training_course"`
	FlightTrainingStatus    *string `json:"flight_training_status"`
	ResponsibleInstructorID *string `json:"responsible_instructor_id"`
}

// SaveProfile 写入训练概要字段（管理员）。
func (u *StudentTrainingSummaryUseCase) SaveProfile(ctx context.Context, studentID string, req *SaveStudentTrainingProfileRequest) (*domain.Student, error) {
	if req == nil {
		return nil, ErrInvalidArgument
	}
	s, err := u.students.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if req.TrainingCourse != nil {
		t := strings.TrimSpace(*req.TrainingCourse)
		if t == "" {
			s.TrainingCourse = ""
		} else if _, ok := domain.ParseTrainingCourse(t); ok {
			s.TrainingCourse = strings.ToUpper(t)
		} else {
			return nil, ErrInvalidArgument
		}
	}
	if req.FlightTrainingStatus != nil {
		t := strings.TrimSpace(*req.FlightTrainingStatus)
		if t == "" {
			return nil, ErrInvalidArgument
		}
		if fs, ok := domain.ParseFlightTrainingStatus(t); ok {
			s.FlightTrainingStatus = fs
		} else {
			return nil, ErrInvalidArgument
		}
	}
	if req.ResponsibleInstructorID != nil {
		rid := strings.TrimSpace(*req.ResponsibleInstructorID)
		if rid == "" {
			s.ResponsibleInstructorID = nil
		} else {
			if _, err := u.instructors.GetByID(ctx, rid); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, ErrNotFound
				}
				return nil, err
			}
			s.ResponsibleInstructorID = &rid
		}
	}
	if err := u.students.Update(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

// ClearTrainingCourse 清空训练课程阶段（管理员）。
func (u *StudentTrainingSummaryUseCase) ClearTrainingCourse(ctx context.Context, studentID string) error {
	s, err := u.students.GetByID(ctx, studentID)
	if err != nil {
		return err
	}
	s.TrainingCourse = ""
	return u.students.Update(ctx, s)
}
