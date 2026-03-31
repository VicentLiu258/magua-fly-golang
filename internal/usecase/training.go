package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// ErrTrainingRecordNotFound 表示指定的 training_record_id 不存在，创建档案前需先有对应训练记录。
var ErrTrainingRecordNotFound = errors.New("training_record_id 不存在，请先创建训练记录或使用已有的记录 ID")

// ErrTrainingRecordExportTooManyRows 导出行数超过 domain.MaxTrainingRecordExportRows。
var ErrTrainingRecordExportTooManyRows = errors.New("导出行数超过上限，请缩小日期范围或筛选条件后再试")

// ErrInvalidSyllabusOutlineCategory 大纲类别非法（与 syllabus_outlines CHECK 不一致）
var ErrInvalidSyllabusOutlineCategory = errors.New("invalid syllabus outline category")

// TrainingUseCase 训练记录、导入认领、视图、档案与科目
type TrainingUseCase struct {
	recordRepo       repository.TrainingRecordRepository
	archiveRepo      repository.TrainingArchiveRepository
	rawImportRepo    repository.RawFlightImportRepository
	viewRepo         repository.TrainingRecordViewRepository
	syllabusRepo     repository.SyllabusSubjectRepository
	outlineRepo      repository.SyllabusOutlineRepository
	secondaryRepo    repository.SyllabusSubjectSecondaryRepository
	syllabusTreeRepo repository.SyllabusTreeRepository
	archiveSubjRepo  repository.TrainingArchiveSubjectRepository
	nextSubjRepo     repository.NextTrainingSubjectRepository
	taskRepo         repository.TaskRepository
}

// NewTrainingUseCase 创建 TrainingUseCase
func NewTrainingUseCase(
	recordRepo repository.TrainingRecordRepository,
	archiveRepo repository.TrainingArchiveRepository,
	rawImportRepo repository.RawFlightImportRepository,
	viewRepo repository.TrainingRecordViewRepository,
	syllabusRepo repository.SyllabusSubjectRepository,
	outlineRepo repository.SyllabusOutlineRepository,
	secondaryRepo repository.SyllabusSubjectSecondaryRepository,
	syllabusTreeRepo repository.SyllabusTreeRepository,
	archiveSubjRepo repository.TrainingArchiveSubjectRepository,
	nextSubjRepo repository.NextTrainingSubjectRepository,
	taskRepo repository.TaskRepository,
) *TrainingUseCase {
	return &TrainingUseCase{
		recordRepo:       recordRepo,
		archiveRepo:      archiveRepo,
		rawImportRepo:    rawImportRepo,
		viewRepo:         viewRepo,
		syllabusRepo:     syllabusRepo,
		outlineRepo:      outlineRepo,
		secondaryRepo:    secondaryRepo,
		syllabusTreeRepo: syllabusTreeRepo,
		archiveSubjRepo:  archiveSubjRepo,
		nextSubjRepo:     nextSubjRepo,
		taskRepo:         taskRepo,
	}
}

// AdminCreateSyllabusSecondaryInput 创建一级科目时可选的二级行
type AdminCreateSyllabusSecondaryInput struct {
	SeqNo               int    `json:"seq_no"`
	Name                string `json:"name" binding:"required"`
	Description         string `json:"description"`
	CompletionStandard  string `json:"completion_standard"`
	SortOrder           int    `json:"sort_order"`
}

// AdminCreateSyllabusSubjectInput 后台创建一级科目（须绑定大纲）
type AdminCreateSyllabusSubjectInput struct {
	SyllabusOutlineID *string                            `json:"syllabus_outline_id"`
	OutlineCategory   *string                            `json:"outline_category"`
	OutlineName       *string                            `json:"outline_name"`
	OutlineSortOrder  int                                `json:"outline_sort_order"`
	Code              string                             `json:"code" binding:"required"`
	Name              string                             `json:"name" binding:"required"`
	SortOrder         int                                `json:"sort_order"`
	Secondaries       []AdminCreateSyllabusSecondaryInput `json:"secondaries"`
}

