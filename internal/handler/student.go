package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/pkg/studentcontract"
	"magua-fly-golang/internal/repository"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// StudentHandler handles student endpoints.
type StudentHandler struct {
	studentRepo  repository.StudentRepository
	userRepo     repository.UserRepository
	taskRepo     repository.TaskRepository
	archiveRepo  repository.TrainingArchiveRepository
	planRepo     repository.TrainingPlanRepository
	instrRepo    repository.InstructorRepository
	aircraftRepo repository.AircraftRepository
	theoryRepo   repository.TheoryAttemptRepository
	contractRepo repository.ContractRepository
	trainingUC   *usecase.TrainingUseCase
	homeUC       *usecase.HomeUseCase
	personnelUC  *usecase.PersonnelUseCase
}

// NewStudentHandler creates a StudentHandler.
func NewStudentHandler(
	studentRepo repository.StudentRepository,
	userRepo repository.UserRepository,
	taskRepo repository.TaskRepository,
	archiveRepo repository.TrainingArchiveRepository,
	planRepo repository.TrainingPlanRepository,
	instrRepo repository.InstructorRepository,
	aircraftRepo repository.AircraftRepository,
	theoryRepo repository.TheoryAttemptRepository,
	contractRepo repository.ContractRepository,
	trainingUC *usecase.TrainingUseCase,
	homeUC *usecase.HomeUseCase,
	personnelUC *usecase.PersonnelUseCase,
) *StudentHandler {
	return &StudentHandler{
		studentRepo:  studentRepo,
		userRepo:     userRepo,
		taskRepo:     taskRepo,
		archiveRepo:  archiveRepo,
		planRepo:     planRepo,
		instrRepo:    instrRepo,
		aircraftRepo: aircraftRepo,
		theoryRepo:   theoryRepo,
		contractRepo: contractRepo,
		trainingUC:   trainingUC,
		homeUC:       homeUC,
		personnelUC:  personnelUC,
	}
}

