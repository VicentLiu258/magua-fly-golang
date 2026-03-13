package handler

import (
	"net/http"
	"time"

	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// ScheduleHandler handles schedule-related endpoints.
type ScheduleHandler struct {
	scheduleUC *usecase.ScheduleUseCase
}

// NewScheduleHandler creates a ScheduleHandler.
func NewScheduleHandler(scheduleUC *usecase.ScheduleUseCase) *ScheduleHandler {
	return &ScheduleHandler{scheduleUC: scheduleUC}
}

// InstructorGrid returns instructor × time grid for a given date.
func (h *ScheduleHandler) InstructorGrid(c *gin.Context) {
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

	grid, err := h.scheduleUC.GetInstructorGrid(c.Request.Context(), day)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": grid,
	})
}

