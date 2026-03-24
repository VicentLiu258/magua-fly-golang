package repository

import (
	"context"
	"time"

	"magua-fly-golang/internal/domain"
)

// UserRepository defines user persistence.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	// ListByRoles returns users with any of the given roles, ordered by role then username.
	ListByRoles(ctx context.Context, roles []domain.UserRole) ([]*domain.User, error)
}

// StudentRepository defines student persistence.
type StudentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Student, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Student, error)
	GetByStudentNo(ctx context.Context, studentNo string) (*domain.Student, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Student, error)
	Create(ctx context.Context, s *domain.Student) error
	Update(ctx context.Context, s *domain.Student) error
}

// InstructorRepository defines instructor persistence.
type InstructorRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Instructor, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Instructor, error)
	List(ctx context.Context) ([]*domain.Instructor, error)
	Create(ctx context.Context, i *domain.Instructor) error
}

// AircraftRepository defines aircraft persistence.
type AircraftRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Aircraft, error)
	List(ctx context.Context) ([]*domain.Aircraft, error)
	Create(ctx context.Context, a *domain.Aircraft) error
	UpdateAirworthiness(ctx context.Context, id string, totalHours, lastInspectionHours *float64, inspectionIntervalHours *int) error
}

// InstructorScheduleRepository defines instructor schedule (duty) persistence.
type InstructorScheduleRepository interface {
	// ListByDate returns all schedules for a given date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.InstructorSchedule, error)
	// ListByDateRange returns schedules in [from, to] inclusive by date; instructorIDs empty means all.
	ListByDateRange(ctx context.Context, from, to time.Time, instructorIDs []string) ([]*domain.InstructorSchedule, error)
	GetByInstructorAndDate(ctx context.Context, instructorID string, date time.Time) (*domain.InstructorSchedule, error)
	UpsertDay(ctx context.Context, s *domain.InstructorSchedule) error
}

// StaffScheduleRepository defines staff (non-flying) duty persistence by users.id.
type StaffScheduleRepository interface {
	ListByDate(ctx context.Context, date time.Time) ([]*domain.StaffSchedule, error)
	ListByDateRange(ctx context.Context, from, to time.Time, userIDs []string) ([]*domain.StaffSchedule, error)
	GetByUserAndDate(ctx context.Context, userID string, date time.Time) (*domain.StaffSchedule, error)
	UpsertDay(ctx context.Context, s *domain.StaffSchedule) error
}

// AircraftAvailabilityRepository defines aircraft availability persistence.
type AircraftAvailabilityRepository interface {
	// ListByDate returns all availability records for a given date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.AircraftAvailability, error)
	Create(ctx context.Context, a *domain.AircraftAvailability) error
	DeleteByMaintenanceWindowID(ctx context.Context, maintenanceWindowID string) error
}

// AircraftMaintenanceRepository 维护时间窗持久化
type AircraftMaintenanceRepository interface {
	Create(ctx context.Context, w *domain.MaintenanceWindow) error
	GetByID(ctx context.Context, id string) (*domain.MaintenanceWindow, error)
	Update(ctx context.Context, w *domain.MaintenanceWindow) error
	// ListByAircraft aircraftID 为空时不过滤飞机（返回全部）；from/to/status 仍按原逻辑可选
	ListByAircraft(ctx context.Context, aircraftID string, from, to *time.Time, status *domain.MaintenanceWindowStatus) ([]*domain.MaintenanceWindow, error)
}

// MaintenanceWindowSyncer 维护窗口事务（创建/更新/取消并同步排班）
type MaintenanceWindowSyncer interface {
	CreateMaintenanceWindowWithSync(ctx context.Context, aircraftID, kind string, startAt, endAt time.Time, hangar, notes string, syncSchedule bool, reviewerID string) (*domain.MaintenanceWindowSyncResult, error)
	UpdateMaintenanceWindowWithSync(ctx context.Context, windowID, aircraftID, kind string, startAt, endAt time.Time, hangar, notes string, syncSchedule bool, reviewerID string) (*domain.MaintenanceWindowSyncResult, error)
	CancelMaintenanceWindow(ctx context.Context, windowID string) error
}

