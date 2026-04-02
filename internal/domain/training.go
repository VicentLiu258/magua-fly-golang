package domain

import (
	"fmt"
	"time"
)

// ImportSource for training record.
type ImportSource string

const (
	ImportManual   ImportSource = "manual"
	ImportImport   ImportSource = "import"
	ImportSchedule ImportSource = "schedule" // 预约/排期通过后自动生成的待补全记录
)

// TrainingRecordStatus 训练记录状态
type TrainingRecordStatus string

const (
	RecordStatusNormal    TrainingRecordStatus = "normal"
	RecordStatusCancelled TrainingRecordStatus = "cancelled"
	RecordStatusDeleted   TrainingRecordStatus = "deleted"
	RecordStatusExpired   TrainingRecordStatus = "expired"
	RecordStatusDraft     TrainingRecordStatus = "draft" // 已关联排期、待补充起落等字段
)

// TrainingRecord is one flight training record.
type TrainingRecord struct {
	ID                     string                `json:"id"`
	StudentID              string                `json:"student_id"`
	InstructorID           string                `json:"instructor_id"`
	AircraftID             string                `json:"aircraft_id"`
	TrainingDate           time.Time             `json:"training_date"`
	EngineOnTime           *time.Time            `json:"engine_on_time,omitempty"`
	EngineOffTime          *time.Time            `json:"engine_off_time,omitempty"`
	FlightDurationMinutes   *int                  `json:"flight_duration_minutes,omitempty"`
	IsLocal                bool                  `json:"is_local"`
	ImportSource           ImportSource          `json:"import_source"`
	RawImportID            string                `json:"raw_import_id,omitempty"`
	Status                 TrainingRecordStatus  `json:"status"`
	PlannedDurationMinutes *int                  `json:"planned_duration_minutes,omitempty"`
	TheoryDurationMinutes  *int                  `json:"theory_duration_minutes,omitempty"`
	SimDurationMinutes     *int                  `json:"sim_duration_minutes,omitempty"`
	SyllabusOutline        string                `json:"syllabus_outline,omitempty"`
	ArchiveDisplayNo       string                `json:"archive_display_no,omitempty"`
	Departure              string                `json:"departure,omitempty"`
	Destination            string                `json:"destination,omitempty"`
	FlightType             string                `json:"flight_type,omitempty"`
	TaskType               string                `json:"task_type,omitempty"`
	Landings               *int                  `json:"landings,omitempty"`
	TrainingPlanID         string                `json:"training_plan_id,omitempty"`
	RecordSerialNo         string                `json:"record_serial_no,omitempty"` // TR+YYYYMMDD+序号
	BusinessType           string                `json:"business_type,omitempty"`
	EngineStartMeter       *float64              `json:"engine_start_meter,omitempty"`
	EngineStopMeter        *float64              `json:"engine_stop_meter,omitempty"`
	VideoTimeliness        string                `json:"video_timeliness,omitempty"`
	VideoOperator          string                `json:"video_operator,omitempty"`
	VideoValidity          string                `json:"video_validity,omitempty"`
	FlightDataRef          string                `json:"flight_data_ref,omitempty"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

// PlannedDurationMinutesFromClock 根据计划起止时刻（仅时钟部分）计算分钟数，跨零点则加 24h。
func PlannedDurationMinutesFromClock(start, end time.Time) int {
	sh, sm, _ := start.Clock()
	eh, em, _ := end.Clock()
	smin := sh*60 + sm
	emin := eh*60 + em
	d := emin - smin
	if d < 0 {
		d += 24 * 60
	}
	return d
}

// TrainingRecordListFilter 列表筛选条件
type TrainingRecordListFilter struct {
	StudentID    *string    `json:"student_id,omitempty"`
	InstructorID *string    `json:"instructor_id,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	Status       *TrainingRecordStatus `json:"status,omitempty"`
	Stage        string     `json:"stage,omitempty"`
	ImportSource *ImportSource `json:"import_source,omitempty"`
}

