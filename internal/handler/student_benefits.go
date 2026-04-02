package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/pkg/studentcontract"
	"magua-fly-golang/internal/repository"
	"magua-fly-golang/internal/usecase"
)

// AdminBenefitsOverview POST /api/v1/admin/students/benefits/overview
func (h *StudentHandler) AdminBenefitsOverview(c *gin.Context) {
	if h.benefitsUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "benefits not configured", "data": nil})
		return
	}
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if !isStaffLikeRole(claims.Role) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.benefitsUC.Overview(c.Request.Context(), body.StudentID)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// AdminBenefitsAdjust POST /api/v1/admin/students/benefits/adjust（仅 admin）
func (h *StudentHandler) AdminBenefitsAdjust(c *gin.Context) {
	if h.benefitsUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "benefits not configured", "data": nil})
		return
	}
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if domain.UserRole(claims.Role) != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	var body struct {
		StudentID      string  `json:"student_id" binding:"required"`
		BenefitKind    string  `json:"benefit_kind" binding:"required"`
		Direction      string  `json:"direction" binding:"required"`
		Amount         float64 `json:"amount" binding:"required"`
		Reason         string  `json:"reason"`
		IdempotencyKey string  `json:"idempotency_key"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	opName := ""
	var opUID *string
	if u, err := h.userRepo.GetByID(c.Request.Context(), claims.UserID); err == nil && u != nil {
		opName = u.Username
		opUID = &claims.UserID
	}
	req := &usecase.BenefitAdjustRequest{
		StudentID:      body.StudentID,
		BenefitKind:    body.BenefitKind,
		Direction:      body.Direction,
		Amount:         body.Amount,
		Reason:         body.Reason,
		IdempotencyKey: body.IdempotencyKey,
	}
	res, err := h.benefitsUC.Adjust(c.Request.Context(), req, opUID, opName)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		if errors.Is(err, usecase.ErrBenefitAdjustExceeds) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, usecase.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, repository.ErrBenefitIdempotencyConflict) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "msg": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": res})
}

// AdminBenefitsLogs POST /api/v1/admin/students/benefits/logs
func (h *StudentHandler) AdminBenefitsLogs(c *gin.Context) {
	if h.benefitsUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "benefits not configured", "data": nil})
		return
	}
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if !isStaffLikeRole(claims.Role) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	var body struct {
		StudentID   string `json:"student_id" binding:"required"`
		BenefitKind string `json:"benefit_kind" binding:"required"`
		Page        int    `json:"page"`
		PageSize    int    `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.benefitsUC.Logs(c.Request.Context(), body.StudentID, body.BenefitKind, body.Page, body.PageSize)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// AdminBenefitsExtrasSave POST /api/v1/admin/students/benefits/extras/save
func (h *StudentHandler) AdminBenefitsExtrasSave(c *gin.Context) {
	if h.benefitsUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "benefits not configured", "data": nil})
		return
	}
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if !isStaffLikeRole(claims.Role) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	var body struct {
		StudentID     string                         `json:"student_id" binding:"required"`
		BookingLimits *studentcontract.BookingLimits   `json:"booking_limits"`
		Extras        *studentcontract.Extras        `json:"extras"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	req := &usecase.SaveExtrasRequest{
		StudentID:     body.StudentID,
		BookingLimits: body.BookingLimits,
		Extras:        body.Extras,
	}
	if err := h.benefitsUC.SaveExtras(c.Request.Context(), req); err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		if errors.Is(err, usecase.ErrInvalidContractInfo) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": nil})
}