// BookingRepository defines booking persistence.
type BookingRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	ListByStudent(ctx context.Context, studentID string, status *domain.BookingStatus) ([]*domain.Booking, error)
	ListPending(ctx context.Context) ([]*domain.Booking, error)
	Create(ctx context.Context, b *domain.Booking) error
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus, rejectReason string, reviewerID string) error
	// ListByDate returns all bookings on a specific plan date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.Booking, error)
	// ListByDateRange returns bookings with plan_date in [from, to] inclusive (calendar dates).
	ListByDateRange(ctx context.Context, from, to time.Time) ([]*domain.Booking, error)
	// ListByAircraftAndDateRange 某飞机在日期区间内的预约（用于维护冲突处理）
	ListByAircraftAndDateRange(ctx context.Context, aircraftID string, from, to time.Time) ([]*domain.Booking, error)
}

// TrainingPlanRepository defines training plan persistence.
type TrainingPlanRepository interface {
	Create(ctx context.Context, p *domain.TrainingPlan) error
	ListByDate(ctx context.Context, from, to string) ([]*domain.TrainingPlan, error)
	DeleteMatchingBookingFields(ctx context.Context, studentID, instructorID, aircraftID string, planDate time.Time, startTime, endTime time.Time) (deletedID *string, err error)
}

// TrainingRecordRepository defines training record persistence.
type TrainingRecordRepository interface {
	GetByID(ctx context.Context, id string) (*domain.TrainingRecord, error)
	ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingRecord, error)
	List(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit, offset int) ([]*domain.TrainingRecord, int, error)
	Create(ctx context.Context, r *domain.TrainingRecord) error
	Update(ctx context.Context, r *domain.TrainingRecord) error
	UpdateStatus(ctx context.Context, id string, status domain.TrainingRecordStatus) error
}

// TrainingArchiveRepository defines training archive persistence.
type TrainingArchiveRepository interface {
	GetByID(ctx context.Context, id string) (*domain.TrainingArchive, error)
	GetByTrainingRecordID(ctx context.Context, recordID string) (*domain.TrainingArchive, error)
	ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingArchive, error)
	Create(ctx context.Context, a *domain.TrainingArchive) error
	Update(ctx context.Context, a *domain.TrainingArchive) error
}

// SyllabusSubjectRepository defines syllabus subject persistence.
type SyllabusSubjectRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SyllabusSubject, error)
	List(ctx context.Context) ([]*domain.SyllabusSubject, error)
	Create(ctx context.Context, s *domain.SyllabusSubject) error
	Update(ctx context.Context, s *domain.SyllabusSubject) error
	Delete(ctx context.Context, id string) error
}

// RawFlightImportRepository defines raw flight import persistence.
type RawFlightImportRepository interface {
	Create(ctx context.Context, r *domain.RawFlightImport) error
	ListPending(ctx context.Context, limit, offset int) ([]*domain.RawFlightImport, error)
	GetByID(ctx context.Context, id string) (*domain.RawFlightImport, error)
	SetTrainingRecordID(ctx context.Context, id string, recordID string) error
}

// TrainingRecordViewRepository defines training record view persistence.
type TrainingRecordViewRepository interface {
	Create(ctx context.Context, v *domain.TrainingRecordView) error
	ListByUserID(ctx context.Context, userID string) ([]*domain.TrainingRecordView, error)
	GetByID(ctx context.Context, id string) (*domain.TrainingRecordView, error)
	Update(ctx context.Context, v *domain.TrainingRecordView) error
	Delete(ctx context.Context, id string) error
}

// TrainingArchiveSubjectRepository defines training archive subject persistence.
type TrainingArchiveSubjectRepository interface {
	ListByArchiveID(ctx context.Context, archiveID string) ([]*domain.TrainingArchiveSubject, error)
	ReplaceForArchive(ctx context.Context, archiveID string, subjects []*domain.TrainingArchiveSubject) error
}

