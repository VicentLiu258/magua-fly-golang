package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
)

// ValidateInstructorBookingWindow 校验预约时段是否完全落在教员当日执勤窗口内；无排班记录则不允许预约。
func (u *ScheduleUseCase) ValidateInstructorBookingWindow(ctx context.Context, instructorID string, planDate time.Time, startTime, endTime time.Time) error {
	day := time.Date(planDate.Year(), planDate.Month(), planDate.Day(), 0, 0, 0, 0, planDate.Location())
	row, err := u.instrScheduleRepo.GetByInstructorAndDate(ctx, instructorID, day)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("教员当日未配置排班，无法预约")
	}
	cfg := defaultScheduleConfig()
	intervals, err := resolveInstructorDutyIntervals(cfg, row.DutyStatus, row.DutyWindowsJSON)
	if err != nil {
		return err
	}
	if len(intervals) == 0 {
		return fmt.Errorf("教员当日为请假/休息或无执勤时段，无法预约")
	}
	if !bookingIntervalFitsDuty(startTime, endTime, intervals) {
		return fmt.Errorf("预约时段不在教员执勤窗口内")
	}
	return nil
}

// ListInstructorDuties 按区间列出教员排班行，并聚合带飞任务（预约 + 训练计划去重）。
func (u *ScheduleUseCase) ListInstructorDuties(ctx context.Context, from, to time.Time, instructorIDs []string) ([]domain.InstructorDutyListItem, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	if toD.Before(fromD) {
		return nil, fmt.Errorf("invalid date range")
	}

	rows, err := u.instrScheduleRepo.ListByDateRange(ctx, fromD, toD, instructorIDs)
	if err != nil {
		return nil, err
	}

	bookings, err := u.bookingRepo.ListByDateRange(ctx, fromD, toD)
	if err != nil {
		return nil, err
	}

	fromStr := fromD.Format("2006-01-02")
	toStr := toD.Format("2006-01-02")
	plans, err := u.planRepo.ListByDate(ctx, fromStr, toStr)
	if err != nil {
		return nil, err
	}

	instructors, err := u.instructorRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	nameByID := make(map[string]string, len(instructors))
	for _, ins := range instructors {
		nameByID[ins.ID] = ins.Name
	}

	cfg := defaultScheduleConfig()
	out := make([]domain.InstructorDutyListItem, 0, len(rows))
	for _, row := range rows {
		intervals, err := resolveInstructorDutyIntervals(cfg, row.DutyStatus, row.DutyWindowsJSON)
		if err != nil {
			return nil, err
		}
		resolved := intervalsToResolved(intervals)
		rawWindows, _ := domain.ParseDutyWindowsJSON(row.DutyWindowsJSON)

		day := time.Date(row.Date.Year(), row.Date.Month(), row.Date.Day(), 0, 0, 0, 0, row.Date.Location())
		flight := aggregateFlightTasks(bookings, plans, row.InstructorID, day)

		out = append(out, domain.InstructorDutyListItem{
			ScheduleID:      row.ID,
			InstructorID:    row.InstructorID,
			InstructorName:  nameByID[row.InstructorID],
			Date:            day.Format("2006-01-02"),
			DutyStatus:      row.DutyStatus,
			Remark:          row.Remark,
			DutyWindows:     rawWindows,
			ResolvedWindows: resolved,
			FlightTasks:     flight,
			GroundCourses:   domain.DutyTaskBundle{Items: []domain.DutyTaskItem{}, TotalMinutes: 0},
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

// ErrInstructorProfileNotFound 当前登录用户未关联 instructors 行。
var ErrInstructorProfileNotFound = errors.New("未绑定教员档案")

// MyInstructorDutiesResult 教员「我的排班与任务」聚合（与接口 data 对齐）。
type MyInstructorDutiesResult struct {
	InstructorID    string                          `json:"instructor_id"`
	InstructorName  string                          `json:"instructor_name"`
	Items           []domain.InstructorDutyListItem `json:"items"`
	PendingBookings []*domain.Booking               `json:"pending_bookings"`
}

// ListMyInstructorDuties 按 JWT 对应教员列出区间内排班与带飞任务，并附带**全部**待审核预约（不按日期过滤）。
func (u *ScheduleUseCase) ListMyInstructorDuties(ctx context.Context, userID string, from, to time.Time) (*MyInstructorDutiesResult, error) {
	ins, err := u.instructorRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if ins == nil {
		return nil, ErrInstructorProfileNotFound
	}
	items, err := u.ListInstructorDuties(ctx, from, to, []string{ins.ID})
	if err != nil {
		return nil, err
	}
	allPending, err := u.bookingRepo.ListPending(ctx)
	if err != nil {
		return nil, err
	}
	var pending []*domain.Booking
	for _, b := range allPending {
		if b.InstructorID == ins.ID {
			pending = append(pending, b)
		}
	}
	return &MyInstructorDutiesResult{
		InstructorID:    ins.ID,
		InstructorName:  ins.Name,
		Items:           items,
		PendingBookings: pending,
	}, nil
}

func intervalsToResolved(intervals []dutyInterval) []domain.ResolvedDutyWindow {
	out := make([]domain.ResolvedDutyWindow, 0, len(intervals))
	for _, inv := range intervals {
		out = append(out, domain.ResolvedDutyWindow{
			Start: formatMinuteClock(inv.startMin),
			End:   formatMinuteClock(inv.endMin),
		})
	}
	return out
}

func formatMinuteClock(min int) string {
	return fmt.Sprintf("%02d:%02d", min/60, min%60)
}

func aggregateFlightTasks(bookings []*domain.Booking, plans []*domain.TrainingPlan, instructorID string, day time.Time) domain.DutyTaskBundle {
	type key struct {
		studentID string
		startMin  int
		endMin    int
	}
	seen := make(map[key]bool)
	var items []domain.DutyTaskItem
	total := 0

	for _, b := range bookings {
		if !strings.EqualFold(b.InstructorID, instructorID) {
			continue
		}
		bd := b.PlanDate
		if bd.Year() != day.Year() || bd.Month() != day.Month() || bd.Day() != day.Day() {
			continue
		}
		if b.Status == domain.BookingRejected {
			continue
		}
		sm := b.StartTime.Hour()*60 + b.StartTime.Minute()
		em := b.EndTime.Hour()*60 + b.EndTime.Minute()
		if em <= sm {
			continue
		}
		k := key{b.StudentID, sm, em}
		if seen[k] {
			continue
		}
		seen[k] = true
		minutes := em - sm
		total += minutes
		items = append(items, domain.DutyTaskItem{
			Source:        "booking",
			RefID:         b.ID,
			StartTime:     b.StartTime.Format("15:04"),
			EndTime:       b.EndTime.Format("15:04"),
			Minutes:       minutes,
			StudentID:     b.StudentID,
			AircraftID:    b.AircraftID,
			BookingStatus: string(b.Status),
		})
	}

	for _, p := range plans {
		if !strings.EqualFold(p.InstructorID, instructorID) {
			continue
		}
		pd := p.PlanDate
		if pd.Year() != day.Year() || pd.Month() != day.Month() || pd.Day() != day.Day() {
			continue
		}
		sm := p.StartTime.Hour()*60 + p.StartTime.Minute()
		em := p.EndTime.Hour()*60 + p.EndTime.Minute()
		if em <= sm {
			continue
		}
		k := key{p.StudentID, sm, em}
		if seen[k] {
			continue
		}
		seen[k] = true
		minutes := em - sm
		total += minutes
		items = append(items, domain.DutyTaskItem{
			Source:     "training_plan",
			RefID:      p.ID,
			StartTime:  p.StartTime.Format("15:04"),
			EndTime:    p.EndTime.Format("15:04"),
			Minutes:    minutes,
			StudentID:  p.StudentID,
			AircraftID: p.AircraftID,
			Title:      string(p.Source),
		})
	}

	return domain.DutyTaskBundle{Items: items, TotalMinutes: total}
}

// UpsertInstructorDuty 单日写入或更新排班。
func (u *ScheduleUseCase) UpsertInstructorDuty(ctx context.Context, s *domain.InstructorSchedule) error {
	if s == nil {
		return fmt.Errorf("nil schedule")
	}
	s.Date = time.Date(s.Date.Year(), s.Date.Month(), s.Date.Day(), 0, 0, 0, 0, s.Date.Location())
	if err := validateDutyPayload(s); err != nil {
		return err
	}
	// 预解析，避免非法 JSON / 非法状态入库后无法展示
	cfg := defaultScheduleConfig()
	if _, err := resolveInstructorDutyIntervals(cfg, s.DutyStatus, s.DutyWindowsJSON); err != nil {
		return err
	}
	return u.instrScheduleRepo.UpsertDay(ctx, s)
}

func validateDutyPayload(s *domain.InstructorSchedule) error {
	switch s.DutyStatus {
	case domain.InstructorDutyMorning, domain.InstructorDutyAfternoon, domain.InstructorDutyFullDay,
		domain.InstructorDutyLeave, domain.InstructorDutyRest, domain.InstructorDutyCustom:
	default:
		return fmt.Errorf("invalid duty_status")
	}
	if s.DutyStatus == domain.InstructorDutyCustom {
		if len(s.DutyWindowsJSON) == 0 {
			return fmt.Errorf("duty_status=custom 时必须提供 duty_windows")
		}
	}
	return nil
}

// BulkInstructorDuties 按日期区间与周几批量写入。weekdays 使用 Go 约定：0=周日 … 6=周六。
func (u *ScheduleUseCase) BulkInstructorDuties(ctx context.Context, instructorID string, from, to time.Time, weekdays []int, dutyStatus domain.InstructorDutyStatus, remark string, replaceExisting bool) (int, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	if toD.Before(fromD) {
		return 0, fmt.Errorf("invalid date range")
	}
	wdSet := make(map[int]struct{}, len(weekdays))
	for _, w := range weekdays {
		if w < 0 || w > 6 {
			return 0, fmt.Errorf("weekday must be 0-6 (Sun-Sat)")
		}
		wdSet[w] = struct{}{}
	}
	if len(wdSet) == 0 {
		return 0, fmt.Errorf("weekdays must not be empty")
	}

	s := &domain.InstructorSchedule{
		InstructorID:    instructorID,
		DutyStatus:      dutyStatus,
		Remark:          remark,
		DutyWindowsJSON: nil,
	}
	if err := validateDutyPayload(s); err != nil {
		return 0, err
	}
	cfg := defaultScheduleConfig()
	if _, err := resolveInstructorDutyIntervals(cfg, dutyStatus, nil); err != nil {
		return 0, err
	}

	n := 0
	for d := fromD; !d.After(toD); d = d.AddDate(0, 0, 1) {
		wd := int(d.Weekday())
		if _, ok := wdSet[wd]; !ok {
			continue
		}
		if !replaceExisting {
			existing, err := u.instrScheduleRepo.GetByInstructorAndDate(ctx, instructorID, d)
			if err != nil {
				return n, err
			}
			if existing != nil {
				continue
			}
		}
		s.Date = d
		s.Remark = remark
		if err := u.instrScheduleRepo.UpsertDay(ctx, s); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
