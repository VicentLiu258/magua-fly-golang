package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/pkg/jwt"
	"magua-fly-golang/internal/pkg/password"
	"magua-fly-golang/internal/pkg/sms"
	"magua-fly-golang/internal/repository"
)

// AuthUseCase handles login and token issuance.
type AuthUseCase struct {
	userRepo         repository.UserRepository
	studentRepo      repository.StudentRepository
	instructorRepo   repository.InstructorRepository
	smsRepo          repository.SmsVerificationRepository
	jwtManager       *jwt.Manager
	smsClient        sms.Client
	smsTemplateLogin   string
	smsTemplateRegister string
}

// NewAuthUseCase creates an AuthUseCase.
func NewAuthUseCase(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	instructorRepo repository.InstructorRepository,
	smsRepo repository.SmsVerificationRepository,
	jwtManager *jwt.Manager,
	smsClient sms.Client,
	smsTemplateLogin string,
	smsTemplateRegister string,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:            userRepo,
		studentRepo:         studentRepo,
		instructorRepo:      instructorRepo,
		smsRepo:             smsRepo,
		jwtManager:          jwtManager,
		smsClient:           smsClient,
		smsTemplateLogin:    smsTemplateLogin,
		smsTemplateRegister: smsTemplateRegister,
	}
}

// LoginRequest is the login input.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse returns the token and user info.
type LoginResponse struct {
	Token string      `json:"token"`
	User  *domain.User `json:"user"`
}

// Login validates credentials and returns a JWT.
func (u *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !password.Check(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}
	if user.Role != domain.RoleAdmin && user.EmploymentStatus != domain.EmploymentActive {
		return nil, ErrInvalidCredentials
	}
	user.PasswordHash = ""
	token, err := u.jwtManager.Issue(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}
	return &LoginResponse{Token: token, User: user}, nil
}