// TrainingRecordListSort 排序
type TrainingRecordListSort struct {
	Field string `json:"field"` // training_date, created_at, flight_duration_minutes 等
	Asc   bool   `json:"asc"`
}

// TrainingRecordListResult 列表结果
type TrainingRecordListResult struct {
	Items []*TrainingRecord `json:"items"`
	Total int               `json:"total"`
}

// MaxTrainingRecordExportRows 单次 CSV 导出行数上限（避免内存与响应过大）。
const MaxTrainingRecordExportRows = 50000

// MaxTrainingRecordCSVImportRows 单次 CSV 导入行数上限。
const MaxTrainingRecordCSVImportRows = 50000

// TrainingRecordCSVHeaders 与业务样稿一致的总表列名（37 列）。
var TrainingRecordCSVHeaders = []string{
	"记录编号", "飞行日期", "星期", "计划开车时刻", "计划关车时刻", "开车时刻", "关车时刻", "机号",
	"始发", "目的", "本场/转场", "任务类型", "带飞/单飞", "业务类型", "飞行教员", "客户", "学员阶段", "客户编号", "备注",
	"飞行时长", "分钟数", "小时数", "起落架次", "此前经历", "发动机开车", "发动机关车", "年月",
	"关联合同", "小时价值", "本次价值", "校验", "校验结果", "视频及时", "拍摄人", "视频有效", "航后讲评", "飞行数据",
}

// TrainingArchive is the detailed archive for one training record.
type TrainingArchive struct {
	ID                            string     `json:"id"`
	TrainingRecordID              string     `json:"training_record_id"`
	StudentID                     string     `json:"student_id"`
	InstructorID                  string     `json:"instructor_id"`
	TrainingDate                  time.Time  `json:"training_date"`
	Stage                         string     `json:"stage,omitempty"`
	PlannedDurationMinutes        *int       `json:"planned_duration_minutes,omitempty"`
	TotalExperienceMinutes        *int       `json:"total_experience_minutes,omitempty"`
	RemainingMinutes              *int       `json:"remaining_minutes,omitempty"`
	EngineOnTime                  *time.Time `json:"engine_on_time,omitempty"`
	EngineOffTime                 *time.Time `json:"engine_off_time,omitempty"`
	FlightDurationMinutes         *int       `json:"flight_duration_minutes,omitempty"`
	ProgressNote                  string     `json:"progress_note,omitempty"`
	InstructorReview              string     `json:"instructor_review,omitempty"`
	StudentSelfReview             string     `json:"student_self_review,omitempty"`
	TaskAssignment                string     `json:"task_assignment,omitempty"`
	StudentSignedAt               *time.Time `json:"student_signed_at,omitempty"`
	ResponsibleInstructorSignedAt *time.Time `json:"responsible_instructor_signed_at,omitempty"`
	ChiefInstructorSignedAt       *time.Time `json:"chief_instructor_signed_at,omitempty"`
	ResponsibleInstructorID       *string    `json:"responsible_instructor_id,omitempty"`
	ChiefInstructorID             *string    `json:"chief_instructor_id,omitempty"`
	CreatedAt                     time.Time  `json:"created_at"`
	UpdatedAt                     time.Time  `json:"updated_at"`
}

// SyllabusOutlineCategory 训练大纲类别（与 syllabus_outlines.category CHECK 一致）
type SyllabusOutlineCategory string

const (
	SyllabusCatSPLLocal      SyllabusOutlineCategory = "spl_local"       // SPL本场
	SyllabusCatSPLTransition SyllabusOutlineCategory = "spl_transition" // SPL转场
	SyllabusCatSPLNight      SyllabusOutlineCategory = "spl_night"      // SPL夜航
	SyllabusCatPPLBasic      SyllabusOutlineCategory = "ppl_basic"      // PPL基本
	SyllabusCatPPLNight      SyllabusOutlineCategory = "ppl_night"      // PPL夜航
)

