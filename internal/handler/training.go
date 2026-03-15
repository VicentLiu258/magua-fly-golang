package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// TrainingHandler 训练记录、导入、视图、档案、科目
type TrainingHandler struct {
	uc *usecase.TrainingUseCase
}

// NewTrainingHandler 创建 TrainingHandler
func NewTrainingHandler(uc *usecase.TrainingUseCase) *TrainingHandler {
	return &TrainingHandler{uc: uc}
}

func respOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func respErr(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"code": code, "msg": msg, "data": nil})
}

// CreateRecord POST /api/v1/training-records
func (h *TrainingHandler) CreateRecord(c *gin.Context) {
	var body struct {
		StudentID              string  `json:"student_id" binding:"required"`
		InstructorID           string  `json:"instructor_id" binding:"required"`
		AircraftID             string  `json:"aircraft_id" binding:"required"`
		TrainingDate           string  `json:"training_date" binding:"required"`
		EngineOnTime           *string `json:"engine_on_time"`
		EngineOffTime          *string `json:"engine_off_time"`
		FlightDurationMinutes  *int   `json:"flight_duration_minutes"`
		IsLocal                *bool  `json:"is_local"`
		PlannedDurationMinutes *int   `json:"planned_duration_minutes"`
		TheoryDurationMinutes  *int   `json:"theory_duration_minutes"`
		SimDurationMinutes     *int   `json:"sim_duration_minutes"`
		Landings               *int   `json:"landings"`
		SyllabusOutline        string `json:"syllabus_outline"`
		ArchiveDisplayNo       string `json:"archive_display_no"`
		Status                 string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	date, err := time.Parse("2006-01-02", body.TrainingDate)
	if err != nil {
		respErr(c, http.StatusBadRequest, "invalid training_date")
		return
	}
	rec := &domain.TrainingRecord{
		StudentID:              body.StudentID,
		InstructorID:           body.InstructorID,
		AircraftID:             body.AircraftID,
		TrainingDate:           date,
		FlightDurationMinutes:  body.FlightDurationMinutes,
		PlannedDurationMinutes: body.PlannedDurationMinutes,
		TheoryDurationMinutes:  body.TheoryDurationMinutes,
		SimDurationMinutes:     body.SimDurationMinutes,
		Landings:               body.Landings,
		SyllabusOutline:        body.SyllabusOutline,
		ArchiveDisplayNo:       body.ArchiveDisplayNo,
		ImportSource:           domain.ImportManual,
		Status:                 domain.RecordStatusNormal,
	}
	if body.IsLocal != nil {
		rec.IsLocal = *body.IsLocal
	} else {
		rec.IsLocal = true
	}
	if body.EngineOnTime != nil {
		t, _ := time.Parse(time.RFC3339, *body.EngineOnTime)
		rec.EngineOnTime = &t
	}
	if body.EngineOffTime != nil {
		t, _ := time.Parse(time.RFC3339, *body.EngineOffTime)
		rec.EngineOffTime = &t
	}
	if body.Status != "" {
		rec.Status = domain.TrainingRecordStatus(body.Status)
	}
	if err := h.uc.CreateRecord(c.Request.Context(), rec); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, rec)
}

// RecordDetail POST /api/v1/training-records/detail
func (h *TrainingHandler) RecordDetail(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	rec, err := h.uc.GetRecord(c.Request.Context(), body.ID)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	if rec == nil {
		respErr(c, http.StatusNotFound, "not found")
		return
	}
	respOK(c, rec)
}

