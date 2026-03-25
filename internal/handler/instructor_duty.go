package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// InstructorDutiesList POST /api/v1/schedules/instructor-duties/list
func (h *ScheduleHandler) InstructorDutiesList(c *gin.Context) {
	var body struct {
		FromDate       string   `json:"from_date" binding:"required"`
		ToDate         string   `json:"to_date" binding:"required"`
		InstructorIDs  []string `json:"instructor_ids"`
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

	items, err := h.scheduleUC.ListInstructorDuties(c.Request.Context(), from, to, body.InstructorIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"items": items}})
}

// InstructorDutiesMyList POST /api/v1/schedules/instructor-duties/my-list
func (h *ScheduleHandler) InstructorDutiesMyList(c *gin.Context) {
	var body struct {
		FromDate string `json:"from_date" binding:"required"`
		ToDate   string `json:"to_date" binding:"required"`
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

	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}

	data, err := h.scheduleUC.ListMyInstructorDuties(c.Request.Context(), claims.UserID, from, to)
	if err != nil {
		if errors.Is(err, usecase.ErrInstructorProfileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// InstructorDutiesUpsert POST /api/v1/schedules/instructor-duties/upsert
func (h *ScheduleHandler) InstructorDutiesUpsert(c *gin.Context) {
	var body struct {
		InstructorID string                   `json:"instructor_id" binding:"required"`
		Date         string                   `json:"date" binding:"required"`
		DutyStatus   domain.InstructorDutyStatus `json:"duty_status" binding:"required"`
		Remark       string                   `json:"remark"`
		DutyWindows  []domain.DutyWindowJSON  `json:"duty_windows"`
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

	s := &domain.InstructorSchedule{
		InstructorID: body.InstructorID,
		Date:         day,
		DutyStatus:   body.DutyStatus,
		Remark:       body.Remark,
	}
	if len(body.DutyWindows) > 0 {
		raw, err := json.Marshal(body.DutyWindows)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		s.DutyWindowsJSON = raw
	}

	if err := h.scheduleUC.UpsertInstructorDuty(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": s})
}

// InstructorDutiesBulk POST /api/v1/schedules/instructor-duties/bulk
func (h *ScheduleHandler) InstructorDutiesBulk(c *gin.Context) {
	var body struct {
		InstructorID    string                   `json:"instructor_id" binding:"required"`
		FromDate        string                   `json:"from_date" binding:"required"`
		ToDate          string                   `json:"to_date" binding:"required"`
		Weekdays        []int                    `json:"weekdays" binding:"required"`
		DutyStatus      domain.InstructorDutyStatus `json:"duty_status" binding:"required"`
		Remark          string                   `json:"remark"`
		ReplaceExisting bool                     `json:"replace_existing"`
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

	n, err := h.scheduleUC.BulkInstructorDuties(c.Request.Context(), body.InstructorID, from, to, body.Weekdays, body.DutyStatus, body.Remark, body.ReplaceExisting)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"upserted": n}})
}
