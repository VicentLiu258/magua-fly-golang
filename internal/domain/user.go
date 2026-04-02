package domain

import (
	"strings"
	"time"
)

// UserRole represents user role in the system.
type UserRole string

const (
	RoleAdmin       UserRole = "admin"
	RoleStaff       UserRole = "staff"
	RoleMaintenance UserRole = "maintenance" // 机务，地面保障，不上机、无教员档案
	RoleInstructor  UserRole = "instructor"
	RoleStudent     UserRole = "student"
)

// EmploymentStatus 任职状态（登录与 JWT 业务接口：非 admin 须为 active）。
type EmploymentStatus string

const (
	EmploymentActive   EmploymentStatus = "active"
	EmploymentResigned EmploymentStatus = "resigned"
	EmploymentPartTime EmploymentStatus = "part_time"
)

// PersonnelCategory 人员类别（由角色推导，不落库）。
type PersonnelCategory string

const (
	PersonnelFlight        PersonnelCategory = "flight"
	PersonnelMaintenance   PersonnelCategory = "maintenance"
	PersonnelGround        PersonnelCategory = "ground"
	PersonnelCategoryAdmin PersonnelCategory = "admin"
)

// DerivePersonnelCategory 从角色推导人员类别（飞行/维修/地面）。
func DerivePersonnelCategory(r UserRole) PersonnelCategory {
	switch r {
	case RoleStudent, RoleInstructor:
		return PersonnelFlight
	case RoleMaintenance:
		return PersonnelMaintenance
	case RoleStaff:
		return PersonnelGround
	case RoleAdmin:
		return PersonnelCategoryAdmin
	default:
		return ""
	}
}

// FlightTrainingStatus 学飞状态（students.flight_training_status）。
type FlightTrainingStatus string

const (
	FlightTrainingInProgress FlightTrainingStatus = "in_progress" // 进行中
	FlightTrainingPaused     FlightTrainingStatus = "paused"      // 已暂停
	FlightTrainingCompleted  FlightTrainingStatus = "completed"   // 已完结
	FlightTrainingTerminated FlightTrainingStatus = "terminated"  // 已终止
)

// ParseFlightTrainingStatus 解析 API/DB 学飞状态码。
func ParseFlightTrainingStatus(s string) (FlightTrainingStatus, bool) {
	t := strings.TrimSpace(s)
	switch FlightTrainingStatus(t) {
	case FlightTrainingInProgress, FlightTrainingPaused, FlightTrainingCompleted, FlightTrainingTerminated:
		return FlightTrainingStatus(t), true
	default:
		return "", false
	}
}

// RoleUsesStaffSchedule 是否使用 staff_schedules 做日排班（地面工作人员，含 staff 与机务）。
func RoleUsesStaffSchedule(r UserRole) bool {
	return r == RoleStaff || r == RoleMaintenance
}

// User is the account entity.
type User struct {
	ID               string           `json:"id"`
	Username         string           `json:"username"`
	PasswordHash     string           `json:"-"`
	Phone            string           `json:"phone,omitempty"`
	Email            string           `json:"email,omitempty"`
	Role             UserRole         `json:"role"`
	EmploymentStatus EmploymentStatus `json:"employment_status"`
	AvatarURL        string           `json:"avatar_url,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
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
	ID        string     `json:"id"`
	Phone     string     `json:"phone"`
	Scene     SmsScene   `json:"scene"`
	Code      string     `json:"-"`
	IP        string     `json:"ip,omitempty"`
	ExpiredAt time.Time  `json:"expired_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Student extends user profile for trainees.
type Student struct {
	ID                      string  `json:"id"`
	UserID                  *string `json:"user_id,omitempty"`
	Name                    string  `json:"name"`
	StudentNo               string  `json:"student_no"`
	ResidentCity            string  `json:"resident_city,omitempty"`
	ContactPhone            string  `json:"contact_phone,omitempty"`
	Gender                  string  `json:"gender,omitempty"`
	BirthDate               *time.Time `json:"birth_date,omitempty"`
	IDDocumentNo            string  `json:"id_document_no,omitempty"`
	BirthPlace              string  `json:"birth_place,omitempty"`
	LicenseArchiveNo        string  `json:"license_archive_no,omitempty"`
	EducationLevel          string  `json:"education_level,omitempty"`
	PermanentAddress        string  `json:"permanent_address,omitempty"`
	MailingAddress          string  `json:"mailing_address,omitempty"`
	Region                  string  `json:"region,omitempty"`
	ServiceStatusRaw        string  `json:"service_status_raw,omitempty"`
	PauseReason             string  `json:"pause_reason,omitempty"`
	TransferBase            string  `json:"transfer_base,omitempty"`
	RegistryExtra           []byte  `json:"registry_extra,omitempty"`
	ContractInfo            []byte  `json:"contract_info,omitempty"`
	ContractDurationMinutes *int    `json:"contract_duration_minutes,omitempty"`
	FlownMinutes            int     `json:"flown_minutes"`
	TrainingProgress        string  `json:"training_progress,omitempty"`
	// TrainingProgressJSON 结构化训练进度，映射 students.training_progress_json
	TrainingProgressJSON    []byte               `json:"training_progress_json,omitempty"`
	ResponsibleInstructorID *string              `json:"responsible_instructor_id,omitempty"`
	FlightTrainingStatus    FlightTrainingStatus `json:"flight_training_status,omitempty"`
	CreatedAt               time.Time            `json:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at"`
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