// List returns a paginated list of students.
// Body: {"limit":20,"offset":0}；空 body 或仅空白时使用默认 limit=20、offset=0。
func (h *StudentHandler) List(c *gin.Context) {
	limit, offset := 20, 0
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	var keyword, responsibleInstructorID, flightTrainingStatus string
	if len(bytes.TrimSpace(raw)) > 0 {
		var body struct {
			Limit                   int    `json:"limit"`
			Offset                  int    `json:"offset"`
			Keyword                 string `json:"keyword"`
			ResponsibleInstructorID string `json:"responsible_instructor_id"`
			FlightTrainingStatus    string `json:"flight_training_status"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		if body.Limit > 0 {
			limit = body.Limit
			if limit > 100 {
				limit = 100
			}
		}
		if body.Offset > 0 {
			offset = body.Offset
		} else if body.Offset < 0 {
			offset = 0
		}
		keyword = body.Keyword
		responsibleInstructorID = body.ResponsibleInstructorID
		flightTrainingStatus = body.FlightTrainingStatus
	}
	if flightTrainingStatus != "" {
		if _, ok := domain.ParseFlightTrainingStatus(flightTrainingStatus); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid flight_training_status", "data": nil})
			return
		}
	}
	onlyActive := true
	if claims := middleware.GetGinClaims(c); claims != nil && domain.UserRole(claims.Role) == domain.RoleAdmin {
		onlyActive = false
	}
	list, err := h.studentRepo.List(c.Request.Context(), repository.StudentListParams{
		Limit:                   limit,
		Offset:                  offset,
		OnlyActiveEmployment:    onlyActive,
		Keyword:                 keyword,
		ResponsibleInstructorID: responsibleInstructorID,
		FlightTrainingStatus:    flightTrainingStatus,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	ctx := c.Request.Context()
	type listItem struct {
		ID                      string  `json:"id"`
		UserID                  *string `json:"user_id,omitempty"`
		Name                    string  `json:"name"`
		StudentNo               string  `json:"student_no"`
		ResidentCity            string  `json:"resident_city,omitempty"`
		Username                string  `json:"username,omitempty"`
		Phone                   string  `json:"phone,omitempty"`
		Email                   string  `json:"email,omitempty"`
		EmploymentStatus        string  `json:"employment_status,omitempty"`
		PersonnelCategory       string  `json:"personnel_category,omitempty"`
		ContractInfo            []byte  `json:"contract_info,omitempty"`
		ContractDurationMinutes *int    `json:"contract_duration_minutes,omitempty"`
		FlownMinutes            int     `json:"flown_minutes"`
		TrainingProgress        string  `json:"training_progress,omitempty"`
		TrainingProgressJSON    []byte  `json:"training_progress_json,omitempty"`
		ResponsibleInstructorID *string `json:"responsible_instructor_id,omitempty"`
		ResponsibleInstructor   *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"responsible_instructor,omitempty"`
		FlightTrainingStatus string    `json:"flight_training_status,omitempty"`
		CreatedAt            time.Time `json:"created_at"`
		UpdatedAt            time.Time `json:"updated_at"`
	}
	out := make([]listItem, 0, len(list))
	for _, s := range list {
		it := listItem{
			ID:                      s.ID,
			UserID:                  s.UserID,
			Name:                    s.Name,
			StudentNo:               s.StudentNo,
			ResidentCity:            s.ResidentCity,
			ContractInfo:            s.ContractInfo,
			ContractDurationMinutes: s.ContractDurationMinutes,
			FlownMinutes:            s.FlownMinutes,
			TrainingProgress:        s.TrainingProgress,
			TrainingProgressJSON:    s.TrainingProgressJSON,
			ResponsibleInstructorID: s.ResponsibleInstructorID,
			FlightTrainingStatus:    string(s.FlightTrainingStatus),
			CreatedAt:               s.CreatedAt,
			UpdatedAt:               s.UpdatedAt,
		}
		if s.UserID != nil && *s.UserID != "" {
			if u, err := h.userRepo.GetByID(ctx, *s.UserID); err == nil && u != nil {
				it.Username = u.Username
				it.Phone = u.Phone
				it.Email = u.Email
				it.EmploymentStatus = string(u.EmploymentStatus)
				it.PersonnelCategory = string(domain.DerivePersonnelCategory(u.Role))
			}
		}
		if s.ResponsibleInstructorID != nil {
			if ins, err := h.instrRepo.GetByID(ctx, *s.ResponsibleInstructorID); err == nil && ins != nil {
				it.ResponsibleInstructor = &struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}{
					ID:   ins.ID,
					Name: ins.Name,
				}
			}
		}
		out = append(out, it)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": out},
	})
}