// UpdateRecord POST /api/v1/training-records/update
func (h *TrainingHandler) UpdateRecord(c *gin.Context) {
	var body struct {
		ID                     string  `json:"id" binding:"required"`
		StudentID              *string `json:"student_id"`
		InstructorID           *string `json:"instructor_id"`
		AircraftID             *string `json:"aircraft_id"`
		TrainingDate           *string `json:"training_date"`
		EngineOnTime           *string `json:"engine_on_time"`
		EngineOffTime          *string `json:"engine_off_time"`
		FlightDurationMinutes  *int    `json:"flight_duration_minutes"`
		IsLocal                *bool   `json:"is_local"`
		PlannedDurationMinutes *int    `json:"planned_duration_minutes"`
		Status                 *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	rec, err := h.uc.GetRecord(c.Request.Context(), body.ID)
	if err != nil || rec == nil {
		respErr(c, http.StatusNotFound, "not found")
		return
	}
	if body.StudentID != nil {
		rec.StudentID = *body.StudentID
	}
	if body.InstructorID != nil {
		rec.InstructorID = *body.InstructorID
	}
	if body.AircraftID != nil {
		rec.AircraftID = *body.AircraftID
	}
	if body.TrainingDate != nil {
		t, _ := time.Parse("2006-01-02", *body.TrainingDate)
		rec.TrainingDate = t
	}
	if body.EngineOnTime != nil {
		t, _ := time.Parse(time.RFC3339, *body.EngineOnTime)
		rec.EngineOnTime = &t
	}
	if body.EngineOffTime != nil {
		t, _ := time.Parse(time.RFC3339, *body.EngineOffTime)
		rec.EngineOffTime = &t
	}
	if body.FlightDurationMinutes != nil {
		rec.FlightDurationMinutes = body.FlightDurationMinutes
	}
	if body.IsLocal != nil {
		rec.IsLocal = *body.IsLocal
	}
	if body.PlannedDurationMinutes != nil {
		rec.PlannedDurationMinutes = body.PlannedDurationMinutes
	}
	if body.Status != nil {
		rec.Status = domain.TrainingRecordStatus(*body.Status)
	}
	if err := h.uc.UpdateRecord(c.Request.Context(), rec); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, rec)
}

// DeleteRecord POST /api/v1/training-records/delete
func (h *TrainingHandler) DeleteRecord(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.uc.DeleteRecord(c.Request.Context(), body.ID); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"status": "deleted"})
}

// ListRecords POST /api/v1/training-records/list
func (h *TrainingHandler) ListRecords(c *gin.Context) {
	var body struct {
		ViewID       *string `json:"view_id"`
		StudentID    *string `json:"student_id"`
		InstructorID *string `json:"instructor_id"`
		DateFrom     *string `json:"date_from"`
		DateTo       *string `json:"date_to"`
		Status       *string `json:"status"`
		ImportSource *string `json:"import_source"`
		SortField    string  `json:"sort_field"`
		SortAsc      bool    `json:"sort_asc"`
		Limit        int     `json:"limit"`
		Offset       int     `json:"offset"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		// body 可为空
		_ = err
	}
	filter := &domain.TrainingRecordListFilter{}
	if body.StudentID != nil {
		filter.StudentID = body.StudentID
	}
	if body.InstructorID != nil {
		filter.InstructorID = body.InstructorID
	}
	if body.DateFrom != nil {
		t, err := time.Parse("2006-01-02", *body.DateFrom)
		if err == nil {
			filter.DateFrom = &t
		}
	}
	if body.DateTo != nil {
		t, err := time.Parse("2006-01-02", *body.DateTo)
		if err == nil {
			filter.DateTo = &t
		}
	}
	if body.Status != nil {
		s := domain.TrainingRecordStatus(*body.Status)
		filter.Status = &s
	}
	if body.ImportSource != nil {
		s := domain.ImportSource(*body.ImportSource)
		filter.ImportSource = &s
	}
	sort := (*domain.TrainingRecordListSort)(nil)
	if body.SortField != "" {
		sort = &domain.TrainingRecordListSort{Field: body.SortField, Asc: body.SortAsc}
	}
	if body.Limit <= 0 {
		body.Limit = 20
	}
	// TODO: 若 body.ViewID != nil 可从视图加载 filter/sort
	result, err := h.uc.ListRecords(c.Request.Context(), filter, sort, body.Limit, body.Offset)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, result)
}

// ImportRecords POST /api/v1/training-records/import
func (h *TrainingHandler) ImportRecords(c *gin.Context) {
	var body struct {
		BatchID string                   `json:"batch_id"`
		Rows    []map[string]interface{} `json:"rows" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.BatchID == "" {
		body.BatchID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	var rows []domain.RawFlightImport
	for _, r := range body.Rows {
		row := domain.RawFlightImport{IsLocal: true}
		if v, ok := r["engine_on_time"].(string); ok {
			t, _ := time.Parse(time.RFC3339, v)
			row.EngineOnTime = &t
		}
		if v, ok := r["engine_off_time"].(string); ok {
			t, _ := time.Parse(time.RFC3339, v)
			row.EngineOffTime = &t
		}
		if v, ok := r["flight_duration_minutes"].(float64); ok {
			m := int(v)
			row.FlightDurationMinutes = &m
		}
		if v, ok := r["is_local"].(bool); ok {
			row.IsLocal = v
		}
		rawJSON, _ := json.Marshal(r)
		row.RawData = rawJSON
		rows = append(rows, row)
	}
	created, err := h.uc.ImportRaw(c.Request.Context(), body.BatchID, rows)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"batch_id": body.BatchID, "items": created})
}

