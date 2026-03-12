package domain

import "time"

// UserRole represents user role in the system.
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleStaff      UserRole = "staff"
	RoleInstructor UserRole = "instructor"
	RoleStudent    UserRole = "student"
)

// User is the account entity.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Phone        string    `json:"phone,omitempty"`
	Email        string    `json:"email,omitempty"`
	Role         UserRole  `json:"role"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Student extends user profile for trainees.
type Student struct {
	ID                      int64     `json:"id"`
	UserID                  *int64    `json:"user_id,omitempty"`
	Name                    string    `json:"name"`
	StudentNo               string    `json:"student_no"`
	ContractInfo            []byte    `json:"contract_info,omitempty"`
	ContractDurationMinutes *int      `json:"contract_duration_minutes,omitempty"`
	FlownMinutes            int       `json:"flown_minutes"`
	TrainingProgress        string    `json:"training_progress,omitempty"`
	ResponsibleInstructorID *int64    `json:"responsible_instructor_id,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// Instructor extends user for instructors.
type Instructor struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Name         string    `json:"name"`
	DisplayColor string    `json:"display_color,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