// RegisterRequest is the registration input.
type RegisterRequest struct {
	Username string          `json:"username" binding:"required"`
	Password string          `json:"password" binding:"required"`
	Phone    string          `json:"phone"`
	Email    string          `json:"email"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

// SmsScene is an alias of domain.SmsScene for JSON binding.
type SmsScene = domain.SmsScene

// Sms scenes
const (
	SmsSceneRegister = domain.SmsSceneRegister
	SmsSceneLogin    = domain.SmsSceneLogin
	SmsSceneResetPwd = domain.SmsSceneResetPwd
	SmsScene2FA      = domain.SmsScene2FA
)

// SendSmsCodeRequest is the input for sending an SMS verification code.
type SendSmsCodeRequest struct {
	Phone string    `json:"phone" binding:"required"`
	Scene SmsScene  `json:"scene" binding:"required"`
	IP    string    `json:"-"`
}

// VerifySmsCodeRequest is the input for verifying an SMS code.
type VerifySmsCodeRequest struct {
	Phone string   `json:"phone" binding:"required"`
	Scene SmsScene `json:"scene" binding:"required"`
	Code  string   `json:"code" binding:"required"`
}

// SmsVerificationResult represents result of verifying a code.
type SmsVerificationResult struct {
	User  *domain.User
	Token string
}

// SendSmsCode sends a verification code, with basic rate limiting.
func (u *AuthUseCase) SendSmsCode(ctx context.Context, req SendSmsCodeRequest) error {
	if req.Phone == "" {
		return fmt.Errorf("phone required")
	}

	now := time.Now()

	// Simple rate limit: at most one per 60s for same phone+scene.
	if latest, err := u.smsRepo.FindLatestByPhoneScene(ctx, req.Phone, req.Scene); err == nil && latest != nil {
		if now.Sub(latest.CreatedAt) < 60*time.Second {
			return fmt.Errorf("too many requests, please try later")
		}
	}

	// Basic 6-digit code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiredAt := now.Add(5 * time.Minute)

	record := &domain.SmsVerificationCode{
		Phone:     req.Phone,
		Scene:     req.Scene,
		Code:      code,
		IP:        req.IP,
		ExpiredAt: expiredAt,
	}
	if err := u.smsRepo.Create(ctx, record); err != nil {
		return err
	}

	templateID := ""
	switch req.Scene {
	case SmsSceneLogin:
		templateID = u.smsTemplateLogin
	case SmsSceneRegister:
		templateID = u.smsTemplateRegister
	default:
		templateID = u.smsTemplateLogin
	}

	// Tencent Cloud requires E.164 format, assume mainland China by default.
	phoneE164 := req.Phone
	if len(req.Phone) > 0 && req.Phone[0] != '+' {
		phoneE164 = "+86" + req.Phone
	}

	if err := u.smsClient.SendCode(ctx, phoneE164, templateID, []string{code}); err != nil {
		return err
	}
	return nil
}

// VerifySmsCode verifies a code and returns user + jwt token for login/register scenes.
func (u *AuthUseCase) VerifySmsCode(ctx context.Context, req VerifySmsCodeRequest) (*SmsVerificationResult, error) {
	now := time.Now()
	record, err := u.smsRepo.FindValidCode(ctx, req.Phone, req.Scene, req.Code, now)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := u.smsRepo.MarkUsed(ctx, record.ID, now); err != nil {
		return nil, err
	}

	switch req.Scene {
	case SmsSceneLogin, SmsSceneRegister:
		var user *domain.User
		user, err = u.userRepo.GetByPhone(ctx, req.Phone)
		if err != nil {
			if req.Scene == SmsSceneRegister {
				// Auto-register a basic student account if not exists.
				user = &domain.User{
					Username:     req.Phone,
					Phone:        req.Phone,
					Role:         domain.RoleStudent,
					PasswordHash: "",
				}
				if err := u.userRepo.Create(ctx, user); err != nil {
					return nil, err
				}
				// Auto-create linked student profile for SMS-registered students.
				if err := u.ensureProfileForUser(ctx, user); err != nil {
					return nil, err
				}
			} else {
				return nil, ErrInvalidCredentials
			}
		}
		if user.Role != domain.RoleAdmin && user.EmploymentStatus != domain.EmploymentActive {
			return nil, ErrInvalidCredentials
		}
		user.PasswordHash = ""
		token, err := u.jwtManager.Issue(user.ID, string(user.Role))
		if err != nil {
			return nil, err
		}
		return &SmsVerificationResult{
			User:  user,
			Token: token,
		}, nil
	case SmsSceneResetPwd, SmsScene2FA:
		// For now just return nil user/token; callers can treat success as a flag.
		return &SmsVerificationResult{}, nil
	default:
		return nil, errors.New("unsupported sms scene")
	}
}

// Register creates a new user.
func (u *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Username:     req.Username,
		PasswordHash: hash,
		Phone:        req.Phone,
		Email:        req.Email,
		Role:         req.Role,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	// Create linked profile in students / instructors based on role.
	if err := u.ensureProfileForUser(ctx, user); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// ensureProfileForUser creates student/instructor profile rows linked to the given user when needed.
func (u *AuthUseCase) ensureProfileForUser(ctx context.Context, user *domain.User) error {
	switch user.Role {
	case domain.RoleStudent:
		// Generate a simple student_no based on user ID (UUID).
		s := &domain.Student{
			UserID:       &user.ID,
			Name:         user.Username,
			StudentNo:    "S-" + user.ID,
			FlownMinutes: 0,
		}
		return u.studentRepo.Create(ctx, s)
	case domain.RoleInstructor:
		i := &domain.Instructor{
			UserID:       user.ID,
			Name:         user.Username,
			DisplayColor: "",
		}
		return u.instructorRepo.Create(ctx, i)
	default:
		return nil
	}
}

// GetMe 返回当前账号信息（不含密码哈希）。
func (u *AuthUseCase) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// UpdateProfileRequest 用户自助更新资料（不可修改 role、employment_status，该两项仅管理员通过人员管理接口维护）。
type UpdateProfileRequest struct {
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Password  string `json:"password"`
}

// UpdateProfile 更新当前用户资料。
func (u *AuthUseCase) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		if err := u.checkUsernameAvailableForProfile(ctx, req.Username, user.ID); err != nil {
			return nil, err
		}
		user.Username = strings.TrimSpace(req.Username)
	}
	if req.Phone != "" {
		if err := u.checkPhoneAvailable(ctx, req.Phone, user.ID); err != nil {
			return nil, err
		}
		user.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Email != "" {
		user.Email = strings.TrimSpace(req.Email)
	}
	if req.AvatarURL != "" {
		user.AvatarURL = strings.TrimSpace(req.AvatarURL)
	}
	if req.Password != "" {
		h, err := password.Hash(req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = h
	}
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (u *AuthUseCase) checkUsernameAvailableForProfile(ctx context.Context, username, excludeUserID string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}
	other, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if excludeUserID != "" && other.ID == excludeUserID {
		return nil
	}
	return ErrUsernameConflict
}

func (u *AuthUseCase) checkPhoneAvailable(ctx context.Context, phone, excludeUserID string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil
	}
	other, err := u.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if excludeUserID != "" && other.ID == excludeUserID {
		return nil
	}
	return ErrPhoneConflict
}