// ClaimImport POST /api/v1/training-records/claim
func (h *TrainingHandler) ClaimImport(c *gin.Context) {
	var body struct {
		RawImportID   string  `json:"raw_import_id" binding:"required"`
		StudentID     string  `json:"student_id" binding:"required"`
		InstructorID  string  `json:"instructor_id" binding:"required"`
		AircraftID    string  `json:"aircraft_id" binding:"required"`
		TrainingDate  *string `json:"training_date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	var date *time.Time
	if body.TrainingDate != nil {
		t, err := time.Parse("2006-01-02", *body.TrainingDate)
		if err == nil {
			date = &t
		}
	}
	rec, err := h.uc.ClaimImport(c.Request.Context(), body.RawImportID, body.StudentID, body.InstructorID, body.AircraftID, date)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	if rec == nil {
		respErr(c, http.StatusBadRequest, "already claimed")
		return
	}
	respOK(c, rec)
}

// ListPendingImports POST /api/v1/training-records/import/pending
func (h *TrainingHandler) ListPendingImports(c *gin.Context) {
	var body struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}
	c.ShouldBindJSON(&body)
	if body.Limit <= 0 {
		body.Limit = 50
	}
	list, err := h.uc.ListPendingImports(c.Request.Context(), body.Limit, body.Offset)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"items": list})
}

// RecordStats POST /api/v1/training-records/stats
func (h *TrainingHandler) RecordStats(c *gin.Context) {
	var body struct {
		DateFrom   *string  `json:"date_from"`
		DateTo     *string  `json:"date_to"`
		Dimensions []string `json:"dimensions"`
	}
	c.ShouldBindJSON(&body)
	var from, to *time.Time
	if body.DateFrom != nil {
		t, _ := time.Parse("2006-01-02", *body.DateFrom)
		from = &t
	}
	if body.DateTo != nil {
		t, _ := time.Parse("2006-01-02", *body.DateTo)
		to = &t
	}
	result, err := h.uc.Stats(c.Request.Context(), from, to, body.Dimensions)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, result)
}

// DashboardSummary POST /api/v1/dashboard/summary
func (h *TrainingHandler) DashboardSummary(c *gin.Context) {
	result, err := h.uc.DashboardSummary(c.Request.Context())
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, result)
}

// CreateView POST /api/v1/training-views
func (h *TrainingHandler) CreateView(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		respErr(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		Name       string          `json:"name" binding:"required"`
		FilterJSON json.RawMessage `json:"filter_json"`
		SortJSON   json.RawMessage `json:"sort_json"`
		GroupBy    string          `json:"group_by"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	v := &domain.TrainingRecordView{
		UserID:     claims.UserID,
		Name:       body.Name,
		FilterJSON: body.FilterJSON,
		SortJSON:   body.SortJSON,
		GroupBy:    body.GroupBy,
	}
	if err := h.uc.CreateView(c.Request.Context(), v); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, v)
}

