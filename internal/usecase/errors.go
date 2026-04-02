package usecase

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidArgument    = errors.New("invalid argument")
	// ErrInvalidContractInfo contract_info JSON 非法。
	ErrInvalidContractInfo = errors.New("invalid contract_info json")
	// ErrInvalidTrainingProgressJSON training_progress_json 非法。
	ErrInvalidTrainingProgressJSON = errors.New("invalid training_progress_json")
	// ErrPersonnelStaffRole 使用 kind=staff 删除时，目标用户角色必须是 staff 或 maintenance。
	ErrPersonnelStaffRole = errors.New("personnel kind staff requires user role staff or maintenance")
	// ErrUsernameConflict 保存人员时用户名已被其他账号占用。
	ErrUsernameConflict = errors.New("username already exists")
)
