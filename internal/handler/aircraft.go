package handler

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const (
	maxAircraftAvatarBytes = 5 << 20 // 5MB
)

var allowedAircraftAvatarTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

// AircraftHandler handles aircraft endpoints.
type AircraftHandler struct {
	aircraftRepo      repository.AircraftRepository
	availabilityRepo  repository.AircraftAvailabilityRepository
	maintenanceRepo   repository.AircraftMaintenanceRepository
	maintSync         repository.MaintenanceWindowSyncer
	avatarGridFS      usecase.GridFSBucket
}

// NewAircraftHandler creates an AircraftHandler.
func NewAircraftHandler(
	aircraftRepo repository.AircraftRepository,
	availabilityRepo repository.AircraftAvailabilityRepository,
	maintenanceRepo repository.AircraftMaintenanceRepository,
	maintSync repository.MaintenanceWindowSyncer,
	avatarGridFS usecase.GridFSBucket,
) *AircraftHandler {
	return &AircraftHandler{
		aircraftRepo:     aircraftRepo,
		availabilityRepo: availabilityRepo,
		maintenanceRepo:  maintenanceRepo,
		maintSync:        maintSync,
		avatarGridFS:     avatarGridFS,
	}
}

func enrichAircraftAPI(a *domain.Aircraft) {
	if a == nil {
		return
	}
	if a.AvatarGridFSID != "" {
		a.AvatarURL = "/api/v1/aircraft/avatar?aircraft_id=" + url.QueryEscape(a.ID)
	} else {
		a.AvatarURL = ""
	}
}

func enrichAircraftList(list []*domain.Aircraft) {
	for _, a := range list {
		enrichAircraftAPI(a)
	}
}

// List returns all aircraft.
func (h *AircraftHandler) List(c *gin.Context) {
	list, err := h.aircraftRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	enrichAircraftList(list)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": list},
	})
}