// ListViews POST /api/v1/training-views/list
func (h *TrainingHandler) ListViews(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		respErr(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.uc.ListViews(c.Request.Context(), claims.UserID)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"items": list})
}

// UpdateView POST /api/v1/training-views/update
func (h *TrainingHandler) UpdateView(c *gin.Context) {
	var body struct {
		ID         string          `json:"id" binding:"required"`
		Name       string          `json:"name"`
		FilterJSON json.RawMessage `json:"filter_json"`
		SortJSON   json.RawMessage `json:"sort_json"`
		GroupBy    string          `json:"group_by"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	v, err := h.uc.GetView(c.Request.Context(), body.ID)
	if err != nil || v == nil {
		respErr(c, http.StatusNotFound, "not found")
		return
	}
	if body.Name != "" {
		v.Name = body.Name
	}
	if body.FilterJSON != nil {
		v.FilterJSON = body.FilterJSON
	}
	if body.SortJSON != nil {
		v.SortJSON = body.SortJSON
	}
	if body.GroupBy != "" {
		v.GroupBy = body.GroupBy
	}
	if err := h.uc.UpdateView(c.Request.Context(), v); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, v)
}

// DeleteView POST /api/v1/training-views/delete
func (h *TrainingHandler) DeleteView(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.uc.DeleteView(c.Request.Context(), body.ID); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"status": "deleted"})
}

// ListArchivesByStudent POST /api/v1/students/training-archives/list
func (h *TrainingHandler) ListArchivesByStudent(c *gin.Context) {
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
		Limit     int    `json:"limit"`
		Offset    int    `json:"offset"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	if body.Limit <= 0 {
		body.Limit = 20
	}
	list, err := h.uc.ListArchivesByStudent(c.Request.Context(), body.StudentID, body.Limit, body.Offset)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"items": list})
}

// ArchiveDetail POST /api/v1/training-archives/detail
func (h *TrainingHandler) ArchiveDetail(c *gin.Context) {
	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	a, err := h.uc.GetArchive(c.Request.Context(), body.ID)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	if a == nil {
		respErr(c, http.StatusNotFound, "not found")
		return
	}
	respOK(c, a)
}

// createArchiveBody 用于接收 JSON，training_date 支持 "2006-01-02" 格式
type createArchiveBody struct {
	TrainingRecordID              string  `json:"training_record_id" binding:"required"`
	StudentID                     string  `json:"student_id" binding:"required"`
	InstructorID                  string  `json:"instructor_id" binding:"required"`
	TrainingDate                  string  `json:"training_date" binding:"required"`
	Stage                         string  `json:"stage"`
	PlannedDurationMinutes        *int    `json:"planned_duration_minutes"`
	TotalExperienceMinutes        *int    `json:"total_experience_minutes"`
	RemainingMinutes              *int    `json:"remaining_minutes"`
	EngineOnTime                  *string `json:"engine_on_time"`
	EngineOffTime                 *string `json:"engine_off_time"`
	FlightDurationMinutes         *int    `json:"flight_duration_minutes"`
	ProgressNote                  string  `json:"progress_note"`
	InstructorReview              string  `json:"instructor_review"`
	StudentSelfReview             string  `json:"student_self_review"`
	TaskAssignment                string  `json:"task_assignment"`
}

func parseArchiveBodyToDomain(b *createArchiveBody) (*domain.TrainingArchive, error) {
	date, err := time.Parse("2006-01-02", b.TrainingDate)
	if err != nil {
		return nil, err
	}
	a := &domain.TrainingArchive{
		TrainingRecordID:       b.TrainingRecordID,
		StudentID:              b.StudentID,
		InstructorID:           b.InstructorID,
		TrainingDate:           date,
		Stage:                  b.Stage,
		PlannedDurationMinutes: b.PlannedDurationMinutes,
		TotalExperienceMinutes: b.TotalExperienceMinutes,
		RemainingMinutes:       b.RemainingMinutes,
		FlightDurationMinutes:  b.FlightDurationMinutes,
		ProgressNote:           b.ProgressNote,
		InstructorReview:       b.InstructorReview,
		StudentSelfReview:      b.StudentSelfReview,
		TaskAssignment:         b.TaskAssignment,
	}
	if b.EngineOnTime != nil {
		t, err := time.Parse(time.RFC3339, *b.EngineOnTime)
		if err == nil {
			a.EngineOnTime = &t
		}
	}
	if b.EngineOffTime != nil {
		t, err := time.Parse(time.RFC3339, *b.EngineOffTime)
		if err == nil {
			a.EngineOffTime = &t
		}
	}
	return a, nil
}

