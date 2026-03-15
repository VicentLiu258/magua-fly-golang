package domain

import "time"

// ImportSource for training record.
type ImportSource string

const (
	ImportManual ImportSource = "manual"
	ImportImport ImportSource = "import"
)

// TrainingRecordStatus 训练记录状态
type TrainingRecordStatus string

const (
	RecordStatusNormal   TrainingRecordStatus = "normal"
	RecordStatusCancelled TrainingRecordStatus = "cancelled"
	RecordStatusDeleted   TrainingRecordStatus = "deleted"
	RecordStatusExpired   TrainingRecordStatus = "expired"
)

// TrainingRecord is one flight training record.
type TrainingRecord struct {
	ID                     int64                 `json:"id"`
	StudentID              int64                 `json:"student_id"`
	InstructorID           int64                 `json:"instructor_id"`
	AircraftID             int64                 `json:"aircraft_id"`
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
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

// TrainingRecordListFilter 列表筛选条件
type TrainingRecordListFilter struct {
	StudentID    *int64     `json:"student_id,omitempty"`
	InstructorID *int64     `json:"instructor_id,omitempty"`
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

// TrainingArchive is the detailed archive for one training record.
type TrainingArchive struct {
	ID                            int64      `json:"id"`
	TrainingRecordID              int64      `json:"training_record_id"`
	StudentID                     int64      `json:"student_id"`
	InstructorID                  int64      `json:"instructor_id"`
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
	ResponsibleInstructorID       *int64     `json:"responsible_instructor_id,omitempty"`
	ChiefInstructorID             *int64     `json:"chief_instructor_id,omitempty"`
	CreatedAt                     time.Time  `json:"created_at"`
	UpdatedAt                     time.Time  `json:"updated_at"`
}

// SyllabusSubject is a subject in the training outline.
type SyllabusSubject struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TrainingArchiveSubject links an archive to subjects with evaluation.
type TrainingArchiveSubject struct {
	ID                int64     `json:"id"`
	TrainingArchiveID int64     `json:"training_archive_id"`
	SyllabusSubjectID int64     `json:"syllabus_subject_id"`
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
	ID                int64     `json:"id"`
	TrainingArchiveID int64     `json:"training_archive_id"`
	SyllabusSubjectID int64     `json:"syllabus_subject_id"`
	AssignedBy        int64     `json:"assigned_by"`
	CreatedAt         time.Time `json:"created_at"`
}

// RawFlightImport 飞行数据导入待认领
type RawFlightImport struct {
	ID                    int64      `json:"id"`
	BatchID               string     `json:"batch_id"`
	EngineOnTime          *time.Time `json:"engine_on_time,omitempty"`
	EngineOffTime        *time.Time `json:"engine_off_time,omitempty"`
	FlightDurationMinutes *int       `json:"flight_duration_minutes,omitempty"`
	IsLocal               bool      `json:"is_local"`
	RawData               []byte     `json:"raw_data,omitempty"`
	TrainingRecordID      *int64     `json:"training_record_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

// TrainingRecordView 自定义视图
type TrainingRecordView struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
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
	IsGraduated bool `json:"is_graduated"`
}