// Create creates a new aircraft.
func (h *AircraftHandler) Create(c *gin.Context) {
	var body struct {
		Code   string `json:"code" binding:"required"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	a := &domain.Aircraft{
		Code:   body.Code,
		Name:   body.Name,
		Status: body.Status,
	}
	if a.Status == "" {
		a.Status = string(domain.AircraftStatusActive)
	}
	if !domain.ValidAircraftStatus(a.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "status 须为 active(现役)、retired(退役)、partnership(合作)", "data": nil})
		return
	}
	if err := h.aircraftRepo.Create(c.Request.Context(), a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	enrichAircraftAPI(a)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": a,
	})
}

// Update POST 更新飞机名称与业务状态。
func (h *AircraftHandler) Update(c *gin.Context) {
	var body struct {
		ID     string  `json:"id" binding:"required"`
		Name   *string `json:"name"`
		Status *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if body.Name == nil && body.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "至少提供 name 或 status 之一", "data": nil})
		return
	}
	if body.Status != nil && !domain.ValidAircraftStatus(*body.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "status 须为 active(现役)、retired(退役)、partnership(合作)", "data": nil})
		return
	}
	if err := h.aircraftRepo.Update(c.Request.Context(), body.ID, body.Name, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	a, err := h.aircraftRepo.GetByID(c.Request.Context(), body.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"updated": true}})
		return
	}
	enrichAircraftAPI(a)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": a})
}

// Delete POST 删除飞机（级联删除可用性、维护窗等；若有头像则从 GridFS 移除）。
func (h *AircraftHandler) Delete(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	a, err := h.aircraftRepo.GetByID(c.Request.Context(), body.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "飞机不存在", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if a.AvatarGridFSID != "" && h.avatarGridFS != nil {
		if delErr := h.avatarGridFS.Delete(c.Request.Context(), a.AvatarGridFSID); delErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除头像文件失败: " + delErr.Error(), "data": nil})
			return
		}
	}
	if err := h.aircraftRepo.Delete(c.Request.Context(), body.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"deleted": true}})
}

// UploadAvatar POST multipart：字段 aircraft_id、file。
func (h *AircraftHandler) UploadAvatar(c *gin.Context) {
	if h.avatarGridFS == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "头像存储未配置", "data": nil})
		return
	}
	aircraftID := c.PostForm("aircraft_id")
	if aircraftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "aircraft_id required", "data": nil})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAircraftAvatarBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if len(data) > maxAircraftAvatarBytes {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "头像超过大小限制（5MB）", "data": nil})
		return
	}

	contentType := strings.TrimSpace(strings.ToLower(header.Header.Get("Content-Type")))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if _, ok := allowedAircraftAvatarTypes[contentType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "仅支持 jpeg/png/webp/gif 图片", "data": nil})
		return
	}

	a, err := h.aircraftRepo.GetByID(c.Request.Context(), aircraftID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "飞机不存在", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}

	prevID := a.AvatarGridFSID
	meta := map[string]interface{}{
		"aircraft_id":  aircraftID,
		"content_type": contentType,
		"kind":         "aircraft_avatar",
	}
	newID, err := h.avatarGridFS.UploadFromStream(c.Request.Context(), header.Filename, data, meta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if err := h.aircraftRepo.SetAvatarGridFSID(c.Request.Context(), aircraftID, newID); err != nil {
		_ = h.avatarGridFS.Delete(c.Request.Context(), newID)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if prevID != "" && prevID != newID {
		_ = h.avatarGridFS.Delete(c.Request.Context(), prevID)
	}

	a, err = h.aircraftRepo.GetByID(c.Request.Context(), aircraftID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{
			"id":         aircraftID,
			"avatar_url": "/api/v1/aircraft/avatar?aircraft_id=" + url.QueryEscape(aircraftID),
		}})
		return
	}
	enrichAircraftAPI(a)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": a})
}

// RemoveAvatar POST JSON：{ "aircraft_id": "..." }，清除头像并删除 GridFS 文件。
func (h *AircraftHandler) RemoveAvatar(c *gin.Context) {
	if h.avatarGridFS == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "头像存储未配置", "data": nil})
		return
	}
	var body struct {
		AircraftID string `json:"aircraft_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	a, err := h.aircraftRepo.GetByID(c.Request.Context(), body.AircraftID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "飞机不存在", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if a.AvatarGridFSID != "" {
		if err := h.avatarGridFS.Delete(c.Request.Context(), a.AvatarGridFSID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
	}
	if err := h.aircraftRepo.SetAvatarGridFSID(c.Request.Context(), body.AircraftID, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	a.AvatarGridFSID = ""
	enrichAircraftAPI(a)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": a})
}

// StreamAvatar GET ?aircraft_id= 输出头像字节流（需 JWT）。
func (h *AircraftHandler) StreamAvatar(c *gin.Context) {
	if h.avatarGridFS == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "头像存储未配置", "data": nil})
		return
	}
	aircraftID := c.Query("aircraft_id")
	if aircraftID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "aircraft_id required", "data": nil})
		return
	}
	a, err := h.aircraftRepo.GetByID(c.Request.Context(), aircraftID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "飞机不存在", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if a.AvatarGridFSID == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未设置头像", "data": nil})
		return
	}
	data, meta, err := h.avatarGridFS.DownloadToStream(c.Request.Context(), a.AvatarGridFSID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "头像文件不存在", "data": nil})
		return
	}
	ct := "application/octet-stream"
	if meta != nil {
		if v, ok := meta["content_type"].(string); ok && v != "" {
			ct = v
		}
	}
	c.Data(http.StatusOK, ct, data)
}

// CreateAvailability configures availability for an aircraft on a given date.
func (h *AircraftHandler) CreateAvailability(c *gin.Context) {
	var body struct {
		AircraftID string `json:"aircraft_id" binding:"required"`
		Date       string `json:"date" binding:"required"`       // YYYY-MM-DD
		StartTime  string `json:"start_time" binding:"required"` // HH:MM
		EndTime    string `json:"end_time" binding:"required"`   // HH:MM
		Status     string `json:"status" binding:"required"`     // available | maintenance | blocked
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	d, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid date", "data": nil})
		return
	}
	start, err := time.Parse("15:04", body.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid start_time", "data": nil})
		return
	}
	end, err := time.Parse("15:04", body.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid end_time", "data": nil})
		return
	}
	if !end.After(start) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "end_time must be after start_time", "data": nil})
		return
	}

	a := &domain.AircraftAvailability{
		AircraftID: body.AircraftID,
		Date:       d,
		StartTime:  start,
		EndTime:    end,
		Status:     domain.AvailabilityStatus(body.Status),
	}
	if err := h.availabilityRepo.Create(c.Request.Context(), a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": a,
	})
}
