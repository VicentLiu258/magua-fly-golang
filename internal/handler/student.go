package handler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

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
	trainingUC       *usecase.TrainingUseCase
	homeUC           *usecase.HomeUseCase
	personnelUC      *usecase.PersonnelUseCase
	studentImportUC  *usecase.StudentImportUseCase
	trainingSummaryUC *usecase.StudentTrainingSummaryUseCase
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
	studentImportUC *usecase.StudentImportUseCase,
	trainingSummaryUC *usecase.StudentTrainingSummaryUseCase,
) *StudentHandler {
	return &StudentHandler{
		studentRepo:       studentRepo,
		userRepo:          userRepo,
		taskRepo:          taskRepo,
		archiveRepo:       archiveRepo,
		planRepo:          planRepo,
		instrRepo:         instrRepo,
		aircraftRepo:      aircraftRepo,
		theoryRepo:        theoryRepo,
		contractRepo:      contractRepo,
		trainingUC:        trainingUC,
		homeUC:            homeUC,
		personnelUC:       personnelUC,
		studentImportUC:   studentImportUC,
		trainingSummaryUC: trainingSummaryUC,
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
	var keyword, responsibleInstructorID, flightTrainingStatus, trainingCourse string
	if len(bytes.TrimSpace(raw)) > 0 {
		var body struct {
			Limit                   int    `json:"limit"`
			Offset                  int    `json:"offset"`
			Keyword                 string `json:"keyword"`
			ResponsibleInstructorID string `json:"responsible_instructor_id"`
			FlightTrainingStatus    string `json:"flight_training_status"`
			TrainingCourse          string `json:"training_course"`
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
		trainingCourse = body.TrainingCourse
	}
	if trainingCourse != "" {
		if _, ok := domain.ParseTrainingCourse(trainingCourse); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid training_course", "data": nil})
			return
		}
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
		TrainingCourse:          trainingCourse,
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
		TrainingCourse       string    `json:"training_course,omitempty"`
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
			TrainingCourse:          s.TrainingCourse,
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
		TrainingCourse          string `json:"training_course"`
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
	if body.TrainingCourse != "" {
		if _, ok := domain.ParseTrainingCourse(body.TrainingCourse); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid training_course", "data": nil})
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
		TrainingCourse:          body.TrainingCourse,
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
		TrainingCourse          string `json:"training_course,omitempty"`
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
			TrainingCourse:          s.TrainingCourse,
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
			"training_course":           s.TrainingCourse,
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
	items := usecase.ContractListItemsForAPI(list)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": items},
	})
}

// AdminRegistryCSVTemplate 下载学员总表 CSV 表头模板 POST /api/v1/admin/students/import/template-csv
func (h *StudentHandler) AdminRegistryCSVTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="student_registry_template.csv"`)
	if _, err := c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return
	}
	w := csv.NewWriter(c.Writer)
	_ = w.Write(domain.StudentRegistryCSVHeaders)
	w.Flush()
}

// AdminRegistryImportCSV 批量导入学员总表 CSV POST /api/v1/admin/students/import/csv — multipart 字段 file；可选 update_existing=1|true 覆盖同学号
func (h *StudentHandler) AdminRegistryImportCSV(c *gin.Context) {
	if h.studentImportUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "student import not configured", "data": nil})
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 32<<20+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if len(data) > 32<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "文件过大", "data": nil})
		return
	}
	updateExisting := c.PostForm("update_existing") == "1" || c.PostForm("update_existing") == "true"
	result, err := h.studentImportUC.ImportRegistryCSV(c.Request.Context(), bytes.NewReader(data), usecase.StudentRegistryImportOptions{
		UpdateExisting: updateExisting,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": result})
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

// MyTrainingSummary 学员本人训练信息汇总 POST /api/v1/students/my-training-summary
func (h *StudentHandler) MyTrainingSummary(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if domain.UserRole(claims.Role) != domain.RoleStudent {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	if h.trainingSummaryUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "training summary not configured", "data": nil})
		return
	}
	s, err := h.studentRepo.GetByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
		return
	}
	data, err := h.trainingSummaryUC.Build(c.Request.Context(), s.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// TrainingSummaryDetail 职员/教员查看某学员训练汇总 POST /api/v1/students/training-summary/detail
func (h *StudentHandler) TrainingSummaryDetail(c *gin.Context) {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if !isStaffLikeRole(claims.Role) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden", "data": nil})
		return
	}
	if h.trainingSummaryUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "training summary not configured", "data": nil})
		return
	}
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.trainingSummaryUC.Build(c.Request.Context(), body.StudentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "student not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// AdminTrainingSummarySave 管理员更新学员训练概要 POST /api/v1/admin/students/training-summary/save
func (h *StudentHandler) AdminTrainingSummarySave(c *gin.Context) {
	if h.trainingSummaryUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "training summary not configured", "data": nil})
		return
	}
	var body struct {
		StudentID               string  `json:"student_id" binding:"required"`
		TrainingCourse          *string `json:"training_course"`
		FlightTrainingStatus    *string `json:"flight_training_status"`
		ResponsibleInstructorID *string `json:"responsible_instructor_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	req := &usecase.SaveStudentTrainingProfileRequest{
		TrainingCourse:          body.TrainingCourse,
		FlightTrainingStatus:    body.FlightTrainingStatus,
		ResponsibleInstructorID: body.ResponsibleInstructorID,
	}
	s, err := h.trainingSummaryUC.SaveProfile(c.Request.Context(), body.StudentID, req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": s})
}

// AdminTrainingSummaryDeleteCourse 管理员清空训练课程阶段 POST /api/v1/admin/students/training-summary/delete
func (h *StudentHandler) AdminTrainingSummaryDeleteCourse(c *gin.Context) {
	if h.trainingSummaryUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "training summary not configured", "data": nil})
		return
	}
	var body struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	if err := h.trainingSummaryUC.ClearTrainingCourse(c.Request.Context(), body.StudentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": nil})
}

// AdminTrainingSummariesList 管理员分页查询学员训练汇总列表 POST /api/v1/admin/students/training-summaries/list
func (h *StudentHandler) AdminTrainingSummariesList(c *gin.Context) {
	if h.trainingSummaryUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "training summary not configured", "data": nil})
		return
	}
	var body struct {
		Keyword                 string `json:"keyword"`
		ResponsibleInstructorID string `json:"responsible_instructor_id"`
		FlightTrainingStatus    string `json:"flight_training_status"`
		TrainingCourse          string `json:"training_course"`
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
	if body.TrainingCourse != "" {
		if _, ok := domain.ParseTrainingCourse(body.TrainingCourse); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid training_course", "data": nil})
			return
		}
	}
	if body.Page <= 0 {
		body.Page = 1
	}
	if body.PageSize <= 0 || body.PageSize > 50 {
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
		TrainingCourse:          body.TrainingCourse,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	ctx := c.Request.Context()
	out := make([]interface{}, 0, len(list))
	for _, s := range list {
		sum, err := h.trainingSummaryUC.Build(ctx, s.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
			return
		}
		out = append(out, sum)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"items": out}})
}

func isStaffLikeRole(role string) bool {
	switch domain.UserRole(role) {
	case domain.RoleAdmin, domain.RoleStaff, domain.RoleMaintenance, domain.RoleInstructor:
		return true
	default:
		return false
	}
}
