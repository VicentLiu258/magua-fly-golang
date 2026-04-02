package handler

import (
	"errors"
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/repository"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// InstructorHandler handles instructor endpoints.
type InstructorHandler struct {
	instructorRepo repository.InstructorRepository
	userRepo       repository.UserRepository
	homeUC         *usecase.HomeUseCase
}

// NewInstructorHandler creates an InstructorHandler.
func NewInstructorHandler(instructorRepo repository.InstructorRepository, userRepo repository.UserRepository, homeUC *usecase.HomeUseCase) *InstructorHandler {
	return &InstructorHandler{instructorRepo: instructorRepo, userRepo: userRepo, homeUC: homeUC}
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
	ctx := c.Request.Context()
	type item struct {
		ID                 string    `json:"id"`
		UserID             string    `json:"user_id"`
		Name               string    `json:"name"`
		DisplayColor       string    `json:"display_color,omitempty"`
		Username           string    `json:"username,omitempty"`
		Phone              string    `json:"phone,omitempty"`
		Email              string    `json:"email,omitempty"`
		EmploymentStatus   string    `json:"employment_status,omitempty"`
		PersonnelCategory  string    `json:"personnel_category,omitempty"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
	}
	out := make([]item, 0, len(list))
	for _, ins := range list {
		it := item{
			ID:           ins.ID,
			UserID:       ins.UserID,
			Name:         ins.Name,
			DisplayColor: ins.DisplayColor,
			CreatedAt:    ins.CreatedAt,
			UpdatedAt:    ins.UpdatedAt,
		}
		if ins.UserID != "" {
			if u, err := h.userRepo.GetByID(ctx, ins.UserID); err == nil && u != nil {
				it.Username = u.Username
				it.Phone = u.Phone
				it.Email = u.Email
				it.EmploymentStatus = string(u.EmploymentStatus)
				it.PersonnelCategory = string(domain.DerivePersonnelCategory(u.Role))
			}
		}
		out = append(out, it)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": out},
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
