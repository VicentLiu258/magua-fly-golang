package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"

	"github.com/jackc/pgx/v5"
)

var utf8BOMStudent = []byte{0xEF, 0xBB, 0xBF}

// StudentRegistryImportOptions 学员总表 CSV 导入选项
type StudentRegistryImportOptions struct {
	// UpdateExisting 为 true 时，若 student_no 已存在则更新该行（否则跳过并记入错误）
	UpdateExisting bool
}

// StudentRegistryRowError 单行失败说明
type StudentRegistryRowError struct {
	Row int    `json:"row"`
	Msg string `json:"msg"`
}

// StudentRegistryImportResult 批量导入结果
type StudentRegistryImportResult struct {
	Created int                       `json:"created"`
	Updated int                       `json:"updated"`
	Skipped int                       `json:"skipped"`
	Errors  []StudentRegistryRowError `json:"errors,omitempty"`
}

// StudentImportUseCase 学员总表 CSV 导入
type StudentImportUseCase struct {
	studentRepo    repository.StudentRepository
	instructorRepo repository.InstructorRepository
}

// NewStudentImportUseCase 构造
func NewStudentImportUseCase(
	studentRepo repository.StudentRepository,
	instructorRepo repository.InstructorRepository,
) *StudentImportUseCase {
	return &StudentImportUseCase{
		studentRepo:    studentRepo,
		instructorRepo: instructorRepo,
	}
}

type registryExtraJSON struct {
	TwoWeeks              string `json:"two_weeks,omitempty"`
	ThreeMonthEnd         string `json:"three_month_end,omitempty"`
	WeekdayFlight         string `json:"weekday_flight,omitempty"`
	MorningFlight         string `json:"morning_flight,omitempty"`
	TrialCount            string `json:"trial_count,omitempty"`
	TrialRemaining        string `json:"trial_remaining,omitempty"`
	VideoRecord           string `json:"video_record,omitempty"`
	CustomShoot           string `json:"custom_shoot,omitempty"`
	CustomShootRemaining  string `json:"custom_shoot_remaining,omitempty"`
	TotalContractHoursRaw string `json:"total_contract_hours_raw,omitempty"`
}

// ImportRegistryCSV 解析学员总表 CSV（首行表头须与 domain.StudentRegistryCSVHeaders 一致）
func (u *StudentImportUseCase) ImportRegistryCSV(ctx context.Context, r io.Reader, opts StudentRegistryImportOptions) (*StudentRegistryImportResult, error) {
	if u == nil || u.studentRepo == nil || u.instructorRepo == nil {
		return nil, errors.New("student import: missing repositories")
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) >= 3 && bytes.Equal(b[0:3], utf8BOMStudent) {
		b = b[3:]
	}
	rd := csv.NewReader(bytes.NewReader(b))
	rd.LazyQuotes = true
	rd.TrimLeadingSpace = true
	records, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv: %w", err)
	}
	out := &StudentRegistryImportResult{}
	if len(records) == 0 {
		return out, nil
	}
	if len(records) > domain.MaxStudentRegistryCSVImportRows+1 {
		return nil, fmt.Errorf("csv 行数超过上限 %d（含表头）", domain.MaxStudentRegistryCSVImportRows+1)
	}
	header := records[0]
	if err := validateStudentRegistryHeader(header); err != nil {
		return nil, err
	}
	dataRows := records[1:]
	for i, row := range dataRows {
		lineNum := i + 2
		if studentRegistryRowEmpty(row) {
			continue
		}
		if len(row) < len(domain.StudentRegistryCSVHeaders) {
			out.Errors = append(out.Errors, StudentRegistryRowError{Row: lineNum, Msg: "列数不足"})
			continue
		}
		stu, err := u.buildStudentFromRegistryRow(ctx, row)
		if err != nil {
			out.Errors = append(out.Errors, StudentRegistryRowError{Row: lineNum, Msg: err.Error()})
			continue
		}
		if strings.TrimSpace(stu.StudentNo) == "" || strings.TrimSpace(stu.Name) == "" {
			out.Skipped++
			continue
		}
		existing, err := u.studentRepo.GetByStudentNo(ctx, strings.TrimSpace(stu.StudentNo))
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				out.Errors = append(out.Errors, StudentRegistryRowError{Row: lineNum, Msg: err.Error()})
				continue
			}
			if err := u.studentRepo.Create(ctx, stu); err != nil {
				out.Errors = append(out.Errors, StudentRegistryRowError{Row: lineNum, Msg: err.Error()})
				continue
			}
			out.Created++
			continue
		}
		if !opts.UpdateExisting {
			out.Errors = append(out.Errors, StudentRegistryRowError{Row: lineNum, Msg: "学号已存在（跳过或开启 update_existing）"})
			continue
		}
		stu.ID = existing.ID
		stu.UserID = existing.UserID
		stu.FlownMinutes = existing.FlownMinutes
		if len(stu.ContractInfo) == 0 {
			stu.ContractInfo = existing.ContractInfo
		}
		if err := u.studentRepo.Update(ctx, stu); err != nil {
			out.Errors = append(out.Errors, StudentRegistryRowError{Row: lineNum, Msg: err.Error()})
			continue
		}
		out.Updated++
	}
	return out, nil
}

func validateStudentRegistryHeader(header []string) error {
	if len(header) < len(domain.StudentRegistryCSVHeaders) {
		return fmt.Errorf("表头列数不足，期望至少 %d 列", len(domain.StudentRegistryCSVHeaders))
	}
	for i, want := range domain.StudentRegistryCSVHeaders {
		if strings.TrimSpace(header[i]) != want {
			return fmt.Errorf("表头第 %d 列应为 %q，实际为 %q", i+1, want, header[i])
		}
	}
	return nil
}

func studentRegistryRowEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func (u *StudentImportUseCase) buildStudentFromRegistryRow(ctx context.Context, row []string) (*domain.Student, error) {
	get := func(i int) string {
		if i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	studentNo := strings.TrimSpace(get(0))
	name := strings.TrimSpace(get(1))
	instructorRaw := get(2)
	region := get(3)
	serviceStatus := get(4)
	pauseReason := get(5)
	transferBase := get(6)
	extra := registryExtraJSON{
		TwoWeeks:              get(7),
		ThreeMonthEnd:         get(8),
		WeekdayFlight:         get(9),
		MorningFlight:         get(10),
		TrialCount:            get(11),
		TrialRemaining:        get(12),
		VideoRecord:           get(13),
		CustomShoot:           get(14),
		CustomShootRemaining:  get(15),
		TotalContractHoursRaw: get(16),
	}
	idDoc := normalizeIDDocument(get(17))
	birthStr := get(18)
	birthPlace := get(19)
	gender := get(20)
	licenseNo := get(21)
	phone := normalizePhone(get(22))
	edu := get(23)
	permAddr := get(24)
	mailAddr := get(25)

	extraJSON, _ := json.Marshal(extra)

	var birth *time.Time
	if t, ok := parseChineseBirthDate(birthStr); ok {
		birth = &t
	}

	insID, insErr := u.resolveResponsibleInstructorID(ctx, instructorRaw)
	if insErr != "" {
		return nil, fmt.Errorf("教员: %s", insErr)
	}

	st := &domain.Student{
		Name:             name,
		StudentNo:        studentNo,
		Region:           region,
		ServiceStatusRaw: serviceStatus,
		PauseReason:      pauseReason,
		TransferBase:     transferBase,
		RegistryExtra:    extraJSON,
		IDDocumentNo:     idDoc,
		BirthPlace:       birthPlace,
		Gender:           gender,
		LicenseArchiveNo: licenseNo,
		ContactPhone:     phone,
		EducationLevel:   edu,
		PermanentAddress: permAddr,
		MailingAddress:   mailAddr,
		BirthDate:        birth,
		ResidentCity:     firstNonEmpty(birthPlace, mailAddr),
		FlightTrainingStatus: mapServiceStatusToFlightTraining(
			serviceStatus,
			pauseReason,
		),
		ResponsibleInstructorID: insID,
	}
	if minutes := parseTotalContractHoursToMinutes(extra.TotalContractHoursRaw); minutes != nil {
		st.ContractDurationMinutes = minutes
	}
	return st, nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return strings.TrimSpace(a)
	}
	return strings.TrimSpace(b)
}

func mapServiceStatusToFlightTraining(serviceStatus, pauseReason string) domain.FlightTrainingStatus {
	s := strings.TrimSpace(serviceStatus)
	switch {
	case strings.Contains(s, "进行中"):
		return domain.FlightTrainingInProgress
	case strings.Contains(s, "已完成"):
		return domain.FlightTrainingCompleted
	case strings.Contains(s, "暂时中断") || strings.Contains(s, "中断"):
		return domain.FlightTrainingPaused
	case strings.Contains(s, "未开始"):
		return domain.FlightTrainingPaused
	default:
		if strings.Contains(strings.TrimSpace(pauseReason), "结业") {
			return domain.FlightTrainingCompleted
		}
		return domain.FlightTrainingInProgress
	}
}

func (u *StudentImportUseCase) resolveResponsibleInstructorID(ctx context.Context, raw string) (*string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "待分配" {
		return nil, ""
	}
	parts := splitInstructorNames(raw)
	for _, name := range parts {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		list, err := u.instructorRepo.ListByNormalizedName(ctx, name)
		if err != nil {
			return nil, err.Error()
		}
		switch len(list) {
		case 0:
			continue
		case 1:
			id := list[0].ID
			return &id, ""
		default:
			return nil, fmt.Sprintf("%q 匹配到多名教员", name)
		}
	}
	return nil, fmt.Sprintf("未匹配到教员: %s", raw)
}

func splitInstructorNames(s string) []string {
	s = strings.NewReplacer("，", ",", "、", ",").Replace(s)
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

var chineseDateRe = regexp.MustCompile(`(\d{4})\s*年\s*(\d{1,2})\s*月\s*(\d{1,2})\s*日?`)

func parseChineseBirthDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "出生")
	s = strings.TrimSpace(s)
	if s == "" || s == "/" {
		return time.Time{}, false
	}
	m := chineseDateRe.FindStringSubmatch(s)
	if len(m) != 4 {
		return time.Time{}, false
	}
	y, _ := strconv.Atoi(m[1])
	mo, _ := strconv.Atoi(m[2])
	d, _ := strconv.Atoi(m[3])
	if y < 1900 || y > 2100 {
		return time.Time{}, false
	}
	return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC), true
}

func parseTotalContractHoursToMinutes(raw string) *int {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "/" {
		return nil
	}
	parts := strings.Split(raw, "+")
	sum := 0.0
	ok := false
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil
		}
		if v < 0 || v > 1e6 {
			return nil
		}
		sum += v
		ok = true
	}
	if !ok {
		return nil
	}
	m := int(sum * 60)
	if m <= 0 {
		return nil
	}
	return &m
}

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		if r == '+' || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeIDDocument(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "/" {
		return ""
	}
	if strings.ContainsAny(s, "Ee") && strings.Contains(s, ".") {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 1e10 {
			return strconv.FormatInt(int64(v), 10)
		}
	}
	return s
}
