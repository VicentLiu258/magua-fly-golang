package usecase

import (
	"context"
	"errors"
	"strings"

	"magua-fly-golang/internal/repository"
)

// PersonnelUseCase 管理员软删除、新建/编辑学员 / 教员 / 地面职员账号。
type PersonnelUseCase struct {
	repo           repository.PersonnelRepository
	userRepo       repository.UserRepository
	studentRepo    repository.StudentRepository
	instructorRepo repository.InstructorRepository
}

// NewPersonnelUseCase 创建 PersonnelUseCase（软删除仅依赖 repo；保存需传入 user/student/instructor 仓储）。
func NewPersonnelUseCase(
	repo repository.PersonnelRepository,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	instructorRepo repository.InstructorRepository,
) *PersonnelUseCase {
	return &PersonnelUseCase{
		repo: repo, userRepo: userRepo, studentRepo: studentRepo, instructorRepo: instructorRepo,
	}
}

// SoftDeletePersonnelRequest 对应 POST /api/v1/admin/personnel/delete
type SoftDeletePersonnelRequest struct {
	// Kind: student | instructor | staff（staff 表示 users.id，且角色须为 staff 或 maintenance）
	Kind string `json:"kind" binding:"required"`
	ID   string `json:"id" binding:"required"`
}

// SoftDelete 按 kind 软删除对应档案；学员 / 教员会同时标记关联 users 行，禁止再登录。
func (u *PersonnelUseCase) SoftDelete(ctx context.Context, req SoftDeletePersonnelRequest) error {
	kind := strings.TrimSpace(strings.ToLower(req.Kind))
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return ErrInvalidArgument
	}
	var err error
	switch kind {
	case "student":
		err = u.repo.SoftDeleteStudent(ctx, id)
	case "instructor":
		err = u.repo.SoftDeleteInstructor(ctx, id)
	case "staff":
		err = u.repo.SoftDeleteStaffUser(ctx, id)
	default:
		return ErrInvalidArgument
	}
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, repository.ErrPersonnelStaffRole) {
		return ErrPersonnelStaffRole
	}
	return err
}
