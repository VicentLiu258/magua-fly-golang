package handler

import (
	"errors"
	"net/http"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles auth endpoints.
type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body usecase.LoginRequest true "credentials"
// @Success 200 {object} usecase.LoginResponse
// @Failure 400,401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	resp, err := h.authUC.Login(c.Request.Context(), req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid credentials", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": resp,
	})
}

// Register godoc
// @Summary Register
// @Tags auth
// @Accept json
// @Produce json
// @Param body body usecase.RegisterRequest true "user"
// @Success 201 {object} domain.User
// @Failure 400,500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req usecase.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	switch req.Role {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleMaintenance, domain.RoleInstructor, domain.RoleStudent:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid role", "data": nil})
		return
	}
	user, err := h.authUC.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": user,
	})
}

// Me POST /api/v1/users/me 当前用户信息（需 JWT）。
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	user, err := h.authUC.GetMe(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": user})
}

// UpdateProfile POST /api/v1/users/update 更新当前用户资料（需 JWT）。
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	var req usecase.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	user, err := h.authUC.UpdateProfile(c.Request.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrUsernameConflict) || errors.Is(err, usecase.ErrPhoneConflict) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": user})
}
