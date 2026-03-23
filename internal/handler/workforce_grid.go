package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// WorkforceGrid POST /api/v1/schedules/workforce-grid — staff + maintenance + instructor 考勤网格（无预约层）。
func (h *ScheduleHandler) WorkforceGrid(c *gin.Context) {
	var body struct {
		Date string `json:"date" binding:"required"`
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

	grid, err := h.scheduleUC.GetWorkforceGrid(c.Request.Context(), day)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": grid})
}
