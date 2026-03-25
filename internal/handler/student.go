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
	if len(bytes.TrimSpace(raw)) > 0 {
		var body struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
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
	}
	list, err := h.studentRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": list},
	})
}

// AdminList 后台学员列表 POST /api/v1/admin/students/list
func (h *StudentHandler) AdminList(c *gin.Context) {
	var body struct {
		Keyword                string `json:"keyword"`
		ResponsibleInstructorID string `json:"responsible_instructor_id"`
		Page                   int    `json:"page"`
		PageSize               int    `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if body.Page <= 0 {
		body.Page = 1
	}
	if body.PageSize <= 0 || body.PageSize > 100 {
		body.PageSize = 20
	}
	limit := body.PageSize
	offset := (body.Page - 1) * body.PageSize

	list, err := h.studentRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}

	type item struct {
		StudentID               string `json:"student_id"`
		Name                    string `json:"name"`
		StudentNo               string `json:"student_no"`
		Username                string `json:"username"`
		Phone                   string `json:"phone"`
		Email                   string `json:"email"`
		ContractDurationMinutes *int   `json:"contract_duration_minutes,omitempty"`
		FlownMinutes            int    `json:"flown_minutes"`
		RemainingMinutes        *int   `json:"remaining_minutes,omitempty"`
		TrainingProgress        string `json:"training_progress"`
		ResponsibleInstructor   *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"responsible_instructor,omitempty"`
	}

	var out []item
	for _, s := range list {
		it := item{
			StudentID:        s.ID,
			Name:             s.Name,
			StudentNo:        s.StudentNo,
			ContractDurationMinutes: s.ContractDurationMinutes,
			FlownMinutes:           s.FlownMinutes,
			TrainingProgress:       s.TrainingProgress,
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
	if s.UserID != nil {
		if u, err := h.userRepo.GetByID(ctx, *s.UserID); err == nil && u != nil {
			username = u.Username
			phone = u.Phone
			email = u.Email
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
			"id":                     a.ID,
			"training_date":          a.TrainingDate.Format("2006-01-02"),
			"stage":                  a.Stage,
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
					"id":            ct.ID,
					"contract_no":   ct.ContractNo,
					"sign_date":     ct.SignDate,
					"start_date":    ct.StartDate,
					"end_date":      ct.EndDate,
					"total_minutes": ct.TotalMinutes,
					"status":        ct.Status,
				})
			}
		}
	}

	resp := gin.H{
		"basic": gin.H{
			"student_id":               s.ID,
			"name":                     s.Name,
			"student_no":               s.StudentNo,
			"contract_duration_minutes": s.ContractDurationMinutes,
			"flown_minutes":            s.FlownMinutes,
			"remaining_minutes":        remaining,
			"training_progress":        s.TrainingProgress,
		},
		"account": gin.H{
			"username": username,
			"phone":    phone,
			"email":    email,
			"role":     "student",
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
		"training_progress":         s.TrainingProgress,
		"training_progress_json":    trainingProgressJSON,
		"pending_task_count":        pendingCount,
		"latest_archive":            latestArchive,
		"upcoming_training":         upcoming,
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
			"flown_minutes":               s.FlownMinutes,
			"remaining_minutes":           remaining,
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
			"profile":            profile,
			"training":           training,
			"syllabus_progress":  syllabusProgress,
			"entitlements":       entitlements,
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
			"id":            ct.ID,
			"contract_no":   ct.ContractNo,
			"total_minutes": ct.TotalMinutes,
			"price_total":   ct.PriceTotal,
			"currency":      ct.Currency,
			"status":        ct.Status,
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
	var body struct {
		StudentID               *string         `json:"student_id"`
		Name                    string          `json:"name" binding:"required"`
		StudentNo               string          `json:"student_no" binding:"required"`
		Username                string          `json:"username"`
		Phone                   string          `json:"phone"`
		Email                   string          `json:"email"`
		Password                string          `json:"password"`
		ContractDurationMinutes *int            `json:"contract_duration_minutes"`
		FlownMinutes            int             `json:"flown_minutes"`
		TrainingProgress        string          `json:"training_progress"`
		ResponsibleInstructorID *string         `json:"responsible_instructor_id"`
		ContractInfo            json.RawMessage `json:"contract_info"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	ctx := c.Request.Context()

	// 简化：创建路径只在传入 student_id 为空时使用，密码作为明文占位，调用 AuthUseCase 更合理，这里仅演示结构。
	if body.StudentID == nil || *body.StudentID == "" {
		// 若用户名已存在，则直接返回 400，避免违反唯一约束
		if body.Username != "" {
			if _, err := h.userRepo.GetByUsername(ctx, body.Username); err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "username already exists", "data": nil})
				return
			}
		}
		// 创建 user（role=student）
		u := &domain.User{
			Username: body.Username,
			Phone:    body.Phone,
			Email:    body.Email,
			Role:     domain.RoleStudent,
		}
		if err := h.userRepo.Create(ctx, u); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
		s := &domain.Student{
			UserID:                  &u.ID,
			Name:                    body.Name,
			StudentNo:               body.StudentNo,
			ContractDurationMinutes: body.ContractDurationMinutes,
			FlownMinutes:            body.FlownMinutes,
			TrainingProgress:        body.TrainingProgress,
			ResponsibleInstructorID: body.ResponsibleInstructorID,
		}
		if err := mergeContractInfoJSON(&s.ContractInfo, body.ContractInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		if err := h.studentRepo.Create(ctx, s); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": s})
		return
	}

	// 更新已有学员
	s, err := h.studentRepo.GetByID(ctx, *body.StudentID)
	if err != nil || s == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
		return
	}
	s.Name = body.Name
	s.StudentNo = body.StudentNo
	s.ContractDurationMinutes = body.ContractDurationMinutes
	s.FlownMinutes = body.FlownMinutes
	s.TrainingProgress = body.TrainingProgress
	s.ResponsibleInstructorID = body.ResponsibleInstructorID
	if err := mergeContractInfoJSON(&s.ContractInfo, body.ContractInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if err := h.studentRepo.Update(ctx, s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": s})
}

// MyHome 学员主页聚合接口 POST /api/v1/students/my-home
func (h *StudentHandler) MyHome(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	ctx := c.Request.Context()
	// 当前学员
	stu, err := h.studentRepo.GetByUserID(ctx, claims.UserID)
	if err != nil || stu == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
		return
	}

	// profile
	var remaining *int
	if stu.ContractDurationMinutes != nil {
		r := *stu.ContractDurationMinutes - stu.FlownMinutes
		remaining = &r
	}
	profile := gin.H{
		"student_id":               stu.ID,
		"name":                     stu.Name,
		"student_no":               stu.StudentNo,
		"contract_duration_minutes": stu.ContractDurationMinutes,
		"flown_minutes":            stu.FlownMinutes,
		"remaining_minutes":        remaining,
		"training_progress":        stu.TrainingProgress,
	}

	// tasks 预览：复用 taskRepo，简单取所有 pending
	var pendingTasks []*domain.Task
	status := domain.TaskStatusPending
	allTasks, _ := h.taskRepo.ListByStudent(ctx, stu.ID, &status)
	if len(allTasks) > 5 {
		pendingTasks = allTasks[:5]
	} else {
		pendingTasks = allTasks
	}
	tasksBlock := gin.H{
		"items": pendingTasks,
		"total": len(allTasks),
	}

	// upcoming_training：未来 7 天内的 training_plans
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
		if p.StudentID != stu.ID {
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

	// last_training_archive：最近一次档案
	archives, _ := h.archiveRepo.ListByStudent(ctx, stu.ID, 1, 0)
	var lastArchive interface{}
	if len(archives) > 0 {
		a := archives[0]
		lastArchive = gin.H{
			"id":                     a.ID,
			"training_date":          a.TrainingDate.Format("2006-01-02"),
			"stage":                  a.Stage,
			"flight_duration_minutes": a.FlightDurationMinutes,
		}
	}

	// theory_summary：简单按当前用户全部 attempt 统计（略简化）
	// NOTE: 现有 TheoryAttemptRepository 只按文档 + 用户筛选，为简化，这里不做遍历，留空结构或后续扩展。
	theorySummary := gin.H{
		"completed_documents": 0,
		"avg_score":           0,
		"last_attempt_at":     nil,
	}

	// subject_progress 目前通过 syllabus-subjects/list 完成统计，聚合放在前端或后续 UseCase 中。
	subjectProgress := gin.H{
		"completed": 0,
		"total":     0,
	}

	data := gin.H{
		"profile":             profile,
		"tasks":               tasksBlock,
		"upcoming_training":   upcoming,
		"last_training_archive": lastArchive,
		"theory_summary":      theorySummary,
		"subject_progress":    subjectProgress,
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

var errInvalidContractInfoJSON = errors.New("invalid contract_info json")

func mergeContractInfoJSON(dst *[]byte, raw json.RawMessage) error {
	b := bytes.TrimSpace(raw)
	if len(b) == 0 {
		return nil
	}
	if !json.Valid(b) {
		return errInvalidContractInfoJSON
	}
	*dst = append([]byte(nil), b...)
	return nil
}

func isStaffLikeRole(role string) bool {
	switch domain.UserRole(role) {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleMaintenance, domain.RoleInstructor:
		return true
	default:
		return false
	}
}