// AdminList 后台学员列表 POST /api/v1/admin/students/list
func (h *StudentHandler) AdminList(c *gin.Context) {
	var body struct {
		Keyword                 string `json:"keyword"`
		ResponsibleInstructorID string `json:"responsible_instructor_id"`
		FlightTrainingStatus    string `json:"flight_training_status"`
		Page                    int    `json:"page"`
		PageSize                int    `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if body.FlightTrainingStatus != "" {
		if _, ok := domain.ParseFlightTrainingStatus(body.FlightTrainingStatus); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid flight_training_status", "data": nil})
			return
		}
	}
	if body.Page <= 0 {
		body.Page = 1
	}
	if body.PageSize <= 0 || body.PageSize > 100 {
		body.PageSize = 20
	}
	limit := body.PageSize
	offset := (body.Page - 1) * body.PageSize

	list, err := h.studentRepo.List(c.Request.Context(), repository.StudentListParams{
		Limit:                   limit,
		Offset:                  offset,
		OnlyActiveEmployment:    false,
		Keyword:                 body.Keyword,
		ResponsibleInstructorID: body.ResponsibleInstructorID,
		FlightTrainingStatus:    body.FlightTrainingStatus,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}

	type item struct {
		StudentID               string `json:"student_id"`
		Name                    string `json:"name"`
		StudentNo               string `json:"student_no"`
		ResidentCity            string `json:"resident_city,omitempty"`
		Username                string `json:"username"`
		Phone                   string `json:"phone"`
		Email                   string `json:"email"`
		EmploymentStatus        string `json:"employment_status,omitempty"`
		PersonnelCategory       string `json:"personnel_category,omitempty"`
		ContractDurationMinutes *int   `json:"contract_duration_minutes,omitempty"`
		FlownMinutes            int    `json:"flown_minutes"`
		RemainingMinutes        *int   `json:"remaining_minutes,omitempty"`
		TrainingProgress        string `json:"training_progress"`
		FlightTrainingStatus    string `json:"flight_training_status,omitempty"`
		ResponsibleInstructor   *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"responsible_instructor,omitempty"`
	}

	var out []item
	for _, s := range list {
		it := item{
			StudentID:               s.ID,
			Name:                    s.Name,
			StudentNo:               s.StudentNo,
			ResidentCity:            s.ResidentCity,
			ContractDurationMinutes: s.ContractDurationMinutes,
			FlownMinutes:            s.FlownMinutes,
			TrainingProgress:        s.TrainingProgress,
			FlightTrainingStatus:    string(s.FlightTrainingStatus),
		}
		if s.ContractDurationMinutes != nil {
			rem := *s.ContractDurationMinutes - s.FlownMinutes
			it.RemainingMinutes = &rem
		}
		if s.UserID != nil {
			if u, err := h.userRepo.GetByID(c.Request.Context(), *s.UserID); err == nil && u != nil {
				it.Username = u.Username
				it.Phone = u.Phone
				it.Email = u.Email
				it.EmploymentStatus = string(u.EmploymentStatus)
				it.PersonnelCategory = string(domain.DerivePersonnelCategory(u.Role))
			}
		}
		if s.ResponsibleInstructorID != nil {
			if ins, err := h.instrRepo.GetByID(c.Request.Context(), *s.ResponsibleInstructorID); err == nil && ins != nil {
				it.ResponsibleInstructor = &struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}{
					ID:   ins.ID,
					Name: ins.Name,
				}
			}
		}
		out = append(out, it)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": out, "total": len(out)},
	})
}

