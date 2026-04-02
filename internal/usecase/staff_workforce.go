package usecase

import (
	"context"
	"fmt"
	"time"

	"magua-fly-golang/internal/domain"
)

// GetWorkforceGrid 工作人员考勤网格：role 为 staff + maintenance + instructor（不含 admin/student），仅执勤可用格，无预约层。
func (u *ScheduleUseCase) GetWorkforceGrid(ctx context.Context, date time.Time) (*domain.WorkforceGridResponse, error) {
	if u.userRepo == nil || u.staffScheduleRepo == nil {
		return nil, fmt.Errorf("workforce grid not configured")
	}
	cfg := defaultScheduleConfig()
	slotCount, slots, slotStarts, err := prepareSlotTimeline(cfg)
	if err != nil {
		return nil, err
	}
	dayDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	users, err := u.userRepo.ListByRoles(ctx, []domain.UserRole{domain.RoleStaff, domain.RoleMaintenance, domain.RoleInstructor})
	if err != nil {
		return nil, err
	}

	instructorSchedules, err := u.instrScheduleRepo.ListByDate(ctx, dayDate)
	if err != nil {
		return nil, err
	}
	staffSchedules, err := u.staffScheduleRepo.ListByDate(ctx, dayDate)
	if err != nil {
		return nil, err
	}

	schedByInstructor := make(map[string][]*domain.InstructorSchedule)
	for _, s := range instructorSchedules {
		schedByInstructor[s.InstructorID] = append(schedByInstructor[s.InstructorID], s)
	}
	schedByStaffUser := make(map[string][]*domain.StaffSchedule)
	for _, s := range staffSchedules {
		schedByStaffUser[s.UserID] = append(schedByStaffUser[s.UserID], s)
	}

	result := &domain.WorkforceGridResponse{
		Date:        dayDate.Format("2006-01-02"),
		SlotMinutes: cfg.slotMinutes,
		DayStart:    fmt.Sprintf("%02d:%02d", int(cfg.dayStart.Hours()), int(cfg.dayStart.Minutes())%60),
		DayEnd:      fmt.Sprintf("%02d:%02d", int(cfg.dayEnd.Hours()), int(cfg.dayEnd.Minutes())%60),
		Slots:       slots,
	}

	for _, usr := range users {
		timeline := make([]domain.TimeSlot, slotCount)
		for i := 0; i < slotCount; i++ {
			start := slotStarts[i]
			end := start + time.Duration(cfg.slotMinutes)*time.Minute
			timeline[i] = domain.TimeSlot{
				Index:     i,
				StartTime: fmt.Sprintf("%02d:%02d", int(start.Hours()), int(start.Minutes())%60),
				EndTime:   fmt.Sprintf("%02d:%02d", int(end.Hours()), int(end.Minutes())%60),
				Status:    domain.SlotOff,
			}
		}

		row := domain.WorkforceMemberRow{
			UserID:   usr.ID,
			Role:     usr.Role,
			Username: usr.Username,
			Timeline: timeline,
		}

		switch usr.Role {
		case domain.RoleInstructor:
			ins, err := u.instructorRepo.GetByUserID(ctx, usr.ID)
			if err != nil {
				return nil, err
			}
			if ins != nil {
				row.Name = ins.Name
				row.DisplayColor = ins.DisplayColor
				iid := ins.ID
				row.InstructorID = &iid
				if list, ok := schedByInstructor[ins.ID]; ok {
					for _, s := range list {
						intervals, err := resolveInstructorDutyIntervals(cfg, s.DutyStatus, s.DutyWindowsJSON)
						if err != nil {
							return nil, err
						}
						for _, inv := range intervals {
							for i := 0; i < slotCount; i++ {
								sm := int(slotStarts[i].Minutes())
								em := sm + cfg.slotMinutes
								if slotOverlapsDuty(sm, em, inv) && timeline[i].Status == domain.SlotOff {
									timeline[i].Status = domain.SlotAvailable
								}
							}
						}
					}
				}
			} else {
				row.Name = usr.Username
			}
		case domain.RoleStaff, domain.RoleMaintenance:
			row.Name = usr.Username
			if list, ok := schedByStaffUser[usr.ID]; ok {
				for _, s := range list {
					intervals, err := resolveInstructorDutyIntervals(cfg, s.DutyStatus, s.DutyWindowsJSON)
					if err != nil {
						return nil, err
					}
					for _, inv := range intervals {
						for i := 0; i < slotCount; i++ {
							sm := int(slotStarts[i].Minutes())
							em := sm + cfg.slotMinutes
							if slotOverlapsDuty(sm, em, inv) && timeline[i].Status == domain.SlotOff {
								timeline[i].Status = domain.SlotAvailable
							}
						}
					}
				}
			}
		default:
			continue
		}

		row.Phone = usr.Phone
		row.Email = usr.Email
		result.Workers = append(result.Workers, row)
	}

	return result, nil
}

