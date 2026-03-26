package handler

import (
	"errors"
	"net/http"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// HomeHandler 管理端首页等聚合接口
type HomeHandler struct {
	uc *usecase.HomeUseCase
}

// NewHomeHandler creates HomeHandler.
func NewHomeHandler(uc *usecase.HomeUseCase) *HomeHandler {
	return &HomeHandler{uc: uc}
}

// AdminHome POST /api/v1/admin/home
func (h *HomeHandler) AdminHome(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if domain.UserRole(claims.Role) != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	data, _, err := h.uc.AdminHome(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "user not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}
