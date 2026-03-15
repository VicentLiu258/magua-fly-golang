package repository

import (
	"context"
	"time"

	"magua-fly-golang/internal/domain"
)

// UserRepository defines user persistence.
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
}

// StudentRepository defines student persistence.
type StudentRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Student, error)
	GetByStudentNo(ctx context.Context, studentNo string) (*domain.Student, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Student, error)
	Create(ctx context.Context, s *domain.Student) error
	Update(ctx context.Context, s *domain.Student) error
}

// InstructorRepository defines instructor persistence.
type InstructorRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Instructor, error)
	List(ctx context.Context) ([]*domain.Instructor, error)
	Create(ctx context.Context, i *domain.Instructor) error
}

// AircraftRepository defines aircraft persistence.
type AircraftRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Aircraft, error)
	List(ctx context.Context) ([]*domain.Aircraft, error)
	Create(ctx context.Context, a *domain.Aircraft) error
}

// InstructorScheduleRepository defines instructor schedule (duty) persistence.
type InstructorScheduleRepository interface {
	// ListByDate returns all schedules for a given date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.InstructorSchedule, error)
}

// AircraftAvailabilityRepository defines aircraft availability persistence.
type AircraftAvailabilityRepository interface {
	// ListByDate returns all availability records for a given date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.AircraftAvailability, error)
	Create(ctx context.Context, a *domain.AircraftAvailability) error
}

// BookingRepository defines booking persistence.
type BookingRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	ListByStudent(ctx context.Context, studentID int64, status *domain.BookingStatus) ([]*domain.Booking, error)
	ListPending(ctx context.Context) ([]*domain.Booking, error)
	Create(ctx context.Context, b *domain.Booking) error
	UpdateStatus(ctx context.Context, id int64, status domain.BookingStatus, rejectReason string, reviewerID int64) error
	// ListByDate returns all bookings on a specific plan date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.Booking, error)
}

// TrainingPlanRepository defines training plan persistence.
type TrainingPlanRepository interface {
	Create(ctx context.Context, p *domain.TrainingPlan) error
	ListByDate(ctx context.Context, from, to string) ([]*domain.TrainingPlan, error)
}

// TrainingRecordRepository defines training record persistence.
type TrainingRecordRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.TrainingRecord, error)
	ListByStudent(ctx context.Context, studentID int64, limit, offset int) ([]*domain.TrainingRecord, error)
	List(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit, offset int) ([]*domain.TrainingRecord, int, error)
	Create(ctx context.Context, r *domain.TrainingRecord) error
	Update(ctx context.Context, r *domain.TrainingRecord) error
	UpdateStatus(ctx context.Context, id int64, status domain.TrainingRecordStatus) error
}

// TrainingArchiveRepository defines training archive persistence.
type TrainingArchiveRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.TrainingArchive, error)
	GetByTrainingRecordID(ctx context.Context, recordID int64) (*domain.TrainingArchive, error)
	ListByStudent(ctx context.Context, studentID int64, limit, offset int) ([]*domain.TrainingArchive, error)
	Create(ctx context.Context, a *domain.TrainingArchive) error
	Update(ctx context.Context, a *domain.TrainingArchive) error
}

// SyllabusSubjectRepository defines syllabus subject persistence.
type SyllabusSubjectRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.SyllabusSubject, error)
	List(ctx context.Context) ([]*domain.SyllabusSubject, error)
	Create(ctx context.Context, s *domain.SyllabusSubject) error
}

// RawFlightImportRepository defines raw flight import persistence.
type RawFlightImportRepository interface {
	Create(ctx context.Context, r *domain.RawFlightImport) error
	ListPending(ctx context.Context, limit, offset int) ([]*domain.RawFlightImport, error)
	GetByID(ctx context.Context, id int64) (*domain.RawFlightImport, error)
	SetTrainingRecordID(ctx context.Context, id int64, recordID int64) error
}

// TrainingRecordViewRepository defines training record view persistence.
type TrainingRecordViewRepository interface {
	Create(ctx context.Context, v *domain.TrainingRecordView) error
	ListByUserID(ctx context.Context, userID int64) ([]*domain.TrainingRecordView, error)
	GetByID(ctx context.Context, id int64) (*domain.TrainingRecordView, error)
	Update(ctx context.Context, v *domain.TrainingRecordView) error
	Delete(ctx context.Context, id int64) error
}

// TrainingArchiveSubjectRepository defines training archive subject persistence.
type TrainingArchiveSubjectRepository interface {
	ListByArchiveID(ctx context.Context, archiveID int64) ([]*domain.TrainingArchiveSubject, error)
	ReplaceForArchive(ctx context.Context, archiveID int64, subjects []*domain.TrainingArchiveSubject) error
}

// NextTrainingSubjectRepository defines next training subject persistence.
type NextTrainingSubjectRepository interface {
	ListByArchiveID(ctx context.Context, archiveID int64) ([]*domain.NextTrainingSubject, error)
	ReplaceForArchive(ctx context.Context, archiveID int64, assignedBy int64, syllabusSubjectIDs []int64) error
}

// TaskRepository defines task persistence.
type TaskRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	ListByStudent(ctx context.Context, studentID int64, status *domain.TaskStatus) ([]*domain.Task, error)
	Create(ctx context.Context, t *domain.Task) error
	UpdateStatus(ctx context.Context, id int64, status domain.TaskStatus) error
}

// TheoryDocumentRepository defines theory document persistence.
type TheoryDocumentRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.TheoryDocument, error)
	List(ctx context.Context) ([]*domain.TheoryDocument, error)
	Create(ctx context.Context, d *domain.TheoryDocument) error
}

// SmsVerificationRepository defines SMS verification code persistence.
type SmsVerificationRepository interface {
	Create(ctx context.Context, v *domain.SmsVerificationCode) error
	// FindLatestByPhoneScene returns the latest code (used or unused) for rate limiting checks.
	FindLatestByPhoneScene(ctx context.Context, phone string, scene domain.SmsScene) (*domain.SmsVerificationCode, error)
	// FindValidCode looks up a not-yet-used, not-expired code.
	FindValidCode(ctx context.Context, phone string, scene domain.SmsScene, code string, now time.Time) (*domain.SmsVerificationCode, error)
	MarkUsed(ctx context.Context, id int64, usedAt time.Time) error
}
