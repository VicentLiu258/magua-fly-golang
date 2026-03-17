package handler

import (
	"io"
	"net/http"

	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// ContractFilesHandler 处理合同文件上传/列表接口。
type ContractFilesHandler struct {
	uc *usecase.ContractFilesUseCase
}

func NewContractFilesHandler(uc *usecase.ContractFilesUseCase) *ContractFilesHandler {
	return &ContractFilesHandler{uc: uc}
}

// Upload 上传合同扫描件
// POST /api/v1/contracts/files/upload
func (h *ContractFilesHandler) Upload(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}

	contractID := c.PostForm("contract_id")
	studentID := c.PostForm("student_id")
	if contractID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "contract_id and student_id required", "data": nil})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	dto, err := h.uc.UploadContractFile(c.Request.Context(), &usecase.UploadContractFileInput{
		ContractID:  contractID,
		StudentID:   studentID,
		UploaderID:  claims.UserID,
		Filename:    header.Filename,
		ContentType: contentType,
		Data:        data,
		Tags:        nil,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": dto})
}

// List 列出某合同下的所有文件
// GET /api/v1/contracts/files/list?contract_id=...
func (h *ContractFilesHandler) List(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	contractID := c.Query("contract_id")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "contract_id required", "data": nil})
		return
	}
	items, err := h.uc.ListContractFiles(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"items": items}})
}

// Download 下载指定文件
// GET /api/v1/contracts/files/download?id=...
func (h *ContractFilesHandler) Download(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "id required", "data": nil})
		return
	}
	data, contentType, err := h.uc.DownloadContractFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment")
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(data)
}


