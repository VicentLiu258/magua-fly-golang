package usecase

import "magua-fly-golang/internal/domain"

// ContractListItemsForAPI 将合同列表序列化为与 16.4 `POST /api/v1/admin/contracts/list-by-student` 的 data.items 一致的结构。
func ContractListItemsForAPI(contracts []*domain.Contract) []map[string]interface{} {
	var items []map[string]interface{}
	for _, ct := range contracts {
		item := map[string]interface{}{
			"id":               ct.ID,
			"contract_no":      ct.ContractNo,
			"total_minutes":    ct.TotalMinutes,
			"price_total":      ct.PriceTotal,
			"currency":         ct.Currency,
			"status":           ct.Status,
			"payment_mode":     ct.PaymentMode,
			"payment_schedule": ct.PaymentSchedule,
		}
		if ct.SignDate != nil {
			item["sign_date"] = ct.SignDate.Format("2006-01-02")
		} else {
			item["sign_date"] = nil
		}
		if ct.StartDate != nil {
			item["start_date"] = ct.StartDate.Format("2006-01-02")
		} else {
			item["start_date"] = nil
		}
		if ct.EndDate != nil {
			item["end_date"] = ct.EndDate.Format("2006-01-02")
		} else {
			item["end_date"] = nil
		}
		items = append(items, item)
	}
	return items
}