// AdminDetail 后台学员详情 POST /api/v1/admin/students/detail
func (h *StudentHandler) AdminDetail(c *gin.Context) {
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	ctx := c.Request.Context()
	s, err := h.studentRepo.GetByID(ctx, body.StudentID)
	if err != nil || s == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
		return
	}

	var username, phone, email string
	var employmentStatus, personnelCategory string
	if s.UserID != nil {
		if u, err := h.userRepo.GetByID(ctx, *s.UserID); err == nil && u != nil {
			username = u.Username
			phone = u.Phone
			email = u.Email
			employmentStatus = string(u.EmploymentStatus)
			personnelCategory = string(domain.DerivePersonnelCategory(u.Role))
		}
	}

	var responsibleInstructor *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if s.ResponsibleInstructorID != nil {
		if ins, err := h.instrRepo.GetByID(ctx, *s.ResponsibleInstructorID); err == nil && ins != nil {
			responsibleInstructor = &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   ins.ID,
				Name: ins.Name,
			}
		}
	}

	archives, err := h.archiveRepo.ListByStudent(ctx, s.ID, 1, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	var latestArchive interface{}
	if len(archives) > 0 {
		a := archives[0]
		latestArchive = gin.H{
			"id":                      a.ID,
			"training_date":           a.TrainingDate.Format("2006-01-02"),
			"stage":                   a.Stage,
			"flight_duration_minutes": a.FlightDurationMinutes,
		}
	}

	// 待办任务数量（直接按 pending 状态统计）
	statusPending := domain.TaskStatusPending
	tasks, _ := h.taskRepo.ListByStudent(ctx, s.ID, &statusPending)
	pendingCount := len(tasks)

	var remaining *int
	if s.ContractDurationMinutes != nil {
		r := *s.ContractDurationMinutes - s.FlownMinutes
		remaining = &r
	}

	// 合同列表摘要（按创建时间倒序）
	var contracts []gin.H
	if h.contractRepo != nil {
		if list, err := h.contractRepo.ListByStudent(ctx, s.ID); err == nil {
			for _, ct := range list {
				contracts = append(contracts, gin.H{
					"id":                ct.ID,
					"contract_no":       ct.ContractNo,
					"sign_date":         ct.SignDate,
					"start_date":        ct.StartDate,
					"end_date":          ct.EndDate,
					"total_minutes":     ct.TotalMinutes,
					"status":            ct.Status,
					"payment_mode":      ct.PaymentMode,
					"payment_schedule":  ct.PaymentSchedule,
				})
			}
		}
	}

	resp := gin.H{
		"basic": gin.H{
			"student_id":                s.ID,
			"name":                      s.Name,
			"student_no":                s.StudentNo,
			"resident_city":             s.ResidentCity,
			"contract_duration_minutes": s.ContractDurationMinutes,
			"flown_minutes":             s.FlownMinutes,
			"remaining_minutes":         remaining,
			"training_progress":         s.TrainingProgress,
			"flight_training_status":    string(s.FlightTrainingStatus),
		},
		"account": gin.H{
			"username":           username,
			"phone":              phone,
			"email":              email,
			"role":               "student",
			"employment_status":  employmentStatus,
			"personnel_category": personnelCategory,
		},
		"responsible_instructor": responsibleInstructor,
		"latest_archive":         latestArchive,
		"pending_task_count":     pendingCount,
		"contracts":              contracts,
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": resp})
}

// StaffOverview 职员/教员端学员概要（训练状态、进度、权益）POST /api/v1/students/staff-overview
func (h *StudentHandler) StaffOverview(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if !isStaffLikeRole(claims.Role) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden: student cannot access this resource", "data": nil})
		return
	}
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	ctx := c.Request.Context()
	s, err := h.studentRepo.GetByID(ctx, body.StudentID)
	if err != nil || s == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
		return
	}

	profile := gin.H{
		"student_id": s.ID,
		"name":       s.Name,
		"student_no": s.StudentNo,
	}

	archives, err := h.archiveRepo.ListByStudent(ctx, s.ID, 1, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	var latestArchive interface{}
	if len(archives) > 0 {
		a := archives[0]
		latestArchive = gin.H{
			"id":                      a.ID,
			"training_date":           a.TrainingDate.Format("2006-01-02"),
			"stage":                   a.Stage,
			"flight_duration_minutes": a.FlightDurationMinutes,
		}
	}

	statusPending := domain.TaskStatusPending
	tasks, _ := h.taskRepo.ListByStudent(ctx, s.ID, &statusPending)
	pendingCount := len(tasks)

	var trainingProgressJSON interface{}
	if len(s.TrainingProgressJSON) > 0 {
		trainingProgressJSON = json.RawMessage(s.TrainingProgressJSON)
	}

	now := time.Now()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, 7).Format("2006-01-02")
	plans, _ := h.planRepo.ListByDate(ctx, from, to)
	type upcomingItem struct {
		TrainingDate   string `json:"training_date"`
		StartTime      string `json:"start_time"`
		EndTime        string `json:"end_time"`
		InstructorName string `json:"instructor_name"`
		AircraftName   string `json:"aircraft_name"`
	}
	var upcoming []upcomingItem
	for _, p := range plans {
		if p.StudentID != s.ID {
			continue
		}
		var insName, acName string
		if ins, err := h.instrRepo.GetByID(ctx, p.InstructorID); err == nil && ins != nil {
			insName = ins.Name
		}
		if ac, err := h.aircraftRepo.GetByID(ctx, p.AircraftID); err == nil && ac != nil {
			acName = ac.Name
		}
		upcoming = append(upcoming, upcomingItem{
			TrainingDate:   p.PlanDate.Format("2006-01-02"),
			StartTime:      p.StartTime.Format("15:04"),
			EndTime:        p.EndTime.Format("15:04"),
			InstructorName: insName,
			AircraftName:   acName,
		})
	}

	training := gin.H{
		"training_progress":      s.TrainingProgress,
		"training_progress_json": trainingProgressJSON,
		"pending_task_count":     pendingCount,
		"latest_archive":         latestArchive,
		"upcoming_training":      upcoming,
	}

	syllabusProgress := gin.H{"completed": 0, "total": 0}
	if h.trainingUC != nil {
		sid := s.ID
		subjects, err := h.trainingUC.ListSyllabusSubjects(ctx, &sid, nil, nil)
		if err == nil {
			completed := 0
			for _, sub := range subjects {
				if sub.IsGraduated {
					completed++
				}
			}
			syllabusProgress = gin.H{"completed": completed, "total": len(subjects)}
		}
	}

	var remaining *int
	if s.ContractDurationMinutes != nil {
		r := *s.ContractDurationMinutes - s.FlownMinutes
		remaining = &r
	}
	ent := studentcontract.ParseEntitlementsFromContractInfo(s.ContractInfo)
	entitlements := gin.H{
		"flight": gin.H{
			"contract_duration_minutes": s.ContractDurationMinutes,
			"flown_minutes":             s.FlownMinutes,
			"remaining_minutes":         remaining,
		},
		"learning_points": ent.LearningPoints,
		"guest_experience": gin.H{
			"quota":     ent.GuestQuota,
			"used":      ent.GuestUsed,
			"remaining": ent.GuestRemaining,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"profile":           profile,
			"training":          training,
			"syllabus_progress": syllabusProgress,
			"entitlements":      entitlements,
		},
	})
}

