package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/pkg/password"
	"magua-fly-golang/internal/repository"
)

// SavePersonnelRequest 统一人员新建/编辑（管理员）。
// 亦被 POST /api/v1/admin/students/save 使用：调用方将 Kind 设为 student。
type SavePersonnelRequest struct {
	Kind string `json:"kind"`

	StudentID    *string `json:"student_id"`
	InstructorID *string `json:"instructor_id"`
	UserID       *string `json:"user_id"`

	Name       string `json:"name"`
	StudentNo  string `json:"student_no"`
	Username   string `json:"username"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	AvatarURL  string `json:"avatar_url"`
	DisplayColor string `json:"display_color"`

	ContractDurationMinutes *int            `json:"contract_duration_minutes"`
	FlownMinutes            int             `json:"flown_minutes"`
	TrainingProgress        string          `json:"training_progress"`
	ResponsibleInstructorID *string         `json:"responsible_instructor_id"`
	ContractInfo            json.RawMessage `json:"contract_info"`
	TrainingProgressJSON    json.RawMessage `json:"training_progress_json"`
	// EmploymentStatus 任职状态：active | resigned | part_time；空表示新建默认 active，更新时不改。
	EmploymentStatus string `json:"employment_status"`
}

// SavePersonnelResult 保存成功后的响应体（按 Kind 仅一个非空）。
type SavePersonnelResult struct {
	Kind        string             `json:"kind"`
	Student     *domain.Student    `json:"student,omitempty"`
	Instructor  *domain.Instructor `json:"instructor,omitempty"`
	User        *domain.User       `json:"user,omitempty"`
}

func mergeJSONColumn(dst *[]byte, raw json.RawMessage, invalidErr error) error {
	b := bytes.TrimSpace(raw)
	if len(b) == 0 {
		return nil
	}
	if !json.Valid(b) {
		return invalidErr
	}
	*dst = append([]byte(nil), b...)
	return nil
}

func mergeContractInfoJSON(dst *[]byte, raw json.RawMessage) error {
	return mergeJSONColumn(dst, raw, ErrInvalidContractInfo)
}

func mergeTrainingProgressJSON(dst *[]byte, raw json.RawMessage) error {
	return mergeJSONColumn(dst, raw, ErrInvalidTrainingProgressJSON)
}

func employmentStatusFromRequest(raw string) (domain.EmploymentStatus, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return "", nil
	}
	switch s {
	case "active":
		return domain.EmploymentActive, nil
	case "resigned":
		return domain.EmploymentResigned, nil
	case "part_time", "part-time":
		return domain.EmploymentPartTime, nil
	default:
		return "", ErrInvalidArgument
	}
}

func (u *PersonnelUseCase) checkUsernameAvailable(ctx context.Context, username, excludeUserID string) error {
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

// SavePersonnel 按 Kind 创建或更新学员 / 教员 / 地面职员账号。
func (u *PersonnelUseCase) SavePersonnel(ctx context.Context, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	kind := strings.TrimSpace(strings.ToLower(req.Kind))
	switch kind {
	case "student":
		return u.saveStudent(ctx, req)
	case "instructor":
		return u.saveInstructor(ctx, req)
	case "staff":
		return u.saveStaff(ctx, req)
	default:
		return nil, ErrInvalidArgument
	}
}

func (u *PersonnelUseCase) saveStudent(ctx context.Context, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	sid := ""
	if req.StudentID != nil {
		sid = strings.TrimSpace(*req.StudentID)
	}
	if sid == "" {
		return u.createStudent(ctx, req)
	}
	return u.updateStudent(ctx, sid, req)
}

func (u *PersonnelUseCase) createStudent(ctx context.Context, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.StudentNo) == "" {
		return nil, ErrInvalidArgument
	}
	if req.Username != "" {
		if err := u.checkUsernameAvailable(ctx, req.Username, ""); err != nil {
			return nil, err
		}
	}
	uu := &domain.User{
		Username:  strings.TrimSpace(req.Username),
		Phone:     strings.TrimSpace(req.Phone),
		Email:     strings.TrimSpace(req.Email),
		Role:      domain.RoleStudent,
		AvatarURL: strings.TrimSpace(req.AvatarURL),
	}
	if es, err := employmentStatusFromRequest(req.EmploymentStatus); err != nil {
		return nil, err
	} else if es != "" {
		uu.EmploymentStatus = es
	}
	if req.Password != "" {
		h, err := password.Hash(req.Password)
		if err != nil {
			return nil, err
		}
		uu.PasswordHash = h
	}
	if err := u.userRepo.Create(ctx, uu); err != nil {
		return nil, err
	}
	s := &domain.Student{
		UserID:                  &uu.ID,
		Name:                    strings.TrimSpace(req.Name),
		StudentNo:               strings.TrimSpace(req.StudentNo),
		ContractDurationMinutes: req.ContractDurationMinutes,
		FlownMinutes:            req.FlownMinutes,
		TrainingProgress:        strings.TrimSpace(req.TrainingProgress),
		ResponsibleInstructorID: req.ResponsibleInstructorID,
	}
	if err := mergeContractInfoJSON(&s.ContractInfo, req.ContractInfo); err != nil {
		return nil, err
	}
	if err := mergeTrainingProgressJSON(&s.TrainingProgressJSON, req.TrainingProgressJSON); err != nil {
		return nil, err
	}
	if err := u.studentRepo.Create(ctx, s); err != nil {
		return nil, err
	}
	return &SavePersonnelResult{Kind: "student", Student: s}, nil
}

func (u *PersonnelUseCase) updateStudent(ctx context.Context, studentID string, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	s, err := u.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.StudentNo) == "" {
		return nil, ErrInvalidArgument
	}
	s.Name = strings.TrimSpace(req.Name)
	s.StudentNo = strings.TrimSpace(req.StudentNo)
	s.ContractDurationMinutes = req.ContractDurationMinutes
	s.FlownMinutes = req.FlownMinutes
	s.TrainingProgress = strings.TrimSpace(req.TrainingProgress)
	s.ResponsibleInstructorID = req.ResponsibleInstructorID
	if err := mergeContractInfoJSON(&s.ContractInfo, req.ContractInfo); err != nil {
		return nil, err
	}
	if err := mergeTrainingProgressJSON(&s.TrainingProgressJSON, req.TrainingProgressJSON); err != nil {
		return nil, err
	}
	if err := u.studentRepo.Update(ctx, s); err != nil {
		return nil, err
	}
	if s.UserID != nil && strings.TrimSpace(*s.UserID) != "" {
		uu, err := u.userRepo.GetByID(ctx, *s.UserID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else if uu != nil {
			if req.Username != "" {
				if err := u.checkUsernameAvailable(ctx, req.Username, uu.ID); err != nil {
					return nil, err
				}
				uu.Username = strings.TrimSpace(req.Username)
			}
			if req.Phone != "" {
				uu.Phone = strings.TrimSpace(req.Phone)
			}
			if req.Email != "" {
				uu.Email = strings.TrimSpace(req.Email)
			}
			if req.AvatarURL != "" {
				uu.AvatarURL = strings.TrimSpace(req.AvatarURL)
			}
			if req.Password != "" {
				h, err := password.Hash(req.Password)
				if err != nil {
					return nil, err
				}
				uu.PasswordHash = h
			}
			if strings.TrimSpace(req.EmploymentStatus) != "" {
				es, err := employmentStatusFromRequest(req.EmploymentStatus)
				if err != nil {
					return nil, err
				}
				uu.EmploymentStatus = es
			}
			if err := u.userRepo.Update(ctx, uu); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return nil, ErrNotFound
				}
				return nil, err
			}
		}
	}
	return &SavePersonnelResult{Kind: "student", Student: s}, nil
}

func (u *PersonnelUseCase) saveInstructor(ctx context.Context, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	iid := ""
	if req.InstructorID != nil {
		iid = strings.TrimSpace(*req.InstructorID)
	}
	if iid == "" {
		return u.createInstructor(ctx, req)
	}
	return u.updateInstructor(ctx, iid, req)
}

func (u *PersonnelUseCase) createInstructor(ctx context.Context, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, ErrInvalidArgument
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidArgument
	}
	if err := u.checkUsernameAvailable(ctx, req.Username, ""); err != nil {
		return nil, err
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	uu := &domain.User{
		Username:     strings.TrimSpace(req.Username),
		Phone:        strings.TrimSpace(req.Phone),
		Email:        strings.TrimSpace(req.Email),
		Role:         domain.RoleInstructor,
		PasswordHash: hash,
		AvatarURL:    strings.TrimSpace(req.AvatarURL),
	}
	if es, err := employmentStatusFromRequest(req.EmploymentStatus); err != nil {
		return nil, err
	} else if es != "" {
		uu.EmploymentStatus = es
	}
	if err := u.userRepo.Create(ctx, uu); err != nil {
		return nil, err
	}
	ins := &domain.Instructor{
		UserID:       uu.ID,
		Name:         strings.TrimSpace(req.Name),
		DisplayColor: strings.TrimSpace(req.DisplayColor),
	}
	if err := u.instructorRepo.Create(ctx, ins); err != nil {
		return nil, err
	}
	uu.PasswordHash = ""
	return &SavePersonnelResult{Kind: "instructor", Instructor: ins, User: uu}, nil
}

func (u *PersonnelUseCase) updateInstructor(ctx context.Context, instructorID string, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	ins, err := u.instructorRepo.GetByID(ctx, instructorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(req.Name) != "" {
		ins.Name = strings.TrimSpace(req.Name)
	}
	ins.DisplayColor = strings.TrimSpace(req.DisplayColor)
	if err := u.instructorRepo.Update(ctx, ins); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	uu, err := u.userRepo.GetByID(ctx, ins.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &SavePersonnelResult{Kind: "instructor", Instructor: ins}, nil
		}
		return nil, err
	}
	if req.Username != "" {
		if err := u.checkUsernameAvailable(ctx, req.Username, uu.ID); err != nil {
			return nil, err
		}
		uu.Username = strings.TrimSpace(req.Username)
	}
	if req.Phone != "" {
		uu.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Email != "" {
		uu.Email = strings.TrimSpace(req.Email)
	}
	if req.AvatarURL != "" {
		uu.AvatarURL = strings.TrimSpace(req.AvatarURL)
	}
	if req.Password != "" {
		h, err := password.Hash(req.Password)
		if err != nil {
			return nil, err
		}
		uu.PasswordHash = h
	}
	if strings.TrimSpace(req.EmploymentStatus) != "" {
		es, err := employmentStatusFromRequest(req.EmploymentStatus)
		if err != nil {
			return nil, err
		}
		uu.EmploymentStatus = es
	}
	if err := u.userRepo.Update(ctx, uu); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	uu.PasswordHash = ""
	return &SavePersonnelResult{Kind: "instructor", Instructor: ins, User: uu}, nil
}

func (u *PersonnelUseCase) saveStaff(ctx context.Context, req *SavePersonnelRequest) (*SavePersonnelResult, error) {
	if req.UserID == nil || strings.TrimSpace(*req.UserID) == "" {
		return nil, ErrInvalidArgument
	}
	uid := strings.TrimSpace(*req.UserID)
	uu, err := u.userRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if uu.Role != domain.RoleStaff && uu.Role != domain.RoleMaintenance {
		return nil, ErrPersonnelStaffRole
	}
	if req.Username != "" {
		if err := u.checkUsernameAvailable(ctx, req.Username, uu.ID); err != nil {
			return nil, err
		}
		uu.Username = strings.TrimSpace(req.Username)
	}
	if req.Phone != "" {
		uu.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Email != "" {
		uu.Email = strings.TrimSpace(req.Email)
	}
	if req.AvatarURL != "" {
		uu.AvatarURL = strings.TrimSpace(req.AvatarURL)
	}
	if req.Password != "" {
		h, err := password.Hash(req.Password)
		if err != nil {
			return nil, err
		}
		uu.PasswordHash = h
	}
	if strings.TrimSpace(req.EmploymentStatus) != "" {
		es, err := employmentStatusFromRequest(req.EmploymentStatus)
		if err != nil {
			return nil, err
		}
		uu.EmploymentStatus = es
	}
	if err := u.userRepo.Update(ctx, uu); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	uu.PasswordHash = ""
	return &SavePersonnelResult{Kind: "staff", User: uu}, nil
}
