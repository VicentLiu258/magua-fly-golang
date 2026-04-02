package handler

import (
	"errors"
	"net/http"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/repository"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// InstructorHandler handles instructor endpoints.
type InstructorHandler struct {
	instructorRepo repository.InstructorRepository
	homeUC         *usecase.HomeUseCase
}

// NewInstructorHandler creates an InstructorHandler.
func NewInstructorHandler(instructorRepo repository.InstructorRepository, homeUC *usecase.HomeUseCase) *InstructorHandler {
	return &InstructorHandler{instructorRepo: instructorRepo, homeUC: homeUC}
}

// List returns all instructors.
func (h *InstructorHandler) List(c *gin.Context) {
	onlyActive := true
	if claims := middleware.GetGinClaims(c); claims != nil && domain.UserRole(claims.Role) == domain.RoleAdmin {
		onlyActive = false
	}
	list, err := h.instructorRepo.List(c.Request.Context(), onlyActive)
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

// MyHome 教员首页聚合 POST /api/v1/instructors/my-home
func (h *InstructorHandler) MyHome(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if domain.UserRole(claims.Role) != domain.RoleInstructor {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	if h.homeUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "home aggregate not configured", "data": nil})
		return
	}
	var body usecase.InstructorHomeRequest
	_ = c.ShouldBindJSON(&body)
	data, _, err := h.homeUC.InstructorHome(c.Request.Context(), claims.UserID, &body)
	if err != nil {
		if errors.Is(err, usecase.ErrInstructorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "instructor not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}
