package handler

import (
	"errors"
	"net/http"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// PersonnelHandler 管理员删除人员（软删除）
type PersonnelHandler struct {
	uc *usecase.PersonnelUseCase
}

// NewPersonnelHandler 创建 PersonnelHandler。
func NewPersonnelHandler(uc *usecase.PersonnelUseCase) *PersonnelHandler {
	return &PersonnelHandler{uc: uc}
}

// AdminDelete POST /api/v1/admin/personnel/delete
func (h *PersonnelHandler) AdminDelete(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if domain.UserRole(claims.Role) != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	var body usecase.SoftDeletePersonnelRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if err := h.uc.SoftDelete(c.Request.Context(), body); err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		case errors.Is(err, usecase.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found", "data": nil})
		case errors.Is(err, usecase.ErrPersonnelStaffRole):
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": usecase.ErrPersonnelStaffRole.Error(), "data": nil})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": nil})
}
