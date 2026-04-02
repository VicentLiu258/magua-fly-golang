package studentcontract

import (
	"bytes"
	"encoding/json"
	"math"
)

// BookingLimits 预约限制展示（contract_info.booking_limits）。
type BookingLimits struct {
	TwoWeekRule    string `json:"two_week_rule"`
	ThreeMonthRule string `json:"three_month_rule"`
	WeekdayOnly    string `json:"weekday_only"`
	MorningOnly    string `json:"morning_only"`
}

// Extras 其他权益（contract_info.extras）。
type Extras struct {
	TrainingVideoHours int `json:"training_video_hours"`
	CustomShootCount   int `json:"custom_shoot_count"`
}

// LearningPointsBreakdown 积分（新结构：总额 / 已消耗 / 剩余）。
type LearningPointsBreakdown struct {
	Total     int `json:"total"`
	Used      int `json:"used"`
	Remaining int `json:"remaining"`
}

// GuestExperienceBreakdown 亲友体验。
type GuestExperienceBreakdown struct {
	Quota     int `json:"quota"`
	Used      int `json:"used"`
	Remaining int `json:"remaining"`
}

// BenefitsContractSnapshot 从 contract_info 归一化后的权益块（用于后台权益 Tab）。
type BenefitsContractSnapshot struct {
	LearningPoints        LearningPointsBreakdown
	GuestExperience       GuestExperienceBreakdown
	FlightTimeAdjustMinutes int
	BookingLimits         BookingLimits
	Extras                Extras
}

const defaultBookingLimitLabel = "无"

// ParseBenefitsContractSnapshot 解析 contract_info；兼容旧版 learning_points 标量、guest 仅 quota/used。
func ParseBenefitsContractSnapshot(raw []byte) BenefitsContractSnapshot {
	var out BenefitsContractSnapshot
	out.BookingLimits = BookingLimits{
		TwoWeekRule:    defaultBookingLimitLabel,
		ThreeMonthRule: defaultBookingLimitLabel,
		WeekdayOnly:    defaultBookingLimitLabel,
		MorningOnly:    defaultBookingLimitLabel,
	}
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return out
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return out
	}
	if lp, ok := root["learning_points"]; ok {
		out.LearningPoints = parseLearningPointsWire(lp)
	}
	if ge, ok := root["guest_experience"]; ok {
		out.GuestExperience = parseGuestExperienceWire(ge)
	}
	if v, ok := root["flight_time_adjust_minutes"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err == nil {
			out.FlightTimeAdjustMinutes = n
		}
	}
	if bl, ok := root["booking_limits"]; ok {
		var b BookingLimits
		if err := json.Unmarshal(bl, &b); err == nil {
			normalizeBookingLimits(&b)
			out.BookingLimits = b
		}
	}
	if ex, ok := root["extras"]; ok {
		var e Extras
		if err := json.Unmarshal(ex, &e); err == nil {
			out.Extras = e
		}
	}
	return out
}

func normalizeBookingLimits(b *BookingLimits) {
	if b.TwoWeekRule == "" {
		b.TwoWeekRule = defaultBookingLimitLabel
	}
	if b.ThreeMonthRule == "" {
		b.ThreeMonthRule = defaultBookingLimitLabel
	}
	if b.WeekdayOnly == "" {
		b.WeekdayOnly = defaultBookingLimitLabel
	}
	if b.MorningOnly == "" {
		b.MorningOnly = defaultBookingLimitLabel
	}
}

// NormalizeBookingLimitsForSave 写入前将空串规范为默认展示「无」。
func NormalizeBookingLimitsForSave(b *BookingLimits) {
	normalizeBookingLimits(b)
}

func parseLearningPointsWire(raw json.RawMessage) LearningPointsBreakdown {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return LearningPointsBreakdown{}
	}
	if raw[0] == '{' {
		var o struct {
			Total     *int `json:"total"`
			Used      *int `json:"used"`
			Remaining *int `json:"remaining"`
		}
		if err := json.Unmarshal(raw, &o); err != nil {
			return LearningPointsBreakdown{}
		}
		t, u, r := 0, 0, 0
		if o.Total != nil {
			t = nonNegInt(*o.Total)
		}
		if o.Used != nil {
			u = nonNegInt(*o.Used)
		}
		if o.Remaining != nil {
			r = nonNegInt(*o.Remaining)
		} else {
			r = nonNegInt(t - u)
		}
		if o.Total == nil && o.Remaining != nil && o.Used != nil {
			t = u + r
		}
		return LearningPointsBreakdown{Total: t, Used: u, Remaining: r}
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return LearningPointsBreakdown{}
	}
	n = nonNegInt(n)
	// 历史单值：视为当前余额
	return LearningPointsBreakdown{Total: n, Used: 0, Remaining: n}
}

func parseGuestExperienceWire(raw json.RawMessage) GuestExperienceBreakdown {
	var w struct {
		Quota     *int `json:"quota"`
		Used      *int `json:"used"`
		Remaining *int `json:"remaining"`
	}
	if err := json.Unmarshal(raw, &w); err != nil {
		return GuestExperienceBreakdown{}
	}
	q, u := 0, 0
	if w.Quota != nil {
		q = nonNegInt(*w.Quota)
	}
	if w.Used != nil {
		u = nonNegInt(*w.Used)
	}
	rem := q - u
	if w.Remaining != nil {
		rem = nonNegInt(*w.Remaining)
	} else if rem < 0 {
		rem = 0
	} else {
		rem = nonNegInt(rem)
	}
	return GuestExperienceBreakdown{Quota: q, Used: u, Remaining: rem}
}

func nonNegInt(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// PatchLearningPointsObject 将积分块写回 JSON 对象（用于合并进 contract_info）。
func PatchLearningPointsObject(b LearningPointsBreakdown) map[string]int {
	return map[string]int{
		"total":     b.Total,
		"used":      b.Used,
		"remaining": b.Remaining,
	}
}

// PatchGuestExperienceObject 亲友体验写回。
func PatchGuestExperienceObject(b GuestExperienceBreakdown) map[string]int {
	return map[string]int{
		"quota":     b.Quota,
		"used":      b.Used,
		"remaining": b.Remaining,
	}
}

// MergeContractInfoMap 将 contract_info 解析为 map，写入/覆盖若干键后再序列化，保留其余键。
func MergeContractInfoMap(base []byte, patch map[string]interface{}) ([]byte, error) {
	var m map[string]json.RawMessage
	if len(base) > 0 {
		if err := json.Unmarshal(base, &m); err != nil {
			return nil, err
		}
	}
	if m == nil {
		m = make(map[string]json.RawMessage)
	}
	for k, v := range patch {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		m[k] = json.RawMessage(b)
	}
	return json.Marshal(m)
}

// RoundHours1 分钟转小时，保留 1 位小数。
func RoundHours1(minutes int) float64 {
	return math.Round(float64(minutes)/6.0) / 10.0
}
