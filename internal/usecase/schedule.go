package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// ScheduleUseCase aggregates instructor schedules, bookings and (optionally) aircraft availability
// into per-slot grids for the Schedule page.
type ScheduleUseCase struct {
	instructorRepo       repository.InstructorRepository
	instrScheduleRepo    repository.InstructorScheduleRepository
	bookingRepo          repository.BookingRepository
	planRepo             repository.TrainingPlanRepository
	aircraftAvailRepo    repository.AircraftAvailabilityRepository
}

// NewScheduleUseCase creates a ScheduleUseCase.
func NewScheduleUseCase(
	instructorRepo repository.InstructorRepository,
	instrScheduleRepo repository.InstructorScheduleRepository,
	bookingRepo repository.BookingRepository,
	planRepo repository.TrainingPlanRepository,
	aircraftAvailRepo repository.AircraftAvailabilityRepository,
) *ScheduleUseCase {
	return &ScheduleUseCase{
		instructorRepo:    instructorRepo,
		instrScheduleRepo: instrScheduleRepo,
		bookingRepo:       bookingRepo,
		planRepo:          planRepo,
		aircraftAvailRepo: aircraftAvailRepo,
	}
}

// 业务日可排期/网格时间轴（与预约、教员执勤解析、列表裁剪共用）。
const (
	defaultBusinessDayStart = 9 * time.Hour
	defaultBusinessDayEnd   = 19 * time.Hour
	// defaultHalfDaySplitMin 上/下午执勤分界（自 00:00 起算的分钟数）：上午 [dayStart, split)、下午 [split, dayEnd)。
	defaultHalfDaySplitMin = 14 * 60 // 14:00
)

// scheduleConfig is a simple in-memory configuration for now.
// If needed, it can be moved to DB-backed configuration later.
type scheduleConfig struct {
	dayStart    time.Duration // minutes from midnight, e.g. 9:00
	dayEnd      time.Duration // minutes from midnight, e.g. 19:00
	slotMinutes int
}

func defaultScheduleConfig() scheduleConfig {
	// 09:00–19:00, 30 min per slot。
	return scheduleConfig{
		dayStart:    defaultBusinessDayStart,
		dayEnd:      defaultBusinessDayEnd,
		slotMinutes: 30,
	}
}

