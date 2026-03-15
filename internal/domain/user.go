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
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Phone        string    `json:"phone,omitempty"`
	Email        string    `json:"email,omitempty"`
	Role         UserRole  `json:"role"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SmsScene represents the scenario for SMS verification codes.
type SmsScene string

const (
	SmsSceneRegister SmsScene = "register"
	SmsSceneLogin    SmsScene = "login"
	SmsSceneResetPwd SmsScene = "reset_pwd"
	SmsScene2FA      SmsScene = "2fa"
)

// SmsVerificationCode records a verification code sent to a phone for a given scene.
type SmsVerificationCode struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	Scene     SmsScene  `json:"scene"`
	Code      string    `json:"-"`
	IP        string    `json:"ip,omitempty"`
	ExpiredAt time.Time `json:"expired_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Student extends user profile for trainees.
type Student struct {
	ID                      string    `json:"id"`
	UserID                  *string   `json:"user_id,omitempty"`
	Name                    string    `json:"name"`
	StudentNo               string    `json:"student_no"`
	ContractInfo            []byte    `json:"contract_info,omitempty"`
	ContractDurationMinutes *int      `json:"contract_duration_minutes,omitempty"`
	FlownMinutes            int       `json:"flown_minutes"`
	TrainingProgress        string    `json:"training_progress,omitempty"`
	ResponsibleInstructorID *string   `json:"responsible_instructor_id,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// Instructor extends user for instructors.
type Instructor struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	DisplayColor string    `json:"display_color,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
