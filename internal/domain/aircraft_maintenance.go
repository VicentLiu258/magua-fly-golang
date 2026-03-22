package domain

import (
	"fmt"
	"time"
)

// HoursReferenceRoleAdvisory 机时/定检间隔仅作参考提醒，不驱动维护窗或自动排期。
const HoursReferenceRoleAdvisory = "advisory"

// 机时提醒分级阈值（小时，可后续接入配置）
const (
	HoursReminderInfoThresholdHours    = 15.0
	HoursReminderWarningThresholdHours = 5.0
)

// MaintenanceWindowStatus 维护时间窗状态
type MaintenanceWindowStatus string

const (
	MaintenancePlanned    MaintenanceWindowStatus = "planned"
	MaintenanceInProgress MaintenanceWindowStatus = "in_progress"
	MaintenanceCompleted  MaintenanceWindowStatus = "completed"
	MaintenanceCancelled  MaintenanceWindowStatus = "cancelled"
)

// MaintenanceWindow 定检/进机库等维护计划
type MaintenanceWindow struct {
	ID                 string                  `json:"id"`
	AircraftID         string                  `json:"aircraft_id"`
	Kind               string                  `json:"kind"`
	ScheduledStartAt   time.Time               `json:"scheduled_start_at"`
	ScheduledEndAt     time.Time               `json:"scheduled_end_at"`
	Status             MaintenanceWindowStatus `json:"status"`
	HangarLocation     string                  `json:"hangar_location,omitempty"`
	Notes              string                  `json:"notes,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// AirworthinessSummary 机时与定检派生信息（排期以维护时间窗为准；本结构为参考与提醒）
type AirworthinessSummary struct {
	HoursReferenceRole      string   `json:"hours_reference_role"` // 固定 advisory
	AircraftID              string   `json:"aircraft_id"`
	TotalAirframeHours      *float64 `json:"total_airframe_hours,omitempty"`
	LastInspectionHours     *float64 `json:"last_inspection_hours,omitempty"`
	InspectionIntervalHours int      `json:"inspection_interval_hours"`
	NextDueHours            *float64 `json:"next_due_hours,omitempty"`
	RemainingHours          *float64 `json:"remaining_hours,omitempty"`
	ReminderLevel           string   `json:"reminder_level"`            // none | info | warning | overdue
	ReminderHint            string   `json:"reminder_hint,omitempty"`   // 简短中文，便于前端展示
}

// ComputeHoursReminder 根据剩余机时（可为 nil）返回提醒级别与提示文案。
func ComputeHoursReminder(remainingHours *float64) (level string, hint string) {
	if remainingHours == nil {
		return "none", ""
	}
	r := *remainingHours
	switch {
	case r <= 0:
		return "overdue", fmt.Sprintf("机时定检参考：已超过参考间隔约 %.1f 小时", -r)
	case r <= HoursReminderWarningThresholdHours:
		return "warning", fmt.Sprintf("机时定检参考：剩余约 %.1f 小时", r)
	case r <= HoursReminderInfoThresholdHours:
		return "info", fmt.Sprintf("机时定检参考：剩余约 %.1f 小时", r)
	default:
		return "none", ""
	}
}

// DefaultMaintenanceWindowKind 人工排定的检修时段默认类型（非机时驱动）
const DefaultMaintenanceWindowKind = "manual"

// MaintenanceWindowSyncResult 创建/更新维护窗后的排班同步结果
type MaintenanceWindowSyncResult struct {
	Window            *MaintenanceWindow `json:"window"`
	AffectedBookings  []string           `json:"affected_booking_ids"`
	CancelledPlans    []string           `json:"cancelled_training_plan_ids"`
}
