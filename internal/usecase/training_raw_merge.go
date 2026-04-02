package usecase

import (
	"encoding/json"
	"strconv"
	"strings"

	"magua-fly-golang/internal/domain"
)

// rawImportPayload CSV/扩展导入在 raw_data 中的结构
type rawImportPayload struct {
	Source   string            `json:"source"`
	FieldMap map[string]string `json:"field_map"`
}

func applyRawImportPayloadToRecord(rawData []byte, rec *domain.TrainingRecord) {
	if len(rawData) == 0 || rec == nil {
		return
	}
	var p rawImportPayload
	if err := json.Unmarshal(rawData, &p); err != nil || len(p.FieldMap) == 0 {
		return
	}
	m := p.FieldMap
	setIf := func(key string, dest *string) {
		if v, ok := m[key]; ok {
			v = strings.TrimSpace(v)
			if v != "" {
				*dest = v
			}
		}
	}
	setIf("departure", &rec.Departure)
	setIf("destination", &rec.Destination)
	setIf("flight_type", &rec.FlightType)
	setIf("task_type", &rec.TaskType)
	setIf("business_type", &rec.BusinessType)
	setIf("syllabus_outline", &rec.SyllabusOutline)
	setIf("archive_display_no", &rec.ArchiveDisplayNo)
	setIf("record_serial_no", &rec.RecordSerialNo)
	setIf("video_timeliness", &rec.VideoTimeliness)
	setIf("video_operator", &rec.VideoOperator)
	setIf("video_validity", &rec.VideoValidity)
	setIf("flight_data_ref", &rec.FlightDataRef)
	if v, ok := m["landings"]; ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			rec.Landings = &n
		}
	}
	if v, ok := m["engine_start_meter"]; ok {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			rec.EngineStartMeter = &f
		}
	}
	if v, ok := m["engine_stop_meter"]; ok {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			rec.EngineStopMeter = &f
		}
	}
}

func applyRawImportPayloadToArchive(rawData []byte, a *domain.TrainingArchive) {
	if len(rawData) == 0 || a == nil {
		return
	}
	var p rawImportPayload
	if err := json.Unmarshal(rawData, &p); err != nil || len(p.FieldMap) == 0 {
		return
	}
	m := p.FieldMap
	if v, ok := m["stage"]; ok {
		a.Stage = strings.TrimSpace(v)
	}
	if v, ok := m["progress_note"]; ok && strings.TrimSpace(v) != "" {
		a.ProgressNote = strings.TrimSpace(v)
	} else if v, ok := m["remark"]; ok && strings.TrimSpace(v) != "" {
		a.ProgressNote = strings.TrimSpace(v)
	}
	if v, ok := m["instructor_review"]; ok {
		a.InstructorReview = strings.TrimSpace(v)
	}
}
