package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// TrainingCSVRowError 导入失败行
type TrainingCSVRowError struct {
	Row int    `json:"row"`
	Msg string `json:"msg"`
}

// TrainingCSVImportResult 批量 CSV 导入结果
type TrainingCSVImportResult struct {
	BatchID      string              `json:"batch_id"`
	InsertedRaw  int                 `json:"inserted_raw"`
	AutoClaimed  int                 `json:"auto_claimed"`
	PendingRaw   int                 `json:"pending_raw"`
	Errors       []TrainingCSVRowError `json:"errors,omitempty"`
	RawImportIDs []string            `json:"raw_import_ids,omitempty"`
}

// ImportTrainingRecordsFromCSV 解析总表 CSV，写入 raw_flight_imports；身份可解析时自动创建 training_record。
func (u *TrainingUseCase) ImportTrainingRecordsFromCSV(ctx context.Context, batchID string, r io.Reader) (*TrainingCSVImportResult, error) {
	if u == nil || u.studentRepo == nil || u.instructorRepo == nil || u.aircraftRepo == nil {
		return nil, errors.New("training csv import: missing repositories")
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) >= 3 && bytes.Equal(b[0:3], utf8BOM) {
		b = b[3:]
	}
	rd := csv.NewReader(bytes.NewReader(b))
	rd.LazyQuotes = true
	rd.TrimLeadingSpace = true
	records, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv: %w", err)
	}
	if len(records) == 0 {
		return &TrainingCSVImportResult{BatchID: batchID}, nil
	}
	if len(records) > domain.MaxTrainingRecordCSVImportRows+1 {
		return nil, fmt.Errorf("csv 行数超过上限 %d（含表头）", domain.MaxTrainingRecordCSVImportRows+1)
	}
	header := records[0]
	if err := validateTrainingCSVHeader(header); err != nil {
		return nil, err
	}
	out := &TrainingCSVImportResult{BatchID: batchID}
	dataRows := records[1:]
	for i, row := range dataRows {
		lineNum := i + 2
		if rowEmpty(row) {
			continue
		}
		if len(row) < len(domain.TrainingRecordCSVHeaders) {
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: "列数不足"})
			continue
		}
		fieldMap := trainingCSVRowToFieldMap(row)
		rawJSON, _ := json.Marshal(rawImportPayload{Source: "csv", FieldMap: fieldMap})

		onT, offT, durMin, isLocal, perr := parseTrainingCSVCoreTimes(row)
		if perr != "" {
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: perr})
			continue
		}

		raw := domain.RawFlightImport{
			BatchID:               batchID,
			EngineOnTime:          onT,
			EngineOffTime:         offT,
			FlightDurationMinutes: durMin,
			IsLocal:               isLocal,
			RawData:               rawJSON,
		}
		if err := u.rawImportRepo.Create(ctx, &raw); err != nil {
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: err.Error()})
			continue
		}
		out.InsertedRaw++
		out.RawImportIDs = append(out.RawImportIDs, raw.ID)

		stu, ins, ac, err := u.resolveTrainingCSVIdentities(ctx, row)
		if err != nil {
			out.PendingRaw++
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: err.Error()})
			continue
		}

		trainingDate, err := parseTrainingCSVDate(strings.TrimSpace(row[1]))
		if err != nil {
			out.PendingRaw++
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: "飞行日期无效: " + err.Error()})
			continue
		}

		flightData := ""
		if len(row) > 36 {
			flightData = strings.TrimSpace(row[36])
		}
		rec := &domain.TrainingRecord{
			StudentID:             stu.ID,
			StudentNo:             stu.StudentNo,
			InstructorID:          ins.ID,
			AircraftID:            ac.ID,
			TrainingDate:          trainingDate,
			EngineOnTime:          onT,
			EngineOffTime:         offT,
			FlightDurationMinutes: durMin,
			IsLocal:               isLocal,
			ImportSource:          domain.ImportImport,
			RawImportID:           raw.ID,
			Status:                domain.RecordStatusNormal,
			Departure:             strings.TrimSpace(row[8]),
			Destination:           strings.TrimSpace(row[9]),
			TaskType:              strings.TrimSpace(row[11]),
			FlightType:            strings.TrimSpace(row[12]),
			BusinessType:          strings.TrimSpace(row[13]),
			VideoTimeliness:       strings.TrimSpace(row[32]),
			VideoOperator:         strings.TrimSpace(row[33]),
			VideoValidity:         strings.TrimSpace(row[34]),
			FlightDataRef:         flightData,
		}
		if rec.Departure == "" {
			rec.Departure = "莲洲"
		}
		if rec.Destination == "" {
			rec.Destination = "莲洲"
		}
		if rs := strings.TrimSpace(row[0]); strings.HasPrefix(strings.ToUpper(rs), "TR") {
			rec.RecordSerialNo = rs
		}
		if lm := strings.TrimSpace(row[22]); lm != "" {
			if n, err := strconv.Atoi(lm); err == nil {
				rec.Landings = &n
			}
		}
		if v := strings.TrimSpace(row[24]); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				rec.EngineStartMeter = &f
			}
		}
		if v := strings.TrimSpace(row[25]); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				rec.EngineStopMeter = &f
			}
		}
		applyRawImportPayloadToRecord(rawJSON, rec)

		if err := u.CreateRecord(ctx, rec); err != nil {
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: "写入训练记录: " + err.Error()})
			out.PendingRaw++
			continue
		}
		if err := u.rawImportRepo.SetTrainingRecordID(ctx, raw.ID, rec.ID); err != nil {
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: "关联 raw: " + err.Error()})
		} else if err := u.rawImportRepo.Delete(ctx, raw.ID); err != nil {
			out.Errors = append(out.Errors, TrainingCSVRowError{Row: lineNum, Msg: "清理导入缓存: " + err.Error()})
		}

		arch := &domain.TrainingArchive{
			TrainingRecordID:       rec.ID,
			StudentID:              rec.StudentID,
			StudentNo:              rec.StudentNo,
			InstructorID:           rec.InstructorID,
			TrainingDate:           rec.TrainingDate,
			EngineOnTime:           rec.EngineOnTime,
			EngineOffTime:          rec.EngineOffTime,
			FlightDurationMinutes:  rec.FlightDurationMinutes,
			Stage:                  strings.TrimSpace(row[16]),
			ProgressNote:           strings.TrimSpace(row[18]),
			InstructorReview:       strings.TrimSpace(row[35]),
		}
		applyRawImportPayloadToArchive(rawJSON, arch)
		_ = u.archiveRepo.Create(ctx, arch)
		out.AutoClaimed++
	}
	return out, nil
}