// CreateRecord 创建训练记录（基础录入）
func (u *TrainingUseCase) CreateRecord(ctx context.Context, rec *domain.TrainingRecord) error {
	return u.recordRepo.Create(ctx, rec)
}

// EnsureScheduleDraftRecord 预约审核通过并生成 training_plan 后，创建一条待补全飞行记录（幂等）。
func (u *TrainingUseCase) EnsureScheduleDraftRecord(ctx context.Context, plan *domain.TrainingPlan) error {
	if u == nil || plan == nil || plan.ID == "" {
		return nil
	}
	existing, err := u.recordRepo.FindByTrainingPlanID(ctx, plan.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	mins := domain.PlannedDurationMinutesFromClock(plan.StartTime, plan.EndTime)
	rec := &domain.TrainingRecord{
		StudentID:              plan.StudentID,
		InstructorID:           plan.InstructorID,
		AircraftID:             plan.AircraftID,
		TrainingDate:           plan.PlanDate,
		IsLocal:                true,
		ImportSource:           domain.ImportSchedule,
		Status:                 domain.RecordStatusDraft,
		TrainingPlanID:         plan.ID,
		PlannedDurationMinutes: &mins,
	}
	return u.recordRepo.Create(ctx, rec)
}

// GetRecord 单条记录详情
func (u *TrainingUseCase) GetRecord(ctx context.Context, id string) (*domain.TrainingRecord, error) {
	return u.recordRepo.GetByID(ctx, id)
}

// UpdateRecord 更新记录
func (u *TrainingUseCase) UpdateRecord(ctx context.Context, rec *domain.TrainingRecord) error {
	return u.recordRepo.Update(ctx, rec)
}

// DeleteRecord 软删记录
func (u *TrainingUseCase) DeleteRecord(ctx context.Context, id string) error {
	return u.recordRepo.UpdateStatus(ctx, id, domain.RecordStatusDeleted)
}

// ListRecords 多维列表
func (u *TrainingUseCase) ListRecords(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit, offset int) (*domain.TrainingRecordListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	items, total, err := u.recordRepo.List(ctx, filter, sort, limit, offset)
	if err != nil {
		return nil, err
	}
	return &domain.TrainingRecordListResult{Items: items, Total: total}, nil
}

// ExportTrainingRecordsCSV 按与列表相同的筛选条件导出 CSV 数据行（不含表头）；行数超过上限时返回 ErrTrainingRecordExportTooManyRows。
func (u *TrainingUseCase) ExportTrainingRecordsCSV(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, includePriorExperience bool) ([][]string, error) {
	n, err := u.recordRepo.CountForExport(ctx, filter)
	if err != nil {
		return nil, err
	}
	if n > domain.MaxTrainingRecordExportRows {
		return nil, ErrTrainingRecordExportTooManyRows
	}
	return u.recordRepo.ListForCSVExport(ctx, filter, sort, n, includePriorExperience)
}

// ImportRaw 批量导入原始飞行数据，写入 raw_flight_imports
func (u *TrainingUseCase) ImportRaw(ctx context.Context, batchID string, rows []domain.RawFlightImport) ([]*domain.RawFlightImport, error) {
	var created []*domain.RawFlightImport
	for i := range rows {
		rows[i].BatchID = batchID
		if err := u.rawImportRepo.Create(ctx, &rows[i]); err != nil {
			return nil, err
		}
		created = append(created, &rows[i])
	}
	return created, nil
}

// ClaimImport 认领一条导入记录，创建 training_record（及可选 archive）
func (u *TrainingUseCase) ClaimImport(ctx context.Context, rawImportID string, studentID, instructorID, aircraftID string, trainingDate *time.Time) (*domain.TrainingRecord, error) {
	raw, err := u.rawImportRepo.GetByID(ctx, rawImportID)
	if err != nil {
		return nil, err
	}
	if raw.TrainingRecordID != nil {
		return nil, nil // 已认领，可返回错误或已存在记录
	}
	rec := &domain.TrainingRecord{
		StudentID:            studentID,
		InstructorID:          instructorID,
		AircraftID:            aircraftID,
		ImportSource:          domain.ImportImport,
		RawImportID:           rawImportID,
		IsLocal:               raw.IsLocal,
		EngineOnTime:          raw.EngineOnTime,
		EngineOffTime:         raw.EngineOffTime,
		FlightDurationMinutes: raw.FlightDurationMinutes,
		Status:                domain.RecordStatusNormal,
	}
	if trainingDate != nil {
		rec.TrainingDate = *trainingDate
	} else if raw.EngineOnTime != nil {
		rec.TrainingDate = raw.EngineOnTime.Truncate(24 * time.Hour)
	} else {
		rec.TrainingDate = time.Now().Truncate(24 * time.Hour)
	}
	if err := u.recordRepo.Create(ctx, rec); err != nil {
		return nil, err
	}
	if err := u.rawImportRepo.SetTrainingRecordID(ctx, rawImportID, rec.ID); err != nil {
		return nil, err
	}
	// 可选：创建 training_archives
	archive := &domain.TrainingArchive{
		TrainingRecordID:       rec.ID,
		StudentID:              rec.StudentID,
		InstructorID:           rec.InstructorID,
		TrainingDate:           rec.TrainingDate,
		EngineOnTime:           rec.EngineOnTime,
		EngineOffTime:          rec.EngineOffTime,
		FlightDurationMinutes:  rec.FlightDurationMinutes,
	}
	_ = u.archiveRepo.Create(ctx, archive)
	return rec, nil
}

// ListPendingImports 待认领列表
func (u *TrainingUseCase) ListPendingImports(ctx context.Context, limit, offset int) ([]*domain.RawFlightImport, error) {
	if limit <= 0 {
		limit = 50
	}
	return u.rawImportRepo.ListPending(ctx, limit, offset)
}

// Stats 按维度统计（简单实现：列表后内存聚合）
func (u *TrainingUseCase) Stats(ctx context.Context, dateFrom, dateTo *time.Time, dimensions []string) (map[string]interface{}, error) {
	filter := &domain.TrainingRecordListFilter{
		Status: ptrRecordStatus(domain.RecordStatusNormal),
	}
	if dateFrom != nil {
		filter.DateFrom = dateFrom
	}
	if dateTo != nil {
		filter.DateTo = dateTo
	}
	items, _, err := u.recordRepo.List(ctx, filter, nil, 10000, 0)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{})
	out["count"] = len(items)
	var totalMin int
	for _, r := range items {
		if r.FlightDurationMinutes != nil {
			totalMin += *r.FlightDurationMinutes
		}
	}
	out["total_flight_minutes"] = totalMin
	return out, nil
}

