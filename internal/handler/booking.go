package handler

import (
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

// BookingHandler handles booking endpoints.
type BookingHandler struct {
	bookingRepo repository.BookingRepository
	planRepo    repository.TrainingPlanRepository
}

// NewBookingHandler creates a BookingHandler.
func NewBookingHandler(bookingRepo repository.BookingRepository, planRepo repository.TrainingPlanRepository) *BookingHandler {
	return &BookingHandler{bookingRepo: bookingRepo, planRepo: planRepo}
}

// ListPending returns all pending bookings (for staff).
func (h *BookingHandler) ListPending(c *gin.Context) {
	list, err := h.bookingRepo.ListPending(c.Request.Context())
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

// Create creates a new booking (student).
func (h *BookingHandler) Create(c *gin.Context) {
	var body struct {
		StudentID    string `json:"student_id" binding:"required"`
		InstructorID string `json:"instructor_id" binding:"required"`
		AircraftID   string `json:"aircraft_id" binding:"required"`
		PlanDate     string `json:"plan_date" binding:"required"`
		StartTime    string `json:"start_time" binding:"required"`
		EndTime      string `json:"end_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	planDate, err := time.Parse("2006-01-02", body.PlanDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid plan_date", "data": nil})
		return
	}
	startTime, err := time.Parse("15:04", body.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid start_time", "data": nil})
		return
	}
	endTime, err := time.Parse("15:04", body.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid end_time", "data": nil})
		return
	}
	b := &domain.Booking{
		StudentID:    body.StudentID,
		InstructorID: body.InstructorID,
		AircraftID:   body.AircraftID,
		PlanDate:     planDate,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       domain.BookingPending,
	}
	if err := h.bookingRepo.Create(c.Request.Context(), b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": b,
	})
}

// Review approves or rejects a booking (staff).
func (h *BookingHandler) Review(c *gin.Context) {
	var body struct {
		ID      string `json:"id" binding:"required"`
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	// Ensure booking exists before updating status.
	booking, err := h.bookingRepo.GetByID(c.Request.Context(), body.ID)
	if err != nil || booking == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "booking not found", "data": nil})
		return
	}

	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	status := domain.BookingApproved
	reason := body.Reason
	if !body.Approve {
		status = domain.BookingRejected
	}
	if err := h.bookingRepo.UpdateStatus(c.Request.Context(), body.ID, status, reason, claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if status == domain.BookingApproved {
		plan := &domain.TrainingPlan{
			StudentID:    booking.StudentID,
			InstructorID: booking.InstructorID,
			AircraftID:   booking.AircraftID,
			PlanDate:     booking.PlanDate,
			StartTime:    booking.StartTime,
			EndTime:      booking.EndTime,
			Source:       domain.PlanSourceBooking,
		}
		_ = h.planRepo.Create(c.Request.Context(), plan)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"status": string(status)},
	})
}
