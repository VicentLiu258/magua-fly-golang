package domain

import (
	"encoding/json"
	"time"
)

// InstructorDutyStatus 教员日出勤状态（与 instructor_schedules.duty_status 一致）。
type InstructorDutyStatus string

const (
	InstructorDutyMorning   InstructorDutyStatus = "morning"
	InstructorDutyAfternoon InstructorDutyStatus = "afternoon"
	InstructorDutyFullDay   InstructorDutyStatus = "full_day"
	InstructorDutyLeave     InstructorDutyStatus = "leave"
	InstructorDutyRest      InstructorDutyStatus = "rest"
	InstructorDutyCustom    InstructorDutyStatus = "custom"
)

// DutyWindowJSON 对应 duty_windows 数组元素：HH:MM。
type DutyWindowJSON struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ParseDutyWindowsJSON 解析 duty_windows JSON 数组。
func ParseDutyWindowsJSON(raw []byte) ([]DutyWindowJSON, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var w []DutyWindowJSON
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	return w, nil
}

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

// InstructorSchedule is one row per instructor per calendar day (duty).
type InstructorSchedule struct {
	ID              string               `json:"id"`
	InstructorID    string               `json:"instructor_id"`
	Date            time.Time            `json:"date"`
	DutyStatus      InstructorDutyStatus `json:"duty_status"`
	Remark          string               `json:"remark,omitempty"`
	DutyWindowsJSON []byte               `json:"duty_windows,omitempty"` // null → 由 duty_status 推导
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// DutyTaskBundle 列表接口中带飞/地面任务聚合结构。
type DutyTaskBundle struct {
	Items        []DutyTaskItem `json:"items"`
	TotalMinutes int            `json:"total_minutes"`
}

// DutyTaskItem 单条任务（带飞或地面，字段一致便于前端复用）。
type DutyTaskItem struct {
	Source    string `json:"source"` // booking | training_plan | ground (预留)
	RefID     string `json:"ref_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Minutes   int    `json:"minutes"`
	StudentID string `json:"student_id,omitempty"`
	AircraftID string `json:"aircraft_id,omitempty"`
	Title     string `json:"title,omitempty"`
}

// ResolvedDutyWindow 响应中解析后的执勤窗口（分钟数从 00:00 起算，用于展示）。
type ResolvedDutyWindow struct {
	Start string `json:"start"` // HH:MM
	End   string `json:"end"`   // HH:MM
}

// InstructorDutyListItem 教员排班列表一行。
type InstructorDutyListItem struct {
	InstructorID     string               `json:"instructor_id"`
	InstructorName   string               `json:"instructor_name"`
	Date             string               `json:"date"` // YYYY-MM-DD
	DutyStatus       InstructorDutyStatus `json:"duty_status"`
	Remark           string               `json:"remark,omitempty"`
	DutyWindows      []DutyWindowJSON     `json:"duty_windows,omitempty"`   // 库内原始或空
	ResolvedWindows  []ResolvedDutyWindow `json:"resolved_windows"`         // 推导/合并后
	FlightTasks      DutyTaskBundle       `json:"flight_tasks"`
	GroundCourses    DutyTaskBundle       `json:"ground_courses"`
	ScheduleID       string               `json:"schedule_id,omitempty"`
	CreatedAt        time.Time            `json:"created_at,omitempty"`
	UpdatedAt        time.Time            `json:"updated_at,omitempty"`
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

// TimeSlot is one grid cell on the timeline; width is given by InstructorGridResponse.SlotMinutes / WorkforceGridResponse.SlotMinutes.
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

// StaffSchedule 地面工作人员日排班（users.id，角色为 staff 或 maintenance），字段语义与 InstructorSchedule 一致，不可预约。
type StaffSchedule struct {
	ID              string               `json:"id"`
	UserID          string               `json:"user_id"`
	Date            time.Time            `json:"date"`
	DutyStatus      InstructorDutyStatus `json:"duty_status"`
	Remark          string               `json:"remark,omitempty"`
	DutyWindowsJSON []byte               `json:"duty_windows,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// WorkforceMemberRow 工作人员网格一行（含教员 + staff + 机务 maintenance，不含 admin/student）。
type WorkforceMemberRow struct {
	UserID       string    `json:"userId"`
	Role         UserRole  `json:"role"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	InstructorID *string   `json:"instructorId,omitempty"`
	DisplayColor string    `json:"displayColor,omitempty"`
	Timeline     []TimeSlot `json:"timeline"`
}

// WorkforceGridResponse is used by /schedules/workforce-grid（考勤网格，无预约覆盖）。
type WorkforceGridResponse struct {
	Date        string               `json:"date"`
	SlotMinutes int                  `json:"slotMinutes"`
	DayStart    string               `json:"dayStart"`
	DayEnd      string               `json:"dayEnd"`
	Slots       []string             `json:"slots"`
	Workers     []WorkforceMemberRow `json:"workers"`
}

// StaffDutyListItem staff 排班列表一行（无带飞/地面任务聚合）。
type StaffDutyListItem struct {
	ScheduleID      string               `json:"schedule_id,omitempty"`
	UserID          string               `json:"user_id"`
	Username        string               `json:"username"`
	Name            string               `json:"name"`
	Date            string               `json:"date"`
	DutyStatus      InstructorDutyStatus `json:"duty_status"`
	Remark          string               `json:"remark,omitempty"`
	DutyWindows     []DutyWindowJSON     `json:"duty_windows,omitempty"`
	ResolvedWindows []ResolvedDutyWindow `json:"resolved_windows"`
	CreatedAt       time.Time            `json:"created_at,omitempty"`
	UpdatedAt       time.Time            `json:"updated_at,omitempty"`
}