func ptrRecordStatus(s domain.TrainingRecordStatus) *domain.TrainingRecordStatus { return &s }

// DashboardSummary 仪表盘概览
func (u *TrainingUseCase) DashboardSummary(ctx context.Context) (map[string]interface{}, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	result, err := u.Stats(ctx, &today, &now, nil)
	if err != nil {
		return nil, err
	}
	result["today_count"] = result["count"]
	result["today_total_flight_minutes"] = result["total_flight_minutes"]
	return result, nil
}

// CreateView 保存自定义视图
func (u *TrainingUseCase) CreateView(ctx context.Context, v *domain.TrainingRecordView) error {
	return u.viewRepo.Create(ctx, v)
}

// ListViews 当前用户视图列表
func (u *TrainingUseCase) ListViews(ctx context.Context, userID string) ([]*domain.TrainingRecordView, error) {
	return u.viewRepo.ListByUserID(ctx, userID)
}

// GetView 获取视图
func (u *TrainingUseCase) GetView(ctx context.Context, id string) (*domain.TrainingRecordView, error) {
	return u.viewRepo.GetByID(ctx, id)
}

// UpdateView 更新视图
func (u *TrainingUseCase) UpdateView(ctx context.Context, v *domain.TrainingRecordView) error {
	return u.viewRepo.Update(ctx, v)
}

