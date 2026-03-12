package domain

import "time"

// ImportSource for training record.
type ImportSource string

const (
	ImportManual  ImportSource = "manual"
	ImportImport  ImportSource = "import"
)

// TrainingRecord is one flight training record.
type TrainingRecord struct {
	ID                    int64        `json:"id"`
	StudentID              int64        `json:"student_id"`
	InstructorID          int64        `json:"instructor_id"`
	AircraftID            int64        `json:"aircraft_id"`
	TrainingDate          time.Time    `json:"training_date"`
	EngineOnTime          *time.Time   `json:"engine_on_time,omitempty"`
	EngineOffTime         *time.Time   `json:"engine_off_time,omitempty"`
	FlightDurationMinutes *int         `json:"flight_duration_minutes,omitempty"`
	IsLocal               bool         `json:"is_local"`
	ImportSource          ImportSource `json:"import_source"`
	RawImportID           string       `json:"raw_import_id,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}

// TrainingArchive is the detailed archive for one training record.
type TrainingArchive struct {
	ID                           int64      `json:"id"`
	TrainingRecordID             int64      `json:"training_record_id"`
	StudentID                    int64      `json:"student_id"`
	InstructorID                 int64      `json:"instructor_id"`
	TrainingDate                 time.Time  `json:"training_date"`
	Stage                        string     `json:"stage,omitempty"`
	PlannedDurationMinutes       *int       `json:"planned_duration_minutes,omitempty"`
	TotalExperienceMinutes       *int       `json:"total_experience_minutes,omitempty"`
	RemainingMinutes             *int       `json:"remaining_minutes,omitempty"`
	EngineOnTime                 *time.Time `json:"engine_on_time,omitempty"`
	EngineOffTime                *time.Time `json:"engine_off_time,omitempty"`
	FlightDurationMinutes        *int       `json:"flight_duration_minutes,omitempty"`
	ProgressNote                 string     `json:"progress_note,omitempty"`
	InstructorReview             string     `json:"instructor_review,omitempty"`
	StudentSelfReview            string     `json:"student_self_review,omitempty"`
	TaskAssignment              string     `json:"task_assignment,omitempty"`
	StudentSignedAt              *time.Time `json:"student_signed_at,omitempty"`
	ResponsibleInstructorSignedAt *time.Time `json:"responsible_instructor_signed_at,omitempty"`
	ChiefInstructorSignedAt      *time.Time `json:"chief_instructor_signed_at,omitempty"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
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
	ID                 int64     `json:"id"`
	TrainingArchiveID   int64     `json:"training_archive_id"`
	SyllabusSubjectID   int64     `json:"syllabus_subject_id"`
	EvaluationScore    *int      `json:"evaluation_score,omitempty"`
	IsGraduated        bool      `json:"is_graduated"`
	SortOrder          int       `json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// NextTrainingSubject is assigned by instructor for next session.
type NextTrainingSubject struct {
	ID                 int64     `json:"id"`
	TrainingArchiveID  int64     `json:"training_archive_id"`
	SyllabusSubjectID  int64     `json:"syllabus_subject_id"`
	AssignedBy         int64     `json:"assigned_by"`
	CreatedAt          time.Time `json:"created_at"`
}
