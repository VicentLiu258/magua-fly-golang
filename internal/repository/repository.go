package repository

import (
	"context"

	"magua-fly-golang/internal/domain"
)

// UserRepository defines user persistence.
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
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

// BookingRepository defines booking persistence.
type BookingRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	ListByStudent(ctx context.Context, studentID int64, status *domain.BookingStatus) ([]*domain.Booking, error)
	ListPending(ctx context.Context) ([]*domain.Booking, error)
	Create(ctx context.Context, b *domain.Booking) error
	UpdateStatus(ctx context.Context, id int64, status domain.BookingStatus, rejectReason string, reviewerID int64) error
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
	Create(ctx context.Context, r *domain.TrainingRecord) error
	Update(ctx context.Context, r *domain.TrainingRecord) error
}

// TrainingArchiveRepository defines training archive persistence.
type TrainingArchiveRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.TrainingArchive, error)
	GetByTrainingRecordID(ctx context.Context, recordID int64) (*domain.TrainingArchive, error)
	Create(ctx context.Context, a *domain.TrainingArchive) error
	Update(ctx context.Context, a *domain.TrainingArchive) error
}

// SyllabusSubjectRepository defines syllabus subject persistence.
type SyllabusSubjectRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.SyllabusSubject, error)
	List(ctx context.Context) ([]*domain.SyllabusSubject, error)
	Create(ctx context.Context, s *domain.SyllabusSubject) error
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