// DeleteView 删除视图
func (u *TrainingUseCase) DeleteView(ctx context.Context, id string) error {
	return u.viewRepo.Delete(ctx, id)
}

// ListArchivesByStudent 学员训练档案列表
func (u *TrainingUseCase) ListArchivesByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingArchive, error) {
	if limit <= 0 {
		limit = 20
	}
	return u.archiveRepo.ListByStudent(ctx, studentID, limit, offset)
}

// GetArchive 档案详情
func (u *TrainingUseCase) GetArchive(ctx context.Context, id string) (*domain.TrainingArchive, error) {
	return u.archiveRepo.GetByID(ctx, id)
}

// CreateArchive 创建档案。若 training_record_id 不存在则返回 ErrTrainingRecordNotFound。
func (u *TrainingUseCase) CreateArchive(ctx context.Context, a *domain.TrainingArchive) error {
	_, err := u.recordRepo.GetByID(ctx, a.TrainingRecordID)
	if err != nil {
		return fmt.Errorf("%w (id=%s)", ErrTrainingRecordNotFound, a.TrainingRecordID)
	}
	return u.archiveRepo.Create(ctx, a)
}

// UpdateArchive 更新档案
func (u *TrainingUseCase) UpdateArchive(ctx context.Context, a *domain.TrainingArchive) error {
	return u.archiveRepo.Update(ctx, a)
}

// SignArchive 签字
func (u *TrainingUseCase) SignArchive(ctx context.Context, archiveID string, role string, signerID string) error {
	a, err := u.archiveRepo.GetByID(ctx, archiveID)
	if err != nil {
		return err
	}
	var nextSubjects []*domain.NextTrainingSubject
	// 业务约束：责任教员 / 主任教员签字前，必须先布置下次训练科目
	if role == "responsible_instructor" || role == "chief_instructor" {
		nextSubjects, err = u.nextSubjRepo.ListByArchiveID(ctx, archiveID)
		if err != nil {
			return err
		}
		if len(nextSubjects) == 0 {
			return fmt.Errorf("请先布置下次训练科目")
		}
	}
	now := time.Now()
	switch role {
	case "student":
		a.StudentSignedAt = &now
	case "responsible_instructor":
		a.ResponsibleInstructorSignedAt = &now
		a.ResponsibleInstructorID = &signerID
	case "chief_instructor":
		a.ChiefInstructorSignedAt = &now
		a.ChiefInstructorID = &signerID
	default:
		return nil
	}
	if err := u.archiveRepo.Update(ctx, a); err != nil {
		return err
	}

	// 责任教员签字成功后，为学员生成基于本课档案与下次科目的「下次飞行任务」。
	if role == "responsible_instructor" && len(nextSubjects) > 0 {
		_ = u.createNextFlightTaskFromArchive(ctx, a, nextSubjects, signerID)
	}

	return nil
}

