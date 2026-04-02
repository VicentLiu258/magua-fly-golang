package domain

import "time"

// 合同付款方式（存储于 contracts.payment_mode）。
const (
	PaymentModeFull        = "full"
	PaymentModeInstallment = "installment"
)

// PaymentInstallment 分期付款计划中的一期（金额 + 应付日期）。
type PaymentInstallment struct {
	Amount  float64 `json:"amount"`
	DueDate string  `json:"due_date"` // YYYY-MM-DD
}

// Contract 表示学员与机构之间的一份训练合同。
type Contract struct {
	ID           string     `json:"id"`
	StudentID    string     `json:"student_id"`
	ContractNo   string     `json:"contract_no,omitempty"`
	SignDate     *time.Time `json:"sign_date,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	TotalMinutes *int       `json:"total_minutes,omitempty"`
	PriceTotal   *float64   `json:"price_total,omitempty"`
	Currency     string     `json:"currency,omitempty"`
	Status       string     `json:"status,omitempty"`
	// PaymentMode 全款 full 或分期 installment。
	PaymentMode string `json:"payment_mode,omitempty"`
	// PaymentSchedule 分期时各期金额与日期；全款时为空切片。
	PaymentSchedule []PaymentInstallment `json:"payment_schedule,omitempty"`
	// MongoFileIDs 保存与本合同关联的 MongoDB 文档 / GridFS 文件 ID 字符串列表。
	MongoFileIDs []string  `json:"mongo_file_ids,omitempty"`
	Remark       string    `json:"remark,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

