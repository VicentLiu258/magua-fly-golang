package usecase

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidArgument    = errors.New("invalid argument")
	// ErrPersonnelStaffRole 使用 kind=staff 删除时，目标用户角色必须是 staff 或 maintenance。
	ErrPersonnelStaffRole = errors.New("personnel kind staff requires user role staff or maintenance")
)