// ListStaffDuties staff / 机务 日排班列表（staff_schedules；role 为 staff 或 maintenance；不含教员）。
func (u *ScheduleUseCase) ListStaffDuties(ctx context.Context, from, to time.Time, userIDs []string) ([]domain.StaffDutyListItem, error) {
	if u.staffScheduleRepo == nil || u.userRepo == nil {
		return nil, fmt.Errorf("staff duties not configured")
	}
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())
	if toD.Before(fromD) {
		return nil, fmt.Errorf("invalid date range")
	}

	rows, err := u.staffScheduleRepo.ListByDateRange(ctx, fromD, toD, userIDs)
	if err != nil {
		return nil, err
	}
	cfg := defaultScheduleConfig()
	out := make([]domain.StaffDutyListItem, 0, len(rows))
	for _, row := range rows {
		uu, err := u.userRepo.GetByID(ctx, row.UserID)
		if err != nil {
			return nil, err
		}
		if !domain.RoleUsesStaffSchedule(uu.Role) {
			continue
		}
		intervals, err := resolveInstructorDutyIntervals(cfg, row.DutyStatus, row.DutyWindowsJSON)
		if err != nil {
			return nil, err
		}
		rawWindows, _ := domain.ParseDutyWindowsJSON(row.DutyWindowsJSON)
		day := time.Date(row.Date.Year(), row.Date.Month(), row.Date.Day(), 0, 0, 0, 0, row.Date.Location())
		out = append(out, domain.StaffDutyListItem{
			ScheduleID:      row.ID,
			UserID:          row.UserID,
			Username:        uu.Username,
			Name:            uu.Username,
			Phone:           uu.Phone,
			Email:           uu.Email,
			Date:            day.Format("2006-01-02"),
			DutyStatus:      row.DutyStatus,
			Remark:          row.Remark,
			DutyWindows:     rawWindows,
			ResolvedWindows: intervalsToResolved(intervals),
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

// UpsertStaffDuty 仅允许 user 为 staff 或 maintenance（共用 staff_schedules）。
func (u *ScheduleUseCase) UpsertStaffDuty(ctx context.Context, s *domain.StaffSchedule) error {
	if u.staffScheduleRepo == nil || u.userRepo == nil {
		return fmt.Errorf("staff duties not configured")
	}
	if s == nil {
		return fmt.Errorf("nil schedule")
	}
	uu, err := u.userRepo.GetByID(ctx, s.UserID)
	if err != nil {
		return err
	}
	if !domain.RoleUsesStaffSchedule(uu.Role) {
		return fmt.Errorf("staff 排班仅适用于 role=staff 或 role=maintenance 的用户")
	}
	s.Date = time.Date(s.Date.Year(), s.Date.Month(), s.Date.Day(), 0, 0, 0, 0, s.Date.Location())
	if err := validateStaffDutyPayload(s); err != nil {
		return err
	}
	cfg := defaultScheduleConfig()
	if _, err := resolveInstructorDutyIntervals(cfg, s.DutyStatus, s.DutyWindowsJSON); err != nil {
		return err
	}
	return u.staffScheduleRepo.UpsertDay(ctx, s)
}

func validateStaffDutyPayload(s *domain.StaffSchedule) error {
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

// BulkStaffDuties 按区间与周几批量写入 staff 排班。
func (u *ScheduleUseCase) BulkStaffDuties(ctx context.Context, userID string, from, to time.Time, weekdays []int, dutyStatus domain.InstructorDutyStatus, remark string, replaceExisting bool) (int, error) {
	if u.staffScheduleRepo == nil || u.userRepo == nil {
		return 0, fmt.Errorf("staff duties not configured")
	}
	uu, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !domain.RoleUsesStaffSchedule(uu.Role) {
		return 0, fmt.Errorf("staff 排班仅适用于 role=staff 或 role=maintenance 的用户")
	}
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

	s := &domain.StaffSchedule{
		UserID:          userID,
		DutyStatus:      dutyStatus,
		Remark:          remark,
		DutyWindowsJSON: nil,
	}
	if err := validateStaffDutyPayload(s); err != nil {
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
			existing, err := u.staffScheduleRepo.GetByUserAndDate(ctx, userID, d)
			if err != nil {
				return n, err
			}
			if existing != nil {
				continue
			}
		}
		s.Date = d
		s.Remark = remark
		if err := u.staffScheduleRepo.UpsertDay(ctx, s); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
