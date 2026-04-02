package handler

import (
	"errors"
	"net/http"
	"strings"

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

// AdminSave POST /api/v1/admin/personnel/save — 统一新建/编辑学员、教员、地面职员（仅 admin）
func (h *PersonnelHandler) AdminSave(c *gin.Context) {
	var body usecase.SavePersonnelRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if strings.TrimSpace(body.Kind) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "kind required", "data": nil})
		return
	}
	res, err := h.uc.SavePersonnel(c.Request.Context(), &body)
	if err != nil {
		writePersonnelSaveErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": res})
}

func writePersonnelSaveErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
	case errors.Is(err, usecase.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found", "data": nil})
	case errors.Is(err, usecase.ErrUsernameConflict):
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": usecase.ErrUsernameConflict.Error(), "data": nil})
	case errors.Is(err, usecase.ErrPersonnelStaffRole):
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": usecase.ErrPersonnelStaffRole.Error(), "data": nil})
	case errors.Is(err, usecase.ErrInvalidContractInfo):
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": usecase.ErrInvalidContractInfo.Error(), "data": nil})
	case errors.Is(err, usecase.ErrInvalidTrainingProgressJSON):
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": usecase.ErrInvalidTrainingProgressJSON.Error(), "data": nil})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
	}
}
