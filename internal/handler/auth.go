package handler

import (
	"net/http"

	"magua-fly-golang/internal/domain"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.authUC.Login(c.Request.Context(), req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != domain.RoleAdmin && req.Role != domain.RoleStaff && req.Role != domain.RoleInstructor && req.Role != domain.RoleStudent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	user, err := h.authUC.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}
