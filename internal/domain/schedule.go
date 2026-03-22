package domain

import "time"

// Aircraft represents an aircraft.
type Aircraft struct {
	ID                      string    `json:"id"`
	Code                    string    `json:"code"`
	Name                    string    `json:"name,omitempty"`
	Status                  string    `json:"status"`
	TotalAirframeHours      *float64  `json:"total_airframe_hours,omitempty"`
	LastInspectionHours     *float64  `json:"last_inspection_hours,omitempty"`
	InspectionIntervalHours int       `json:"inspection_interval_hours,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// AvailabilityStatus for aircraft slots.
type AvailabilityStatus string

const (
	AvailAvailable  AvailabilityStatus = "available"
	AvailMaintenance AvailabilityStatus = "maintenance"
	AvailBlocked    AvailabilityStatus = "blocked"
)

// AircraftAvailability is a time slot for one aircraft.
type AircraftAvailability struct {
	ID                  string             `json:"id"`
	AircraftID          string             `json:"aircraft_id"`
	Date                time.Time          `json:"date"`
	StartTime           time.Time          `json:"start_time"`
	EndTime             time.Time          `json:"end_time"`
	Status              AvailabilityStatus `json:"status"`
	MaintenanceWindowID *string            `json:"maintenance_window_id,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

// InstructorSchedule is a duty slot for an instructor.
type InstructorSchedule struct {
	ID           string    `json:"id"`
	InstructorID string    `json:"instructor_id"`
	Date         time.Time `json:"date"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PlanSource indicates how a training plan was created.
type PlanSource string

const (
	PlanSourceBooking PlanSource = "booking"
	PlanSourceManual  PlanSource = "manual"
)

// TrainingPlan is a confirmed training slot.
type TrainingPlan struct {
	ID           string     `json:"id"`
	StudentID    string     `json:"student_id"`
	InstructorID string     `json:"instructor_id"`
	AircraftID   string     `json:"aircraft_id"`
	PlanDate     time.Time  `json:"plan_date"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	Source       PlanSource `json:"source"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// BookingStatus for student booking.
type BookingStatus string

const (
	BookingPending  BookingStatus = "pending"
	BookingApproved BookingStatus = "approved"
	BookingRejected BookingStatus = "rejected"
)

// Booking is a student booking request.
type Booking struct {
	ID           string        `json:"id"`
	StudentID    string        `json:"student_id"`
	InstructorID string        `json:"instructor_id"`
	AircraftID   string        `json:"aircraft_id"`
	PlanDate     time.Time     `json:"plan_date"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Status       BookingStatus `json:"status"`
	RejectReason string        `json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time    `json:"reviewed_at,omitempty"`
	ReviewerID   *string       `json:"reviewer_id,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// SlotStatus represents the status of a single time slot in the schedule grid.
type SlotStatus string

const (
	SlotOff                SlotStatus = "off"
	SlotAvailable          SlotStatus = "available"
	SlotBooked             SlotStatus = "booked"
	SlotAircraftUnavailable SlotStatus = "aircraft_unavailable"
)

// TimeSlot is one half-hour block on the timeline.
type TimeSlot struct {
	Index      int     `json:"slotIndex"`
	StartTime  string  `json:"startTime"`           // "09:00"
	EndTime    string  `json:"endTime"`             // "09:30"
	Status     SlotStatus `json:"status"`           // off/available/booked/aircraft_unavailable
	BookingID  *string `json:"bookingId,omitempty"` // when booked
	AircraftID *string `json:"aircraftId,omitempty"`
	StudentID  *string `json:"studentId,omitempty"`
}

// InstructorTimeline is the per-instructor row in the grid.
type InstructorTimeline struct {
	ID       string   `json:"id"`
	Name     string     `json:"name"`
	Color    string     `json:"displayColor,omitempty"`
	Timeline []TimeSlot `json:"timeline"`
}

// InstructorGridResponse is used by /schedules/instructor-grid.
type InstructorGridResponse struct {
	Date        string               `json:"date"`
	SlotMinutes int                  `json:"slotMinutes"`
	DayStart    string               `json:"dayStart"`
	DayEnd      string               `json:"dayEnd"`
	Slots       []string             `json:"slots"`
	Instructors []InstructorTimeline `json:"instructors"`
}

