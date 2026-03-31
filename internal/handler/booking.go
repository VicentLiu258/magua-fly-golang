package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/repository"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// BookingHandler handles booking endpoints.
type BookingHandler struct {
	bookingRepo repository.BookingRepository
	planRepo    repository.TrainingPlanRepository
	taskRepo    repository.TaskRepository
	scheduleUC  *usecase.ScheduleUseCase
	trainingUC  *usecase.TrainingUseCase
}

// NewBookingHandler creates a BookingHandler.
func NewBookingHandler(bookingRepo repository.BookingRepository, planRepo repository.TrainingPlanRepository, taskRepo repository.TaskRepository, scheduleUC *usecase.ScheduleUseCase, trainingUC *usecase.TrainingUseCase) *BookingHandler {
	return &BookingHandler{
		bookingRepo: bookingRepo,
		planRepo:    planRepo,
		taskRepo:    taskRepo,
		scheduleUC:  scheduleUC,
		trainingUC:  trainingUC,
	}
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
	if h.scheduleUC != nil {
		if err := h.scheduleUC.ValidateInstructorBookingWindow(c.Request.Context(), body.InstructorID, planDate, startTime, endTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
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
	ctx := c.Request.Context()
	if err := h.bookingRepo.UpdateStatus(ctx, body.ID, status, reason, claims.UserID); err != nil {
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
		if err := h.planRepo.Create(ctx, plan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
		if h.trainingUC != nil {
			if err := h.trainingUC.EnsureScheduleDraftRecord(ctx, plan); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
				return
			}
		}
		// 关键：这里不要再 `_ =` 吞掉错误，直接返回给前端看
		if err := h.createFlightTaskFromBooking(ctx, booking, claims.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
		// _ = h.planRepo.Create(ctx, plan)

		// // 审核通过后为学员创建一条飞行任务（首次任务），失败仅记录日志，不阻断审核流程。
		// _ = h.createFlightTaskFromBooking(ctx, booking, claims.UserID)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"status": string(status)},
	})
}

// createFlightTaskFromBooking 在预约审核通过后，为学员创建一条飞行类任务（type=flight，ref_id=booking_id）。
// 若任务已存在或学生 ID 为空，则直接返回。
func (h *BookingHandler) createFlightTaskFromBooking(ctx context.Context, booking *domain.Booking, assignerID string) error {
	if booking == nil || booking.StudentID == "" || h.taskRepo == nil {
		return nil
	}
	// 幂等：若已存在同 type+ref_id 的任务则不再创建。
	existing, err := h.taskRepo.FindByTypeAndRefID(ctx, domain.TaskTypeFlight, booking.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	title := fmt.Sprintf(
		"飞行训练：%s %s-%s",
		booking.PlanDate.Format("2006-01-02"),
		booking.StartTime.Format("15:04"),
		booking.EndTime.Format("15:04"),
	)
	desc := fmt.Sprintf("预约飞行任务，教员ID=%s，飞机ID=%s", booking.InstructorID, booking.AircraftID)

	// 将 plan_date 与 start_time 组合成一个 dueAt 时间（使用 plan_date 的时区）。
	dueAt := time.Date(
		booking.PlanDate.Year(), booking.PlanDate.Month(), booking.PlanDate.Day(),
		booking.StartTime.Hour(), booking.StartTime.Minute(), 0, 0,
		booking.PlanDate.Location(),
	)

	refID := booking.ID
	task := &domain.Task{
		StudentID:   booking.StudentID,
		AssignerID:  assignerID,
		Type:        domain.TaskTypeFlight,
		RefID:       &refID,
		Title:       title,
		Description: desc,
		DueAt:       &dueAt,
		Status:      domain.TaskStatusPending,
	}
	return h.taskRepo.Create(ctx, task)
}
