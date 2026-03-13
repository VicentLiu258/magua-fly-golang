package handler

import (
	"net/http"

	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SmsHandler handles sms verification endpoints.
type SmsHandler struct {
	authUC *usecase.AuthUseCase
}

// NewSmsHandler creates an SmsHandler.
func NewSmsHandler(authUC *usecase.AuthUseCase) *SmsHandler {
	return &SmsHandler{authUC: authUC}
}

// SendCode sends an SMS verification code for a given scene.
func (h *SmsHandler) SendCode(c *gin.Context) {
	var req usecase.SendSmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.IP = c.ClientIP()

	if err := h.authUC.SendSmsCode(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// VerifyCode verifies an SMS code and, for login/register scenes, returns token+user.
func (h *SmsHandler) VerifyCode(c *gin.Context) {
	var req usecase.VerifySmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authUC.VerifySmsCode(c.Request.Context(), req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For login/register scenes we may have token+user; for others just return success.
	if result != nil && result.Token != "" && result.User != nil {
		c.JSON(http.StatusOK, gin.H{
			"token": result.Token,
			"user":  result.User,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

