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
	// ErrPhoneConflict 更新资料时手机号已被其他账号占用。
	ErrPhoneConflict = errors.New("phone already exists")
	// 合同付款方式 / 分期计划校验。
	ErrInvalidPaymentMode         = errors.New("invalid payment_mode")
	ErrInvalidPaymentSchedule     = errors.New("invalid payment_schedule")
	ErrPaymentScheduleSumMismatch = errors.New("payment_schedule sum must match price_total")
	ErrPriceTotalRequired         = errors.New("price_total required for installment contract")
)
