package handler

import (
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AirworthinessGet POST 查询机时与定检派生信息
func (h *AircraftHandler) AirworthinessGet(c *gin.Context) {
	var body struct {
		AircraftID string `json:"aircraft_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	a, err := h.aircraftRepo.GetByID(c.Request.Context(), body.AircraftID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": summarizeAirworthiness(a)})
}

// AirworthinessUpdate POST 更新机时与定检基准
func (h *AircraftHandler) AirworthinessUpdate(c *gin.Context) {
	var body struct {
		AircraftID              string   `json:"aircraft_id" binding:"required"`
		TotalAirframeHours      *float64 `json:"total_airframe_hours"`
		LastInspectionHours     *float64 `json:"last_inspection_hours"`
		InspectionIntervalHours *int     `json:"inspection_interval_hours"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if err := h.aircraftRepo.UpdateAirworthiness(c.Request.Context(), body.AircraftID, body.TotalAirframeHours, body.LastInspectionHours, body.InspectionIntervalHours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	a, err := h.aircraftRepo.GetByID(c.Request.Context(), body.AircraftID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"updated": true}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": summarizeAirworthiness(a)})
}

// summarizeAirworthiness 计算机时参考与提醒。排期依据仅为维护时间窗；机时字段不驱动自动排期。
func summarizeAirworthiness(a *domain.Aircraft) *domain.AirworthinessSummary {
	s := &domain.AirworthinessSummary{
		HoursReferenceRole:      domain.HoursReferenceRoleAdvisory,
		AircraftID:              a.ID,
		TotalAirframeHours:      a.TotalAirframeHours,
		LastInspectionHours:     a.LastInspectionHours,
		InspectionIntervalHours: 50,
	}
	if a.InspectionIntervalHours > 0 {
		s.InspectionIntervalHours = a.InspectionIntervalHours
	}
	if a.LastInspectionHours != nil {
		next := *a.LastInspectionHours + float64(s.InspectionIntervalHours)
		s.NextDueHours = &next
	}
	if a.TotalAirframeHours != nil && s.NextDueHours != nil {
		rem := *s.NextDueHours - *a.TotalAirframeHours
		s.RemainingHours = &rem
	}
	s.ReminderLevel, s.ReminderHint = domain.ComputeHoursReminder(s.RemainingHours)
	return s
}

// MaintenanceWindowsList POST 列出维护时间窗
func (h *AircraftHandler) MaintenanceWindowsList(c *gin.Context) {
	var body struct {
		AircraftID string  `json:"aircraft_id"` // 可选；省略表示所有飞机
		FromDate   *string `json:"from_date"`   // YYYY-MM-DD optional
		ToDate     *string `json:"to_date"`
		Status     *string `json:"status"` // planned | in_progress | completed | cancelled
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	var from, to *time.Time
	if body.FromDate != nil && *body.FromDate != "" {
		t, err := time.Parse("2006-01-02", *body.FromDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid from_date", "data": nil})
			return
		}
		from = &t
	}
	if body.ToDate != nil && *body.ToDate != "" {
		t, err := time.Parse("2006-01-02", *body.ToDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid to_date", "data": nil})
			return
		}
		// inclusive end of day for range filter on scheduled_start_at
		eod := t.Add(24*time.Hour - time.Nanosecond)
		to = &eod
	}
	var st *domain.MaintenanceWindowStatus
	if body.Status != nil && *body.Status != "" {
		s := domain.MaintenanceWindowStatus(*body.Status)
		st = &s
	}
	list, err := h.maintenanceRepo.ListByAircraft(c.Request.Context(), body.AircraftID, from, to, st)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"items": list}})
}

// MaintenanceWindowsCreate POST 创建维护窗并写入 availability，可选同步拒绝冲突预约
func (h *AircraftHandler) MaintenanceWindowsCreate(c *gin.Context) {
	var body struct {
		AircraftID         string `json:"aircraft_id" binding:"required"`
		Kind               string `json:"kind"`
		ScheduledStartAt   string `json:"scheduled_start_at" binding:"required"` // RFC3339
		ScheduledEndAt     string `json:"scheduled_end_at"`                      // optional RFC3339，省略则取开始日 23:59:59
		HangarLocation     string `json:"hangar_location"`
		Notes              string `json:"notes"`
		SyncSchedule       bool   `json:"sync_schedule"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	startAt, err := time.Parse(time.RFC3339, body.ScheduledStartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid scheduled_start_at (use RFC3339)", "data": nil})
		return
	}
	var endAt time.Time
	if body.ScheduledEndAt != "" {
		endAt, err = time.Parse(time.RFC3339, body.ScheduledEndAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid scheduled_end_at", "data": nil})
			return
		}
	} else {
		loc := startAt.Location()
		y, m, d := startAt.Date()
		endAt = time.Date(y, m, d, 23, 59, 59, 0, loc)
	}
	if !endAt.After(startAt) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "scheduled_end must be after scheduled_start", "data": nil})
		return
	}
	reviewerID := ""
	if claims := middleware.GetGinClaims(c); claims != nil {
		reviewerID = claims.UserID
	}
	res, err := h.maintSync.CreateMaintenanceWindowWithSync(c.Request.Context(), body.AircraftID, body.Kind, startAt, endAt, body.HangarLocation, body.Notes, body.SyncSchedule, reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": res})
}

// MaintenanceWindowsUpdate POST 更新维护窗
func (h *AircraftHandler) MaintenanceWindowsUpdate(c *gin.Context) {
	var body struct {
		ID                 string `json:"id" binding:"required"`
		AircraftID         string `json:"aircraft_id"`
		Kind               string `json:"kind"`
		ScheduledStartAt   string `json:"scheduled_start_at" binding:"required"`
		ScheduledEndAt     string `json:"scheduled_end_at"`
		HangarLocation     string `json:"hangar_location"`
		Notes              string `json:"notes"`
		SyncSchedule       bool   `json:"sync_schedule"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	startAt, err := time.Parse(time.RFC3339, body.ScheduledStartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid scheduled_start_at", "data": nil})
		return
	}
	var endAt time.Time
	if body.ScheduledEndAt != "" {
		endAt, err = time.Parse(time.RFC3339, body.ScheduledEndAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid scheduled_end_at", "data": nil})
			return
		}
	} else {
		loc := startAt.Location()
		y, m, d := startAt.Date()
		endAt = time.Date(y, m, d, 23, 59, 59, 0, loc)
	}
	if !endAt.After(startAt) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "scheduled_end must be after scheduled_start", "data": nil})
		return
	}
	reviewerID := ""
	if claims := middleware.GetGinClaims(c); claims != nil {
		reviewerID = claims.UserID
	}
	kind := body.Kind
	if kind == "" {
		kind = domain.DefaultMaintenanceWindowKind
	}
	res, err := h.maintSync.UpdateMaintenanceWindowWithSync(c.Request.Context(), body.ID, body.AircraftID, kind, startAt, endAt, body.HangarLocation, body.Notes, body.SyncSchedule, reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": res})
}

// MaintenanceWindowsCancel POST 取消维护窗
func (h *AircraftHandler) MaintenanceWindowsCancel(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if err := h.maintSync.CancelMaintenanceWindow(c.Request.Context(), body.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"cancelled": true}})
}