// NextTrainingSubjectRepository defines next training subject persistence.
type NextTrainingSubjectRepository interface {
	ListByArchiveID(ctx context.Context, archiveID string) ([]*domain.NextTrainingSubject, error)
	ReplaceForArchive(ctx context.Context, archiveID string, assignedBy string, syllabusSubjectIDs []string) error
}

// TaskRepository defines task persistence.
type TaskRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	ListByStudent(ctx context.Context, studentID string, status *domain.TaskStatus) ([]*domain.Task, error)
	Create(ctx context.Context, t *domain.Task) error
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error
	// FindByTypeAndRefID 根据任务类型与业务引用 ID 查找单条任务，用于幂等创建。
	FindByTypeAndRefID(ctx context.Context, taskType domain.TaskType, refID string) (*domain.Task, error)
}

// TheoryDocumentRepository defines theory document persistence.
type TheoryDocumentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.TheoryDocument, error)
	List(ctx context.Context) ([]*domain.TheoryDocument, error)
	Create(ctx context.Context, d *domain.TheoryDocument) error
	Update(ctx context.Context, d *domain.TheoryDocument) error
	Delete(ctx context.Context, id string) error
}

// TheoryAttemptRepository defines theory attempt persistence.
type TheoryAttemptRepository interface {
	Create(ctx context.Context, a *domain.TheoryAttempt) error
	UpdateOnSubmit(ctx context.Context, a *domain.TheoryAttempt) error
	ListByUserAndDocument(ctx context.Context, userID, documentID string, limit, offset int) ([]*domain.TheoryAttempt, error)
}

// TheoryPaperRepository defines theory paper (exam) persistence.
type TheoryPaperRepository interface {
	Create(ctx context.Context, p *domain.TheoryPaper) error
	Update(ctx context.Context, p *domain.TheoryPaper) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.TheoryPaper, error)
	ListByDocumentID(ctx context.Context, documentID string, limit, offset int) ([]*domain.TheoryPaper, int, error)
}

// TheoryPaperQuestionRepository defines theory paper-question link persistence.
type TheoryPaperQuestionRepository interface {
	ReplaceForPaper(ctx context.Context, paperID string, questions []*domain.TheoryPaperQuestion) error
	ListByPaperID(ctx context.Context, paperID string) ([]*domain.TheoryPaperQuestion, error)
}

// SubjectLearningRepository defines subject learning (per syllabus subject) persistence.
type SubjectLearningRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SubjectLearning, error)
	List(ctx context.Context) ([]*domain.SubjectLearning, error)
}

// SubjectAttemptRepository defines subject exam attempt persistence (summary in PostgreSQL).
type SubjectAttemptRepository interface {
	Create(ctx context.Context, a *domain.SubjectAttempt) error
	UpdateOnSubmit(ctx context.Context, a *domain.SubjectAttempt) error
	ListByUserAndSubjectLearning(ctx context.Context, userID, subjectLearningID string, limit, offset int) ([]*domain.SubjectAttempt, error)
}

// ContractRepository defines student contract persistence.
type ContractRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Contract, error)
	ListByStudent(ctx context.Context, studentID string) ([]*domain.Contract, error)
	Create(ctx context.Context, c *domain.Contract) error
	Update(ctx context.Context, c *domain.Contract) error
}

// SmsVerificationRepository defines SMS verification code persistence.
type SmsVerificationRepository interface {
	Create(ctx context.Context, v *domain.SmsVerificationCode) error
	// FindLatestByPhoneScene returns the latest code (used or unused) for rate limiting checks.
	FindLatestByPhoneScene(ctx context.Context, phone string, scene domain.SmsScene) (*domain.SmsVerificationCode, error)
	// FindValidCode looks up a not-yet-used, not-expired code.
	FindValidCode(ctx context.Context, phone string, scene domain.SmsScene, code string, now time.Time) (*domain.SmsVerificationCode, error)
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
}