// GetInstructorGrid returns the instructor × time grid for a given date.
func (u *ScheduleUseCase) GetInstructorGrid(ctx context.Context, date time.Time) (*domain.InstructorGridResponse, error) {
	cfg := defaultScheduleConfig()

	if cfg.slotMinutes <= 0 {
		return nil, fmt.Errorf("invalid slotMinutes")
	}

	// Build slots.
	totalMinutes := int((cfg.dayEnd - cfg.dayStart).Minutes())
	if totalMinutes <= 0 {
		return nil, fmt.Errorf("invalid day window")
	}
	slotCount := totalMinutes / cfg.slotMinutes
	if slotCount <= 0 {
		return nil, fmt.Errorf("no slots in config")
	}

	slots := make([]string, 0, slotCount)
	slotStarts := make([]time.Duration, 0, slotCount)
	for i := 0; i < slotCount; i++ {
		minutesFromStart := i * cfg.slotMinutes
		t := cfg.dayStart + time.Duration(minutesFromStart)*time.Minute
		h := int(t.Hours())
		m := int(t.Minutes()) % 60
		slots = append(slots, fmt.Sprintf("%02d:%02d", h, m))
		slotStarts = append(slotStarts, t)
	}

	// Fetch instructors.
	instructors, err := u.instructorRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch schedules and bookings for the date.
	dayDate := date
	dayDate = time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), 0, 0, 0, 0, dayDate.Location())

	schedules, err := u.instrScheduleRepo.ListByDate(ctx, dayDate)
	if err != nil {
		return nil, err
	}
	bookings, err := u.bookingRepo.ListByDate(ctx, dayDate)
	if err != nil {
		return nil, err
	}

	var aircraftMaint []*domain.AircraftAvailability
	if u.aircraftAvailRepo != nil {
		aircraftMaint, err = u.aircraftAvailRepo.ListByDate(ctx, dayDate)
		if err != nil {
			return nil, err
		}
	}

	// Index schedules and bookings by instructor.
	schedByInstructor := make(map[string][]*domain.InstructorSchedule)
	for _, s := range schedules {
		schedByInstructor[s.InstructorID] = append(schedByInstructor[s.InstructorID], s)
	}

	bookingsByInstructor := make(map[string][]*domain.Booking)
	for _, b := range bookings {
		bookingsByInstructor[b.InstructorID] = append(bookingsByInstructor[b.InstructorID], b)
	}

	// Helper: convert time-of-day to slot index.
	toSlotIndex := func(t time.Time) int {
		// Use minutes from midnight.
		minutes := t.Hour()*60 + t.Minute()
		rel := minutes - int(cfg.dayStart.Minutes())
		if rel <= 0 {
			return 0
		}
		idx := rel / cfg.slotMinutes
		if idx < 0 {
			return 0
		}
		if idx > slotCount {
			return slotCount
		}
		return idx
	}

	// Build instructor timelines.
	result := &domain.InstructorGridResponse{
		Date:        dayDate.Format("2006-01-02"),
		SlotMinutes: cfg.slotMinutes,
		DayStart:    fmt.Sprintf("%02d:%02d", int(cfg.dayStart.Hours()), int(cfg.dayStart.Minutes())%60),
		DayEnd:      fmt.Sprintf("%02d:%02d", int(cfg.dayEnd.Hours()), int(cfg.dayEnd.Minutes())%60),
		Slots:       slots,
	}

	for _, inst := range instructors {
		// Initialize all slots as off.
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

		// Apply schedules: mark working slots as available (duty_status + duty_windows 解析结果).
		if list, ok := schedByInstructor[inst.ID]; ok {
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

		// Apply bookings: override with booked。
		// 业务日仅含 [dayStart, dayEnd) 的格点；开始时间在 dayEnd 及之后的预约不在网格展示（与 19:00 后不可用一致）。
		dayEndMin := int(cfg.dayEnd.Minutes())
		if list, ok := bookingsByInstructor[inst.ID]; ok {
			for _, b := range list {
				bs := b.StartTime.Hour()*60 + b.StartTime.Minute()
				if bs >= dayEndMin {
					continue
				}
				startIdx := toSlotIndex(b.StartTime)
				endIdx := toSlotIndex(b.EndTime)
				if startIdx < 0 {
					startIdx = 0
				}
				if endIdx > slotCount {
					endIdx = slotCount
				}
				for i := startIdx; i < endIdx; i++ {
					id := b.ID
					aircraftID := b.AircraftID
					studentID := b.StudentID
					slotStart := slotStarts[i]
					slotEnd := slotStart + time.Duration(cfg.slotMinutes)*time.Minute
					st := domain.SlotBooked
					if slotOverlapsAircraftMaintenance(b.AircraftID, aircraftMaint, slotStart, slotEnd) {
						st = domain.SlotAircraftUnavailable
					}
					timeline[i].Status = st
					timeline[i].BookingID = &id
					timeline[i].AircraftID = &aircraftID
					timeline[i].StudentID = &studentID
				}
			}
		}

		result.Instructors = append(result.Instructors, domain.InstructorTimeline{
			ID:       inst.ID,
			Name:     inst.Name,
			Color:    inst.DisplayColor,
			Timeline: timeline,
		})
	}

	return result, nil
}

func slotOverlapsAircraftMaintenance(aircraftID string, maintList []*domain.AircraftAvailability, slotStart, slotEnd time.Duration) bool {
	for _, m := range maintList {
		if m.AircraftID != aircraftID || m.Status != domain.AvailMaintenance {
			continue
		}
		ms := time.Duration(m.StartTime.Hour())*time.Hour + time.Duration(m.StartTime.Minute())*time.Minute + time.Duration(m.StartTime.Second())*time.Second
		me := time.Duration(m.EndTime.Hour())*time.Hour + time.Duration(m.EndTime.Minute())*time.Minute + time.Duration(m.EndTime.Second())*time.Second
		if slotStart < me && ms < slotEnd {
			return true
		}
	}
	return false
}

