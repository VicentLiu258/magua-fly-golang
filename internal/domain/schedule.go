package domain

import "time"

// Aircraft represents an aircraft.
type Aircraft struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	ID         int64              `json:"id"`
	AircraftID int64              `json:"aircraft_id"`
	Date       time.Time          `json:"date"`
	StartTime  time.Time          `json:"start_time"`
	EndTime    time.Time          `json:"end_time"`
	Status     AvailabilityStatus `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// InstructorSchedule is a duty slot for an instructor.
type InstructorSchedule struct {
	ID           int64     `json:"id"`
	InstructorID int64     `json:"instructor_id"`
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
	ID           int64      `json:"id"`
	StudentID    int64      `json:"student_id"`
	InstructorID int64      `json:"instructor_id"`
	AircraftID   int64      `json:"aircraft_id"`
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
	ID           int64         `json:"id"`
	StudentID    int64         `json:"student_id"`
	InstructorID int64         `json:"instructor_id"`
	AircraftID   int64         `json:"aircraft_id"`
	PlanDate     time.Time     `json:"plan_date"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Status       BookingStatus `json:"status"`
	RejectReason string        `json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time    `json:"reviewed_at,omitempty"`
	ReviewerID   *int64        `json:"reviewer_id,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