// ParseSyllabusOutlineCategory 校验类别字符串
func ParseSyllabusOutlineCategory(s string) (SyllabusOutlineCategory, error) {
	c := SyllabusOutlineCategory(s)
	switch c {
	case SyllabusCatSPLLocal, SyllabusCatSPLTransition, SyllabusCatSPLNight, SyllabusCatPPLBasic, SyllabusCatPPLNight:
		return c, nil
	default:
		return "", fmt.Errorf("invalid syllabus outline category: %q", s)
	}
}

// SyllabusOutline 训练大纲实例（某类别下可多版本/多条）
type SyllabusOutline struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyllabusSubject is a level-1 subject under a syllabus outline.
type SyllabusSubject struct {
	ID                string    `json:"id"`
	SyllabusOutlineID string    `json:"syllabus_outline_id"`
	Code              string    `json:"code"`
	Name              string    `json:"name"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SyllabusSubjectListFilter 一级科目列表过滤（nil 表示不过滤）
type SyllabusSubjectListFilter struct {
	OutlineID *string
	Category  *string
}

// SyllabusSubjectSecondary 一级科目下的二级科目
type SyllabusSubjectSecondary struct {
	ID                  string    `json:"id"`
	SyllabusSubjectID   string    `json:"syllabus_subject_id"`
	SeqNo               int       `json:"seq_no"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty"`
	CompletionStandard  string    `json:"completion_standard,omitempty"`
	SortOrder           int       `json:"sort_order"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// SyllabusImportSecondaryRow 大纲导入时的二级科目行
type SyllabusImportSecondaryRow struct {
	SeqNo              int
	Name               string
	Description        string
	CompletionStandard string
	SortOrder          int
}

// SyllabusImportSubjectRow 大纲导入时的一级科目行
type SyllabusImportSubjectRow struct {
	Code        string
	Name        string
	SortOrder   int
	Secondaries []SyllabusImportSecondaryRow
}

// TrainingArchiveSubject links an archive to subjects with evaluation.
type TrainingArchiveSubject struct {
	ID                string    `json:"id"`
	TrainingArchiveID string    `json:"training_archive_id"`
	SyllabusSubjectID string    `json:"syllabus_subject_id"`
	EvaluationScore   *int      `json:"evaluation_score,omitempty"`
	IsGraduated       bool      `json:"is_graduated"`
	SortOrder         int       `json:"sort_order"`
	SubjectContent    string    `json:"subject_content,omitempty"`
	Type              string    `json:"type,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NextTrainingSubject is assigned by instructor for next session.
type NextTrainingSubject struct {
	ID                string    `json:"id"`
	TrainingArchiveID string    `json:"training_archive_id"`
	SyllabusSubjectID string    `json:"syllabus_subject_id"`
	AssignedBy        string    `json:"assigned_by"`
	CreatedAt         time.Time `json:"created_at"`
}

// RawFlightImport 飞行数据导入待认领
type RawFlightImport struct {
	ID                    string     `json:"id"`
	BatchID               string     `json:"batch_id"`
	EngineOnTime          *time.Time `json:"engine_on_time,omitempty"`
	EngineOffTime         *time.Time `json:"engine_off_time,omitempty"`
	FlightDurationMinutes *int       `json:"flight_duration_minutes,omitempty"`
	IsLocal               bool       `json:"is_local"`
	RawData               []byte     `json:"raw_data,omitempty"`
	TrainingRecordID      *string    `json:"training_record_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

// TrainingRecordView 自定义视图
type TrainingRecordView struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Name       string    `json:"name"`
	FilterJSON []byte    `json:"filter_json,omitempty"`
	SortJSON   []byte    `json:"sort_json,omitempty"`
	GroupBy    string    `json:"group_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SyllabusSubjectWithGraduated 带已结业标记的科目（用于科目选择界面）
type SyllabusSubjectWithGraduated struct {
	SyllabusSubject
	IsGraduated bool                        `json:"is_graduated"`
	Status      string                      `json:"status,omitempty"`
	Secondaries []*SyllabusSubjectSecondary `json:"secondaries,omitempty"`
}