// CreateArchive POST /api/v1/training-archives
func (h *TrainingHandler) CreateArchive(c *gin.Context) {
	var body createArchiveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	a, err := parseArchiveBodyToDomain(&body)
	if err != nil {
		respErr(c, http.StatusBadRequest, "invalid training_date, use YYYY-MM-DD")
		return
	}
	if err := h.uc.CreateArchive(c.Request.Context(), a); err != nil {
		if errors.Is(err, usecase.ErrTrainingRecordNotFound) {
			respErr(c, http.StatusBadRequest, err.Error())
			return
		}
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, a)
}

// UpdateArchive POST /api/v1/training-archives/update
func (h *TrainingHandler) UpdateArchive(c *gin.Context) {
	var body struct {
		ID                             string   `json:"id" binding:"required"`
		TrainingRecordID               *string  `json:"training_record_id"`
		StudentID                      *string  `json:"student_id"`
		InstructorID                   *string  `json:"instructor_id"`
		TrainingDate                   *string  `json:"training_date"`
		Stage                          *string  `json:"stage"`
		PlannedDurationMinutes         *int     `json:"planned_duration_minutes"`
		TotalExperienceMinutes         *int     `json:"total_experience_minutes"`
		RemainingMinutes               *int     `json:"remaining_minutes"`
		EngineOnTime                   *string  `json:"engine_on_time"`
		EngineOffTime                  *string  `json:"engine_off_time"`
		FlightDurationMinutes         *int     `json:"flight_duration_minutes"`
		ProgressNote                   *string  `json:"progress_note"`
		InstructorReview               *string  `json:"instructor_review"`
		StudentSelfReview             *string  `json:"student_self_review"`
		TaskAssignment                 *string  `json:"task_assignment"`
		StudentSignedAt                *string  `json:"student_signed_at"`
		ResponsibleInstructorSignedAt  *string  `json:"responsible_instructor_signed_at"`
		ChiefInstructorSignedAt        *string  `json:"chief_instructor_signed_at"`
		ResponsibleInstructorID        *string  `json:"responsible_instructor_id"`
		ChiefInstructorID              *string  `json:"chief_instructor_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	a, err := h.uc.GetArchive(c.Request.Context(), body.ID)
	if err != nil || a == nil {
		respErr(c, http.StatusNotFound, "not found")
		return
	}
	if body.TrainingRecordID != nil {
		a.TrainingRecordID = *body.TrainingRecordID
	}
	if body.StudentID != nil {
		a.StudentID = *body.StudentID
	}
	if body.InstructorID != nil {
		a.InstructorID = *body.InstructorID
	}
	if body.TrainingDate != nil {
		date, err := time.Parse("2006-01-02", *body.TrainingDate)
		if err != nil {
			respErr(c, http.StatusBadRequest, "invalid training_date, use YYYY-MM-DD")
			return
		}
		a.TrainingDate = date
	}
	if body.Stage != nil {
		a.Stage = *body.Stage
	}
	if body.PlannedDurationMinutes != nil {
		a.PlannedDurationMinutes = body.PlannedDurationMinutes
	}
	if body.TotalExperienceMinutes != nil {
		a.TotalExperienceMinutes = body.TotalExperienceMinutes
	}
	if body.RemainingMinutes != nil {
		a.RemainingMinutes = body.RemainingMinutes
	}
	if body.EngineOnTime != nil {
		t, _ := time.Parse(time.RFC3339, *body.EngineOnTime)
		a.EngineOnTime = &t
	}
	if body.EngineOffTime != nil {
		t, _ := time.Parse(time.RFC3339, *body.EngineOffTime)
		a.EngineOffTime = &t
	}
	if body.FlightDurationMinutes != nil {
		a.FlightDurationMinutes = body.FlightDurationMinutes
	}
	if body.ProgressNote != nil {
		a.ProgressNote = *body.ProgressNote
	}
	if body.InstructorReview != nil {
		a.InstructorReview = *body.InstructorReview
	}
	if body.StudentSelfReview != nil {
		a.StudentSelfReview = *body.StudentSelfReview
	}
	if body.TaskAssignment != nil {
		a.TaskAssignment = *body.TaskAssignment
	}
	if body.ResponsibleInstructorID != nil {
		a.ResponsibleInstructorID = body.ResponsibleInstructorID
	}
	if body.ChiefInstructorID != nil {
		a.ChiefInstructorID = body.ChiefInstructorID
	}
	if err := h.uc.UpdateArchive(c.Request.Context(), a); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, a)
}

// SignArchive POST /api/v1/training-archives/sign
func (h *TrainingHandler) SignArchive(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		respErr(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		ID   string `json:"id" binding:"required"`
		Role string `json:"role" binding:"required"` // student | responsible_instructor | chief_instructor
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	// signerID: 学员签字用 user_id，教员签字用 instructor_id（此处简化为 user_id，可后续按角色解析 instructor_id）
	if err := h.uc.SignArchive(c.Request.Context(), body.ID, body.Role, claims.UserID); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"status": "signed"})
}

// ListSyllabusSubjects POST /api/v1/syllabus-subjects/list
func (h *TrainingHandler) ListSyllabusSubjects(c *gin.Context) {
	var body struct {
		StudentID *string `json:"student_id"`
	}
	c.ShouldBindJSON(&body)
	list, err := h.uc.ListSyllabusSubjects(c.Request.Context(), body.StudentID)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"items": list})
}

// ListArchiveSubjects POST /api/v1/training-archives/subjects/list
func (h *TrainingHandler) ListArchiveSubjects(c *gin.Context) {
	var body struct {
		TrainingArchiveID string `json:"training_archive_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	list, err := h.uc.ListArchiveSubjects(c.Request.Context(), body.TrainingArchiveID)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"items": list})
}

// UpdateArchiveSubjects POST /api/v1/training-archives/subjects/update
func (h *TrainingHandler) UpdateArchiveSubjects(c *gin.Context) {
	var body struct {
		TrainingArchiveID string                            `json:"training_archive_id" binding:"required"`
		Subjects          []*domain.TrainingArchiveSubject `json:"subjects"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.uc.UpdateArchiveSubjects(c.Request.Context(), body.TrainingArchiveID, body.Subjects); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"status": "ok"})
}

// ListNextSubjects POST /api/v1/training-archives/next-subjects/list
func (h *TrainingHandler) ListNextSubjects(c *gin.Context) {
	var body struct {
		TrainingArchiveID string `json:"training_archive_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	list, err := h.uc.ListNextSubjects(c.Request.Context(), body.TrainingArchiveID)
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"items": list})
}

// UpdateNextSubjects POST /api/v1/training-archives/next-subjects/update
func (h *TrainingHandler) UpdateNextSubjects(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		respErr(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		TrainingArchiveID  string   `json:"training_archive_id" binding:"required"`
		SyllabusSubjectIDs []string `json:"syllabus_subject_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respErr(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.uc.UpdateNextSubjects(c.Request.Context(), body.TrainingArchiveID, claims.UserID, body.SyllabusSubjectIDs); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"status": "ok"})
}

// ConfigTrainingRecordFields POST /api/v1/config/training-record-fields
func (h *TrainingHandler) ConfigTrainingRecordFields(c *gin.Context) {
	cfg, err := usecase.TrainingRecordFieldsConfig()
	if err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	var out []map[string]interface{}
	if err := jsonUnmarshal(cfg, &out); err != nil {
		respErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	respOK(c, gin.H{"fields": out})
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
