package handler

import (
	"net/http"

	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// TaskHandler 处理学员任务相关接口。
type TaskHandler struct {
	uc *usecase.TaskUseCase
}

func NewTaskHandler(uc *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{uc: uc}
}

// getUserIDFromClaims 复用 JWT 中的 user_id。
func getUserIDFromClaims(c *gin.Context) string {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// MyList 学员“我的任务”列表。
func (h *TaskHandler) MyList(c *gin.Context) {
	var body usecase.ListMyTasksRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	userID := getUserIDFromClaims(c)
	resp, err := h.uc.ListMyTasks(c.Request.Context(), userID, &body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": resp})
}

// UpdateStatus 更新任务状态。
func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	var body usecase.UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	userID := getUserIDFromClaims(c)
	if err := h.uc.UpdateTaskStatus(c.Request.Context(), userID, &body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