// createNextFlightTaskFromArchive 基于当前档案与已布置的下次科目，为学员生成一条飞行类任务。
// 若任务已存在或缺少必要信息，则直接返回。
func (u *TrainingUseCase) createNextFlightTaskFromArchive(ctx context.Context, a *domain.TrainingArchive, nextSubjects []*domain.NextTrainingSubject, assignerID string) error {
	if u.taskRepo == nil || a == nil || a.StudentID == "" || len(nextSubjects) == 0 {
		return nil
	}
	// 幂等：同一档案只生成一条「下次飞行任务」。
	existing, err := u.taskRepo.FindByTypeAndRefID(ctx, domain.TaskTypeFlight, a.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	// 任务标题 / 描述聚合科目信息。
	title := "下次飞行科目任务"
	if len(nextSubjects) == 1 {
		title = "下次飞行科目任务（1 个科目）"
	} else {
		title = fmt.Sprintf("下次飞行科目任务（%d 个科目）", len(nextSubjects))
	}
	desc := fmt.Sprintf("基于档案 %s 布置的下次飞行训练任务", a.ID)

	refID := a.ID
	task := &domain.Task{
		StudentID:   a.StudentID,
		AssignerID:  assignerID,
		Type:        domain.TaskTypeFlight,
		RefID:       &refID,
		Title:       title,
		Description: desc,
		Status:      domain.TaskStatusPending,
	}
	return u.taskRepo.Create(ctx, task)
}

// ListSyllabusSubjects 大纲下一级科目（可选按大纲 ID 或类别过滤），带二级列表；可选带学员已结业标记
func (u *TrainingUseCase) ListSyllabusSubjects(ctx context.Context, studentID *string, syllabusOutlineID *string, outlineCategory *string) ([]*domain.SyllabusSubjectWithGraduated, error) {
	var listFilter *domain.SyllabusSubjectListFilter
	if syllabusOutlineID != nil && *syllabusOutlineID != "" {
		listFilter = &domain.SyllabusSubjectListFilter{OutlineID: syllabusOutlineID}
	} else if outlineCategory != nil && *outlineCategory != "" {
		if _, err := domain.ParseSyllabusOutlineCategory(*outlineCategory); err != nil {
			return nil, fmt.Errorf("%w (%v)", ErrInvalidSyllabusOutlineCategory, err)
		}
		c := *outlineCategory
		listFilter = &domain.SyllabusSubjectListFilter{Category: &c}
	}

	list, err := u.syllabusRepo.List(ctx, listFilter)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	secMap, err := u.secondaryRepo.ListGroupedBySubjectID(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.SyllabusSubjectWithGraduated, len(list))
	for i := range list {
		out[i] = &domain.SyllabusSubjectWithGraduated{
			SyllabusSubject: *list[i],
			Secondaries:     secMap[list[i].ID],
		}
		if out[i].Secondaries == nil {
			out[i].Secondaries = []*domain.SyllabusSubjectSecondary{}
		}
	}
	// 若未传学员 ID，则仅返回基础科目信息
	if studentID == nil {
		return out, nil
	}

	// 按学员聚合其历史档案中的科目信息，以推导科目状态：
	// completed | in_progress | not_started
	archives, err := u.archiveRepo.ListByStudent(ctx, *studentID, 1000, 0)
	if err != nil {
		return nil, err
	}
	type subjStat struct {
		hasAny      bool
		hasGraduate bool
	}
	stats := make(map[string]*subjStat)
	for _, a := range archives {
		subjects, err := u.archiveSubjRepo.ListByArchiveID(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		for _, s := range subjects {
			st, ok := stats[s.SyllabusSubjectID]
			if !ok {
				st = &subjStat{}
				stats[s.SyllabusSubjectID] = st
			}
			st.hasAny = true
			if s.IsGraduated {
				st.hasGraduate = true
			}
		}
	}

	for _, item := range out {
		if st, ok := stats[item.ID]; ok {
			item.IsGraduated = st.hasGraduate
			if st.hasGraduate {
				item.Status = "completed"
			} else {
				item.Status = "in_progress"
			}
		} else {
			item.IsGraduated = false
			item.Status = "not_started"
		}
	}

	return out, nil
}

// ListSyllabusOutlines 训练大纲列表（后台选题）
func (u *TrainingUseCase) ListSyllabusOutlines(ctx context.Context) ([]*domain.SyllabusOutline, error) {
	return u.outlineRepo.List(ctx)
}

// GetSyllabusOutlineByID 获取单条大纲
func (u *TrainingUseCase) GetSyllabusOutlineByID(ctx context.Context, id string) (*domain.SyllabusOutline, error) {
	return u.outlineRepo.GetByID(ctx, id)
}

// CreateSyllabusSubjectAdmin 创建一级科目：须指定已有 syllabus_outline_id，或 outline_category+outline_name 新建大纲
func (u *TrainingUseCase) CreateSyllabusSubjectAdmin(ctx context.Context, in *AdminCreateSyllabusSubjectInput) (*domain.SyllabusSubject, []*domain.SyllabusSubjectSecondary, error) {
	hasOutlineID := in.SyllabusOutlineID != nil && *in.SyllabusOutlineID != ""
	hasNewOutline := in.OutlineCategory != nil && *in.OutlineCategory != "" && in.OutlineName != nil && *in.OutlineName != ""
	if hasOutlineID && hasNewOutline {
		return nil, nil, fmt.Errorf("syllabus_outline_id 与 outline_category/outline_name 不能同时指定")
	}
	if !hasOutlineID && !hasNewOutline {
		return nil, nil, fmt.Errorf("需提供 syllabus_outline_id 或 outline_category+outline_name")
	}
	if hasOutlineID {
		if _, err := u.outlineRepo.GetByID(ctx, *in.SyllabusOutlineID); err != nil {
			return nil, nil, fmt.Errorf("大纲不存在: %w", err)
		}
	}
	var outline *domain.SyllabusOutline
	var outlineIDPtr *string
	if hasNewOutline {
		cat, err := domain.ParseSyllabusOutlineCategory(*in.OutlineCategory)
		if err != nil {
			return nil, nil, err
		}
		outline = &domain.SyllabusOutline{
			Category:  string(cat),
			Name:      *in.OutlineName,
			SortOrder: in.OutlineSortOrder,
		}
		outlineIDPtr = nil
	} else {
		outlineIDPtr = in.SyllabusOutlineID
	}
	subject := &domain.SyllabusSubject{
		Code:      in.Code,
		Name:      in.Name,
		SortOrder: in.SortOrder,
	}
	var secs []*domain.SyllabusSubjectSecondary
	for _, x := range in.Secondaries {
		secs = append(secs, &domain.SyllabusSubjectSecondary{
			SeqNo:              x.SeqNo,
			Name:               x.Name,
			Description:        x.Description,
			CompletionStandard: x.CompletionStandard,
			SortOrder:          x.SortOrder,
		})
	}
	if err := u.syllabusTreeRepo.CreateSubjectUnderOutline(ctx, outlineIDPtr, outline, subject, secs); err != nil {
		return nil, nil, err
	}
	list, err := u.secondaryRepo.ListBySubjectID(ctx, subject.ID)
	if err != nil {
		return subject, nil, err
	}
	return subject, list, nil
}

// ImportSyllabusOutline 导入整棵大纲树（总是新建一条大纲记录）
func (u *TrainingUseCase) ImportSyllabusOutline(ctx context.Context, category, name string, sortOrder int, subjects []domain.SyllabusImportSubjectRow) (*domain.SyllabusOutline, error) {
	cat, err := domain.ParseSyllabusOutlineCategory(category)
	if err != nil {
		return nil, err
	}
	o := &domain.SyllabusOutline{Category: string(cat), Name: name, SortOrder: sortOrder}
	return u.syllabusTreeRepo.ImportOutlineTree(ctx, o, subjects)
}

// CreateSyllabusSubjectSecondary 新增二级科目
func (u *TrainingUseCase) CreateSyllabusSubjectSecondary(ctx context.Context, row *domain.SyllabusSubjectSecondary) error {
	if _, err := u.syllabusRepo.GetByID(ctx, row.SyllabusSubjectID); err != nil {
		return fmt.Errorf("一级科目不存在: %w", err)
	}
	return u.secondaryRepo.Create(ctx, row)
}

// UpdateSyllabusSubjectSecondary 更新二级科目
func (u *TrainingUseCase) UpdateSyllabusSubjectSecondary(ctx context.Context, row *domain.SyllabusSubjectSecondary) error {
	return u.secondaryRepo.Update(ctx, row)
}

// DeleteSyllabusSubjectSecondary 删除二级科目
func (u *TrainingUseCase) DeleteSyllabusSubjectSecondary(ctx context.Context, id string) error {
	return u.secondaryRepo.Delete(ctx, id)
}

// ReorderSyllabusSubjectSecondaries 按 ordered_ids 重写 sort_order
func (u *TrainingUseCase) ReorderSyllabusSubjectSecondaries(ctx context.Context, syllabusSubjectID string, orderedIDs []string) error {
	if _, err := u.syllabusRepo.GetByID(ctx, syllabusSubjectID); err != nil {
		return fmt.Errorf("一级科目不存在: %w", err)
	}
	return u.secondaryRepo.Reorder(ctx, syllabusSubjectID, orderedIDs)
}

// GetSyllabusSubject 获取单个大纲科目
func (u *TrainingUseCase) GetSyllabusSubject(ctx context.Context, id string) (*domain.SyllabusSubject, error) {
	return u.syllabusRepo.GetByID(ctx, id)
}

// UpdateSyllabusSubject 更新大纲科目
func (u *TrainingUseCase) UpdateSyllabusSubject(ctx context.Context, s *domain.SyllabusSubject) error {
	return u.syllabusRepo.Update(ctx, s)
}

// DeleteSyllabusSubject 删除大纲科目
func (u *TrainingUseCase) DeleteSyllabusSubject(ctx context.Context, id string) error {
	return u.syllabusRepo.Delete(ctx, id)
}

// ListArchiveSubjects 本课科目列表
func (u *TrainingUseCase) ListArchiveSubjects(ctx context.Context, archiveID string) ([]*domain.TrainingArchiveSubject, error) {
	return u.archiveSubjRepo.ListByArchiveID(ctx, archiveID)
}

// UpdateArchiveSubjects 设置本课科目
func (u *TrainingUseCase) UpdateArchiveSubjects(ctx context.Context, archiveID string, subjects []*domain.TrainingArchiveSubject) error {
	return u.archiveSubjRepo.ReplaceForArchive(ctx, archiveID, subjects)
}

// ListNextSubjects 下次布置科目
func (u *TrainingUseCase) ListNextSubjects(ctx context.Context, archiveID string) ([]*domain.NextTrainingSubject, error) {
	return u.nextSubjRepo.ListByArchiveID(ctx, archiveID)
}

// UpdateNextSubjects 布置下次科目
func (u *TrainingUseCase) UpdateNextSubjects(ctx context.Context, archiveID string, assignedBy string, syllabusSubjectIDs []string) error {
	resolved := make([]string, 0, len(syllabusSubjectIDs))
	for _, sid := range syllabusSubjectIDs {
		if sid == "" {
			continue
		}
		pid, err := u.resolveSyllabusSubjectIDForNextTraining(ctx, sid)
		if err != nil {
			return err
		}
		resolved = append(resolved, pid)
	}
	return u.nextSubjRepo.ReplaceForArchive(ctx, archiveID, assignedBy, resolved)
}

// resolveSyllabusSubjectIDForNextTraining 接受一级科目 id 或二级科目 id，统一解析为一级科目 id（与 next_training_subjects 外键一致）。
func (u *TrainingUseCase) resolveSyllabusSubjectIDForNextTraining(ctx context.Context, id string) (string, error) {
	_, err := u.syllabusRepo.GetByID(ctx, id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	sec, err := u.secondaryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("无效的科目 id（需为一级大纲科目或二级科目）: %s", id)
		}
		return "", err
	}
	return sec.SyllabusSubjectID, nil
}

// TrainingRecordFieldsConfig 返回录入字段配置（静态）
func TrainingRecordFieldsConfig() ([]byte, error) {
	cfg := []map[string]interface{}{
		{"key": "student_id", "label": "学员", "type": "select", "required": true},
		{"key": "instructor_id", "label": "带飞教员", "type": "select", "required": true},
		{"key": "training_date", "label": "训练日期", "type": "date", "required": true},
		{"key": "training_plan_id", "label": "关联训练计划", "type": "string", "readonly": true},
		{"key": "engine_on_time", "label": "开车时间", "type": "time"},
		{"key": "engine_off_time", "label": "关车时间", "type": "time"},
		{"key": "flight_duration_minutes", "label": "飞行时长(分钟)", "type": "number"},
		{"key": "is_local", "label": "本场/转场", "type": "boolean"},
		{"key": "import_source", "label": "来源", "type": "select", "options": []string{"manual", "import", "schedule"}},
		{"key": "status", "label": "状态", "type": "select", "options": []string{"draft", "normal", "cancelled", "deleted", "expired"}},
	}
	return json.Marshal(cfg)
}
