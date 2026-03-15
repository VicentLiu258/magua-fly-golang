package usecase

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// TrainingUseCase 训练记录、导入认领、视图、档案与科目
type TrainingUseCase struct {
	recordRepo   repository.TrainingRecordRepository
	archiveRepo  repository.TrainingArchiveRepository
	rawImportRepo repository.RawFlightImportRepository
	viewRepo     repository.TrainingRecordViewRepository
	syllabusRepo repository.SyllabusSubjectRepository
	archiveSubjRepo repository.TrainingArchiveSubjectRepository
	nextSubjRepo repository.NextTrainingSubjectRepository
}

// NewTrainingUseCase 创建 TrainingUseCase
func NewTrainingUseCase(
	recordRepo repository.TrainingRecordRepository,
	archiveRepo repository.TrainingArchiveRepository,
	rawImportRepo repository.RawFlightImportRepository,
	viewRepo repository.TrainingRecordViewRepository,
	syllabusRepo repository.SyllabusSubjectRepository,
	archiveSubjRepo repository.TrainingArchiveSubjectRepository,
	nextSubjRepo repository.NextTrainingSubjectRepository,
) *TrainingUseCase {
	return &TrainingUseCase{
		recordRepo:     recordRepo,
		archiveRepo:    archiveRepo,
		rawImportRepo:  rawImportRepo,
		viewRepo:       viewRepo,
		syllabusRepo:   syllabusRepo,
		archiveSubjRepo: archiveSubjRepo,
		nextSubjRepo:   nextSubjRepo,
	}
}

// CreateRecord 创建训练记录（基础录入）
func (u *TrainingUseCase) CreateRecord(ctx context.Context, rec *domain.TrainingRecord) error {
	return u.recordRepo.Create(ctx, rec)
}

// GetRecord 单条记录详情
func (u *TrainingUseCase) GetRecord(ctx context.Context, id int64) (*domain.TrainingRecord, error) {
	return u.recordRepo.GetByID(ctx, id)
}

// UpdateRecord 更新记录
func (u *TrainingUseCase) UpdateRecord(ctx context.Context, rec *domain.TrainingRecord) error {
	return u.recordRepo.Update(ctx, rec)
}

// DeleteRecord 软删记录
func (u *TrainingUseCase) DeleteRecord(ctx context.Context, id int64) error {
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
func (u *TrainingUseCase) ClaimImport(ctx context.Context, rawImportID int64, studentID, instructorID, aircraftID int64, trainingDate *time.Time) (*domain.TrainingRecord, error) {
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
		RawImportID:           strconv.FormatInt(rawImportID, 10),
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
func (u *TrainingUseCase) ListViews(ctx context.Context, userID int64) ([]*domain.TrainingRecordView, error) {
	return u.viewRepo.ListByUserID(ctx, userID)
}

// GetView 获取视图
func (u *TrainingUseCase) GetView(ctx context.Context, id int64) (*domain.TrainingRecordView, error) {
	return u.viewRepo.GetByID(ctx, id)
}

// UpdateView 更新视图
func (u *TrainingUseCase) UpdateView(ctx context.Context, v *domain.TrainingRecordView) error {
	return u.viewRepo.Update(ctx, v)
}

// DeleteView 删除视图
func (u *TrainingUseCase) DeleteView(ctx context.Context, id int64) error {
	return u.viewRepo.Delete(ctx, id)
}

// ListArchivesByStudent 学员训练档案列表
func (u *TrainingUseCase) ListArchivesByStudent(ctx context.Context, studentID int64, limit, offset int) ([]*domain.TrainingArchive, error) {
	if limit <= 0 {
		limit = 20
	}
	return u.archiveRepo.ListByStudent(ctx, studentID, limit, offset)
}

// GetArchive 档案详情
func (u *TrainingUseCase) GetArchive(ctx context.Context, id int64) (*domain.TrainingArchive, error) {
	return u.archiveRepo.GetByID(ctx, id)
}

// CreateArchive 创建档案
func (u *TrainingUseCase) CreateArchive(ctx context.Context, a *domain.TrainingArchive) error {
	return u.archiveRepo.Create(ctx, a)
}

// UpdateArchive 更新档案
func (u *TrainingUseCase) UpdateArchive(ctx context.Context, a *domain.TrainingArchive) error {
	return u.archiveRepo.Update(ctx, a)
}

// SignArchive 签字
func (u *TrainingUseCase) SignArchive(ctx context.Context, archiveID int64, role string, signerID int64) error {
	a, err := u.archiveRepo.GetByID(ctx, archiveID)
	if err != nil {
		return err
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
	return u.archiveRepo.Update(ctx, a)
}

// ListSyllabusSubjects 全部大纲科目，可选带学员已结业标记
func (u *TrainingUseCase) ListSyllabusSubjects(ctx context.Context, studentID *int64) ([]*domain.SyllabusSubjectWithGraduated, error) {
	list, err := u.syllabusRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.SyllabusSubjectWithGraduated, len(list))
	for i := range list {
		out[i] = &domain.SyllabusSubjectWithGraduated{SyllabusSubject: *list[i]}
		if studentID != nil {
			// 简化：未实现按学员查已结业科目，可后续在 repo 层加 SQL
			_ = studentID
		}
	}
	return out, nil
}

// ListArchiveSubjects 本课科目列表
func (u *TrainingUseCase) ListArchiveSubjects(ctx context.Context, archiveID int64) ([]*domain.TrainingArchiveSubject, error) {
	return u.archiveSubjRepo.ListByArchiveID(ctx, archiveID)
}

// UpdateArchiveSubjects 设置本课科目
func (u *TrainingUseCase) UpdateArchiveSubjects(ctx context.Context, archiveID int64, subjects []*domain.TrainingArchiveSubject) error {
	return u.archiveSubjRepo.ReplaceForArchive(ctx, archiveID, subjects)
}

// ListNextSubjects 下次布置科目
func (u *TrainingUseCase) ListNextSubjects(ctx context.Context, archiveID int64) ([]*domain.NextTrainingSubject, error) {
	return u.nextSubjRepo.ListByArchiveID(ctx, archiveID)
}

// UpdateNextSubjects 布置下次科目
func (u *TrainingUseCase) UpdateNextSubjects(ctx context.Context, archiveID int64, assignedBy int64, syllabusSubjectIDs []int64) error {
	return u.nextSubjRepo.ReplaceForArchive(ctx, archiveID, assignedBy, syllabusSubjectIDs)
}

// TrainingRecordFieldsConfig 返回录入字段配置（静态）
func TrainingRecordFieldsConfig() ([]byte, error) {
	cfg := []map[string]interface{}{
		{"key": "student_id", "label": "学员", "type": "select", "required": true},
		{"key": "instructor_id", "label": "带飞教员", "type": "select", "required": true},
		{"key": "training_date", "label": "训练日期", "type": "date", "required": true},
		{"key": "engine_on_time", "label": "开车时间", "type": "time"},
		{"key": "engine_off_time", "label": "关车时间", "type": "time"},
		{"key": "flight_duration_minutes", "label": "飞行时长(分钟)", "type": "number"},
		{"key": "is_local", "label": "本场/转场", "type": "boolean"},
		{"key": "status", "label": "状态", "type": "select", "options": []string{"normal", "cancelled", "deleted", "expired"}},
	}
	return json.Marshal(cfg)
}
