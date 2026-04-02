package studentcontract

import (
	"bytes"
	"encoding/json"
)

// Entitlements 从 students.contract_info JSON 中解析的权益字段（缺省为 0）。
type Entitlements struct {
	LearningPoints int `json:"learning_points"`
	GuestQuota     int `json:"guest_quota"`
	GuestUsed      int `json:"guest_used"`
	GuestRemaining int `json:"guest_remaining"`
}

type guestExperienceWire struct {
	Quota *int `json:"quota"`
	Used  *int `json:"used"`
}

type contractInfoWire struct {
	LearningPoints  *int                 `json:"learning_points"`
	GuestExperience *guestExperienceWire `json:"guest_experience"`
}

// ParseEntitlementsFromContractInfo 解析 contract_info JSON；非法 JSON 或缺字段时返回零值。
// learning_points 支持标量（历史）或 { total, used, remaining }，学员端取 remaining 作为当前可用积分。
func ParseEntitlementsFromContractInfo(raw []byte) Entitlements {
	var out Entitlements
	if len(raw) == 0 {
		return out
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return out
	}
	if lp, ok := root["learning_points"]; ok {
		lp = bytes.TrimSpace(lp)
		if len(lp) > 0 && lp[0] == '{' {
			snap := parseLearningPointsWire(lp)
			out.LearningPoints = snap.Remaining
		} else {
			var n int
			if err := json.Unmarshal(lp, &n); err == nil {
				out.LearningPoints = nonNeg(n)
			}
		}
	}
	var w contractInfoWire
	// 仅用于 guest（learning_points 已单独处理）
	wb, _ := json.Marshal(root)
	if err := json.Unmarshal(wb, &w); err != nil {
		return out
	}
	if w.GuestExperience != nil {
		if w.GuestExperience.Quota != nil {
			out.GuestQuota = nonNeg(*w.GuestExperience.Quota)
		}
		if w.GuestExperience.Used != nil {
			out.GuestUsed = nonNeg(*w.GuestExperience.Used)
		}
	}
	rem := out.GuestQuota - out.GuestUsed
	if rem < 0 {
		rem = 0
	}
	out.GuestRemaining = rem
	return out
}

func nonNeg(n int) int {
	if n < 0 {
		return 0
	}
	return n
}