func validateTrainingCSVHeader(h []string) error {
	if len(h) < len(domain.TrainingRecordCSVHeaders) {
		return fmt.Errorf("表头列数不足，需与模板一致（%d 列）", len(domain.TrainingRecordCSVHeaders))
	}
	for i := range domain.TrainingRecordCSVHeaders {
		if strings.TrimSpace(h[i]) != domain.TrainingRecordCSVHeaders[i] {
			return fmt.Errorf("表头第 %d 列应为 %q，当前为 %q", i+1, domain.TrainingRecordCSVHeaders[i], h[i])
		}
	}
	return nil
}

func rowEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func trainingCSVRowToFieldMap(row []string) map[string]string {
	m := make(map[string]string)
	keys := []string{
		"record_serial_no", "training_date", "weekday", "plan_engine_on", "plan_engine_off",
		"engine_on", "engine_off", "aircraft_code", "departure", "destination", "local_or_cross",
		"task_type", "flight_type", "business_type", "instructor_name", "student_name", "stage",
		"customer_id", "remark", "flight_duration", "minutes", "hours", "landings", "prior_experience",
		"engine_start_meter", "engine_stop_meter", "year_month", "contract", "hourly_value", "session_value",
		"verify", "verify_result", "video_timeliness", "video_operator", "video_validity", "instructor_review", "flight_data",
	}
	for i, k := range keys {
		if i < len(row) {
			m[k] = row[i]
		}
	}
	return m
}

func parseTrainingCSVCoreTimes(row []string) (on, off *time.Time, durMin *int, isLocal bool, errMsg string) {
	day, err := parseTrainingCSVDate(strings.TrimSpace(row[1]))
	if err != nil {
		return nil, nil, nil, true, "飞行日期: " + err.Error()
	}
	on, err = parseTrainingCSVDateTime(day, strings.TrimSpace(row[5]))
	if err != nil || on == nil {
		return nil, nil, nil, true, "开车时刻必填且格式无效"
	}
	off, err = parseTrainingCSVDateTime(day, strings.TrimSpace(row[6]))
	if err != nil || off == nil {
		return nil, nil, nil, true, "关车时刻必填且格式无效"
	}
	localStr := strings.TrimSpace(row[10])
	isLocal = localStr == "本场" || localStr == ""
	if localStr != "" && localStr != "本场" && localStr != "转场" {
		errMsg = "本场/转场须为 本场 或 转场"
		return nil, nil, nil, true, errMsg
	}
	if localStr == "转场" {
		isLocal = false
	}

	var minutes *int
	if s := strings.TrimSpace(row[20]); s != "" {
		if n, e := strconv.Atoi(s); e == nil {
			minutes = &n
		}
	}
	if minutes == nil {
		d := int(off.Sub(*on).Minutes())
		if d < 0 {
			d += 24 * 60
		}
		minutes = &d
	}
	return on, off, minutes, isLocal, ""
}

func parseTrainingCSVDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty")
	}
	layouts := []string{"2006/1/2", "2006/01/02", "2006-1-2", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date %q", s)
}

func parseTrainingCSVDateTime(day time.Time, s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	d := day.In(time.Local)
	// 导出格式 2006/1/2 15:04
	for _, layout := range []string{"2006/1/2 15:04", "2006/01/02 15:04", "2006-1-2 15:04", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}
	// HH:mm
	if len(s) == 5 && s[2] == ':' {
		if h, m, err := clockParts(s); err == nil {
			t := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, time.Local)
			return &t, nil
		}
	}
	// 四位数 1301
	if len(s) == 4 {
		if h, err := strconv.Atoi(s[:2]); err == nil {
			if m, err2 := strconv.Atoi(s[2:]); err2 == nil && h >= 0 && h < 24 && m >= 0 && m < 60 {
				t := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, time.Local)
				return &t, nil
			}
		}
	}
	return nil, fmt.Errorf("invalid time %q", s)
}

func clockParts(s string) (h, m int, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("clock")
	}
	h, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	m, err = strconv.Atoi(parts[1])
	return h, m, err
}

func trainingCSVNamesMatchNormalized(csvName, dbName string) bool {
	return strings.EqualFold(strings.TrimSpace(csvName), strings.TrimSpace(dbName))
}

// trainingCSVCellLooksLikeUUID 用于「客户编号」列：先按学号查，仅当像 UUID 时再按学员 id 查，避免学号形如 UUID 时被误判为 id。
func trainingCSVCellLooksLikeUUID(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) != 36 {
		return false
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	for _, c := range s {
		if c == '-' {
			continue
		}
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return true
}

func (u *TrainingUseCase) resolveTrainingCSVIdentities(ctx context.Context, row []string) (*domain.Student, *domain.Instructor, *domain.Aircraft, error) {
	code := strings.TrimSpace(row[7])
	if code == "" {
		return nil, nil, nil, fmt.Errorf("机号为空")
	}
	ac, err := u.aircraftRepo.GetByCode(ctx, code)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, err
		}
		ac = nil
	}
	if ac == nil {
		list, err := u.aircraftRepo.ListByCodeOrNameNormalized(ctx, code)
		if err != nil {
			return nil, nil, nil, err
		}
		switch len(list) {
		case 0:
			return nil, nil, nil, fmt.Errorf("机号未找到")
		case 1:
			ac = list[0]
		default:
			return nil, nil, nil, fmt.Errorf("飞机匹配到 %d 条", len(list))
		}
	}

	insName := strings.TrimSpace(row[14])
	if insName == "" {
		return nil, nil, nil, fmt.Errorf("飞行教员为空")
	}
	inss, err := u.instructorRepo.ListByExactName(ctx, insName)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(inss) != 1 {
		inss, err = u.instructorRepo.ListByNormalizedName(ctx, insName)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if len(inss) != 1 {
		return nil, nil, nil, fmt.Errorf("教员姓名匹配到 %d 条", len(inss))
	}

	custID := strings.TrimSpace(row[17])
	nameWant := strings.TrimSpace(row[15])

	var stu *domain.Student
	if custID != "" {
		stu, err = u.studentRepo.GetByStudentNo(ctx, custID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, err
		}
		if stu == nil && trainingCSVCellLooksLikeUUID(custID) {
			stu, err = u.studentRepo.GetByID(ctx, custID)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return nil, nil, nil, err
				}
				stu = nil
			}
		}
	}

	if stu != nil {
		if nameWant != "" && !trainingCSVNamesMatchNormalized(nameWant, stu.Name) {
			return nil, nil, nil, fmt.Errorf("客户姓名与库内不一致")
		}
		return stu, inss[0], ac, nil
	}

	if nameWant == "" {
		if custID != "" {
			return nil, nil, nil, fmt.Errorf("学员未找到")
		}
		return nil, nil, nil, fmt.Errorf("客户编号与客户均为空")
	}

	students, err := u.studentRepo.ListByNormalizedName(ctx, nameWant)
	if err != nil {
		return nil, nil, nil, err
	}
	switch len(students) {
	case 0:
		if custID != "" {
			return nil, nil, nil, fmt.Errorf("学员未找到")
		}
		return nil, nil, nil, fmt.Errorf("学员姓名未匹配")
	case 1:
		return students[0], inss[0], ac, nil
	default:
		return nil, nil, nil, fmt.Errorf("学员姓名匹配到 %d 条", len(students))
	}
}
