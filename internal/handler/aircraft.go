package handler

import (
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

// AircraftHandler handles aircraft endpoints.
type AircraftHandler struct {
	aircraftRepo     repository.AircraftRepository
	availabilityRepo repository.AircraftAvailabilityRepository
}

// NewAircraftHandler creates an AircraftHandler.
func NewAircraftHandler(aircraftRepo repository.AircraftRepository, availabilityRepo repository.AircraftAvailabilityRepository) *AircraftHandler {
	return &AircraftHandler{
		aircraftRepo:     aircraftRepo,
		availabilityRepo: availabilityRepo,
	}
}

// List returns all aircraft.
func (h *AircraftHandler) List(c *gin.Context) {
	list, err := h.aircraftRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": list},
	})
}

// Create creates a new aircraft.
func (h *AircraftHandler) Create(c *gin.Context) {
	var body struct {
		Code   string `json:"code" binding:"required"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	a := &domain.Aircraft{
		Code:   body.Code,
		Name:   body.Name,
		Status: body.Status,
	}
	if a.Status == "" {
		a.Status = "active"
	}
	if err := h.aircraftRepo.Create(c.Request.Context(), a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": a,
	})
}

// CreateAvailability configures availability for an aircraft on a given date.
func (h *AircraftHandler) CreateAvailability(c *gin.Context) {
	var body struct {
		AircraftID int64  `json:"aircraft_id" binding:"required"`
		Date       string `json:"date" binding:"required"`       // YYYY-MM-DD
		StartTime  string `json:"start_time" binding:"required"` // HH:MM
		EndTime    string `json:"end_time" binding:"required"`   // HH:MM
		Status     string `json:"status" binding:"required"`     // available | maintenance | blocked
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	d, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid date", "data": nil})
		return
	}
	start, err := time.Parse("15:04", body.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid start_time", "data": nil})
		return
	}
	end, err := time.Parse("15:04", body.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid end_time", "data": nil})
		return
	}
	if !end.After(start) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "end_time must be after start_time", "data": nil})
		return
	}

	a := &domain.AircraftAvailability{
		AircraftID: body.AircraftID,
		Date:       d,
		StartTime:  start,
		EndTime:    end,
		Status:     domain.AvailabilityStatus(body.Status),
	}
	if err := h.availabilityRepo.Create(c.Request.Context(), a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": a,
	})
}

