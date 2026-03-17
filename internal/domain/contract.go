package domain

import "time"

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
	// MongoFileIDs 保存与本合同关联的 MongoDB 文档 / GridFS 文件 ID 字符串列表。
	MongoFileIDs []string   `json:"mongo_file_ids,omitempty"`
	Remark       string     `json:"remark,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

