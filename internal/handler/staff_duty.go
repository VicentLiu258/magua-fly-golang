package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"

	"github.com/gin-gonic/gin"
)

// StaffDutiesList POST /api/v1/schedules/staff-duties/list
func (h *ScheduleHandler) StaffDutiesList(c *gin.Context) {
	var body struct {
		FromDate string   `json:"from_date" binding:"required"`
		ToDate   string   `json:"to_date" binding:"required"`
		UserIDs  []string `json:"user_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	from, err := time.Parse("2006-01-02", body.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid from_date", "data": nil})
		return
	}
	to, err := time.Parse("2006-01-02", body.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid to_date", "data": nil})
		return
	}

	items, err := h.scheduleUC.ListStaffDuties(c.Request.Context(), from, to, body.UserIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"items": items}})
}

// StaffDutiesUpsert POST /api/v1/schedules/staff-duties/upsert
func (h *ScheduleHandler) StaffDutiesUpsert(c *gin.Context) {
	var body struct {
		UserID      string                      `json:"user_id" binding:"required"`
		Date        string                      `json:"date" binding:"required"`
		DutyStatus  domain.InstructorDutyStatus `json:"duty_status" binding:"required"`
		Remark      string                      `json:"remark"`
		DutyWindows []domain.DutyWindowJSON     `json:"duty_windows"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	day, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid date", "data": nil})
		return
	}

	s := &domain.StaffSchedule{
		UserID:     body.UserID,
		Date:       day,
		DutyStatus: body.DutyStatus,
		Remark:     body.Remark,
	}
	if len(body.DutyWindows) > 0 {
		raw, err := json.Marshal(body.DutyWindows)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		s.DutyWindowsJSON = raw
	}

	if err := h.scheduleUC.UpsertStaffDuty(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": s})
}

// StaffDutiesBulk POST /api/v1/schedules/staff-duties/bulk
func (h *ScheduleHandler) StaffDutiesBulk(c *gin.Context) {
	var body struct {
		UserID          string                      `json:"user_id" binding:"required"`
		FromDate        string                      `json:"from_date" binding:"required"`
		ToDate          string                      `json:"to_date" binding:"required"`
		Weekdays        []int                       `json:"weekdays" binding:"required"`
		DutyStatus      domain.InstructorDutyStatus `json:"duty_status" binding:"required"`
		Remark          string                      `json:"remark"`
		ReplaceExisting bool                        `json:"replace_existing"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	from, err := time.Parse("2006-01-02", body.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid from_date", "data": nil})
		return
	}
	to, err := time.Parse("2006-01-02", body.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid to_date", "data": nil})
		return
	}

	n, err := h.scheduleUC.BulkStaffDuties(c.Request.Context(), body.UserID, from, to, body.Weekdays, body.DutyStatus, body.Remark, body.ReplaceExisting)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"upserted": n}})
}
