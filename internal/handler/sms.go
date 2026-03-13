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
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	req.IP = c.ClientIP()

	if err := h.authUC.SendSmsCode(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": nil,
	})
}

// VerifyCode verifies an SMS code and, for login/register scenes, returns token+user.
func (h *SmsHandler) VerifyCode(c *gin.Context) {
	var req usecase.VerifySmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	result, err := h.authUC.VerifySmsCode(c.Request.Context(), req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid code", "data": nil})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	// For login/register scenes we may have token+user; for others just return success.
	if result != nil && result.Token != "" && result.User != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "ok",
			"data": gin.H{
				"token": result.Token,
				"user":  result.User,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": nil,
	})
}