// AdminContractsListByStudent 按学员查询合同列表 POST /api/v1/admin/contracts/list-by-student
func (h *StudentHandler) AdminContractsListByStudent(c *gin.Context) {
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	ctx := c.Request.Context()
	s, err := h.studentRepo.GetByID(ctx, body.StudentID)
	if err != nil || s == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
		return
	}
	list, err := h.contractRepo.ListByStudent(ctx, body.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	var items []gin.H
	for _, ct := range list {
		item := gin.H{
			"id":               ct.ID,
			"contract_no":      ct.ContractNo,
			"total_minutes":      ct.TotalMinutes,
			"price_total":        ct.PriceTotal,
			"currency":           ct.Currency,
			"status":             ct.Status,
			"payment_mode":       ct.PaymentMode,
			"payment_schedule":   ct.PaymentSchedule,
		}
		if ct.SignDate != nil {
			item["sign_date"] = ct.SignDate.Format("2006-01-02")
		} else {
			item["sign_date"] = nil
		}
		if ct.StartDate != nil {
			item["start_date"] = ct.StartDate.Format("2006-01-02")
		} else {
			item["start_date"] = nil
		}
		if ct.EndDate != nil {
			item["end_date"] = ct.EndDate.Format("2006-01-02")
		} else {
			item["end_date"] = nil
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": items},
	})
}

// AdminSave 后台学员创建 / 更新 POST /api/v1/admin/students/save
func (h *StudentHandler) AdminSave(c *gin.Context) {
	var body usecase.SavePersonnelRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	body.Kind = "student"
	if h.personnelUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "personnel save not configured", "data": nil})
		return
	}
	res, err := h.personnelUC.SavePersonnel(c.Request.Context(), &body)
	if err != nil {
		writePersonnelSaveErr(c, err)
		return
	}
	if res == nil || res.Student == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "empty result", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": res.Student})
}

// MyHome 学员主页聚合接口 POST /api/v1/students/my-home
func (h *StudentHandler) MyHome(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if domain.UserRole(claims.Role) != domain.RoleStudent {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	if h.homeUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "home aggregate not configured", "data": nil})
		return
	}
	var body usecase.StudentHomeRequest
	_ = c.ShouldBindJSON(&body)
	data, _, err := h.homeUC.StudentHome(c.Request.Context(), claims.UserID, &body)
	if err != nil {
		if errors.Is(err, usecase.ErrStudentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func isStaffLikeRole(role string) bool {
	switch domain.UserRole(role) {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleMaintenance, domain.RoleInstructor:
		return true
	default:
		return false
	}
}