// dutyInterval 为半开区间 [startMin, endMin)，分钟从 00:00 起算。
type dutyInterval struct {
	startMin int
	endMin   int
}

// resolveInstructorDutyIntervals 与 GetInstructorGrid、预约校验共用。
func resolveInstructorDutyIntervals(cfg scheduleConfig, status domain.InstructorDutyStatus, rawJSON []byte) ([]dutyInterval, error) {
	ds := int(cfg.dayStart.Minutes())
	de := int(cfg.dayEnd.Minutes())
	split := defaultHalfDaySplitMin

	hasExplicit := len(rawJSON) > 0 && string(rawJSON) != "null"
	if hasExplicit {
		windows, err := domain.ParseDutyWindowsJSON(rawJSON)
		if err != nil {
			return nil, err
		}
		if len(windows) == 0 && status == domain.InstructorDutyCustom {
			return nil, nil
		}
		if len(windows) > 0 {
			var out []dutyInterval
			for _, w := range windows {
				sm, em, err := parseHHMMRange(w.Start, w.End)
				if err != nil {
					return nil, err
				}
				sm, em = clipDutyInterval(sm, em, ds, de)
				if sm < em {
					out = append(out, dutyInterval{startMin: sm, endMin: em})
				}
			}
			return out, nil
		}
	}

	switch status {
	case domain.InstructorDutyLeave, domain.InstructorDutyRest:
		return nil, nil
	case domain.InstructorDutyFullDay:
		if ds >= de {
			return nil, nil
		}
		return []dutyInterval{{startMin: ds, endMin: de}}, nil
	case domain.InstructorDutyMorning:
		end := split
		if end > de {
			end = de
		}
		start := ds
		if start >= end {
			return nil, nil
		}
		return []dutyInterval{{startMin: start, endMin: end}}, nil
	case domain.InstructorDutyAfternoon:
		start := split
		if start < ds {
			start = ds
		}
		if start >= de {
			return nil, nil
		}
		return []dutyInterval{{startMin: start, endMin: de}}, nil
	case domain.InstructorDutyCustom:
		return nil, fmt.Errorf("duty_status=custom 时必须提供非空 duty_windows")
	default:
		return nil, fmt.Errorf("unknown duty_status: %s", status)
	}
}

func clipDutyInterval(startMin, endMin, dayStart, dayEnd int) (int, int) {
	if startMin < dayStart {
		startMin = dayStart
	}
	if endMin > dayEnd {
		endMin = dayEnd
	}
	return startMin, endMin
}

func parseHHMMRange(start, end string) (int, int, error) {
	s, err := parseHHMM(start)
	if err != nil {
		return 0, 0, err
	}
	e, err := parseHHMM(end)
	if err != nil {
		return 0, 0, err
	}
	if e <= s {
		return 0, 0, fmt.Errorf("duty window end must be after start")
	}
	return s, e, nil
}

func parseHHMM(s string) (int, error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time %q", s)
	}
	return h*60 + m, nil
}

func slotOverlapsDuty(slotStartMin, slotEndMin int, inv dutyInterval) bool {
	return slotStartMin < inv.endMin && inv.startMin < slotEndMin
}

func bookingIntervalFitsDuty(startTime, endTime time.Time, intervals []dutyInterval) bool {
	if !endTime.After(startTime) {
		return false
	}
	bs := startTime.Hour()*60 + startTime.Minute()
	be := endTime.Hour()*60 + endTime.Minute()
	if be <= bs {
		return false
	}
	for _, inv := range intervals {
		if bs >= inv.startMin && be <= inv.endMin {
			return true
		}
	}
	return false
}

