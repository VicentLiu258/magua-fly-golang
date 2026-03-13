package usecase

import (
	"context"
	"fmt"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// ScheduleUseCase aggregates instructor schedules, bookings and (optionally) aircraft availability
// into per-slot grids for the Schedule page.
type ScheduleUseCase struct {
	instructorRepo    repository.InstructorRepository
	instrScheduleRepo repository.InstructorScheduleRepository
	bookingRepo       repository.BookingRepository
}

// NewScheduleUseCase creates a ScheduleUseCase.
func NewScheduleUseCase(
	instructorRepo repository.InstructorRepository,
	instrScheduleRepo repository.InstructorScheduleRepository,
	bookingRepo repository.BookingRepository,
) *ScheduleUseCase {
	return &ScheduleUseCase{
		instructorRepo:    instructorRepo,
		instrScheduleRepo: instrScheduleRepo,
		bookingRepo:       bookingRepo,
	}
}

// scheduleConfig is a simple in-memory configuration for now.
// If needed, it can be moved to DB-backed configuration later.
type scheduleConfig struct {
	dayStart    time.Duration // minutes from midnight, e.g. 9:00
	dayEnd      time.Duration // minutes from midnight, e.g. 21:00
	slotMinutes int
}

func defaultScheduleConfig() scheduleConfig {
	// 09:00–21:00, 30 min per slot.
	return scheduleConfig{
		dayStart:    9 * time.Hour,
		dayEnd:      21 * time.Hour,
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

	// Index schedules and bookings by instructor.
	schedByInstructor := make(map[int64][]*domain.InstructorSchedule)
	for _, s := range schedules {
		schedByInstructor[s.InstructorID] = append(schedByInstructor[s.InstructorID], s)
	}

	bookingsByInstructor := make(map[int64][]*domain.Booking)
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

		// Apply schedules: mark working slots as available.
		if list, ok := schedByInstructor[inst.ID]; ok {
			for _, s := range list {
				startIdx := toSlotIndex(s.StartTime)
				endIdx := toSlotIndex(s.EndTime)
				if startIdx < 0 {
					startIdx = 0
				}
				if endIdx > slotCount {
					endIdx = slotCount
				}
				for i := startIdx; i < endIdx; i++ {
					if timeline[i].Status == domain.SlotOff {
						timeline[i].Status = domain.SlotAvailable
					}
				}
			}
		}

		// Apply bookings: override with booked.
		if list, ok := bookingsByInstructor[inst.ID]; ok {
			for _, b := range list {
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
					timeline[i].Status = domain.SlotBooked
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

