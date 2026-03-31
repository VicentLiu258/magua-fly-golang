package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingRecordRepo struct{ db *DB }

func (db *DB) TrainingRecord() repository.TrainingRecordRepository { return &TrainingRecordRepo{db: db} }

var _ repository.TrainingRecordRepository = (*TrainingRecordRepo)(nil)

var trainingRecordColumns = `id::text, student_id::text, instructor_id::text, aircraft_id::text, training_date, engine_on_time, engine_off_time,
	flight_duration_minutes, is_local, import_source, COALESCE(raw_import_id::text,''), COALESCE(status::text,'normal'),
	planned_duration_minutes, theory_duration_minutes, sim_duration_minutes, COALESCE(syllabus_outline,''), COALESCE(archive_display_no,''),
	COALESCE(departure,''), COALESCE(destination,''), COALESCE(flight_type,''), COALESCE(task_type,''), landings,
	COALESCE(training_plan_id::text,''),
	created_at, updated_at`

func (r *TrainingRecordRepo) GetByID(ctx context.Context, id string) (*domain.TrainingRecord, error) {
	var rec domain.TrainingRecord
	err := r.db.Pool.QueryRow(ctx, `SELECT `+trainingRecordColumns+` FROM training_records WHERE id = $1::uuid`, id).Scan(
		&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
		&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
		&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
		&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *TrainingRecordRepo) FindByTrainingPlanID(ctx context.Context, trainingPlanID string) (*domain.TrainingRecord, error) {
	var rec domain.TrainingRecord
	err := r.db.Pool.QueryRow(ctx, `SELECT `+trainingRecordColumns+` FROM training_records WHERE training_plan_id = $1::uuid`, trainingPlanID).Scan(
		&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
		&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
		&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
		&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *TrainingRecordRepo) ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingRecord, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT `+trainingRecordColumns+` FROM training_records WHERE student_id = $1::uuid ORDER BY training_date DESC LIMIT $2 OFFSET $3`, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingRecord
	for rows.Next() {
		var rec domain.TrainingRecord
		if err := rows.Scan(
			&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
			&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
			&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
			&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &rec)
	}
	return list, rows.Err()
}

func (r *TrainingRecordRepo) List(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit, offset int) ([]*domain.TrainingRecord, int, error) {
	var conds []string
	var args []interface{}
	argNum := 1
	if filter != nil {
		if filter.StudentID != nil {
			conds = append(conds, fmt.Sprintf("student_id = $%d::uuid", argNum))
			args = append(args, *filter.StudentID)
			argNum++
		}
		if filter.InstructorID != nil {
			conds = append(conds, fmt.Sprintf("instructor_id = $%d::uuid", argNum))
			args = append(args, *filter.InstructorID)
			argNum++
		}
		if filter.DateFrom != nil {
			conds = append(conds, fmt.Sprintf("training_date >= $%d", argNum))
			args = append(args, *filter.DateFrom)
			argNum++
		}
		if filter.DateTo != nil {
			conds = append(conds, fmt.Sprintf("training_date <= $%d", argNum))
			args = append(args, *filter.DateTo)
			argNum++
		}
		if filter.Status != nil {
			conds = append(conds, fmt.Sprintf("status = $%d", argNum))
			args = append(args, string(*filter.Status))
			argNum++
		}
		if filter.ImportSource != nil {
			conds = append(conds, fmt.Sprintf("import_source = $%d", argNum))
			args = append(args, string(*filter.ImportSource))
			argNum++
		}
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	order := " ORDER BY training_date DESC"
	if sort != nil && sort.Field != "" {
		order = " ORDER BY " + sort.Field
		if !sort.Asc {
			order += " DESC"
		} else {
			order += " ASC"
		}
	}
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM training_records`+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitOff := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
	rows, err := r.db.Pool.Query(ctx, `SELECT `+trainingRecordColumns+` FROM training_records`+where+order+limitOff, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.TrainingRecord
	for rows.Next() {
		var rec domain.TrainingRecord
		if err := rows.Scan(
			&rec.ID, &rec.StudentID, &rec.InstructorID, &rec.AircraftID, &rec.TrainingDate, &rec.EngineOnTime, &rec.EngineOffTime,
			&rec.FlightDurationMinutes, &rec.IsLocal, &rec.ImportSource, &rec.RawImportID, &rec.Status,
			&rec.PlannedDurationMinutes, &rec.TheoryDurationMinutes, &rec.SimDurationMinutes, &rec.SyllabusOutline, &rec.ArchiveDisplayNo,
			&rec.Departure, &rec.Destination, &rec.FlightType, &rec.TaskType, &rec.Landings, &rec.TrainingPlanID,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// trainingRecordTableFilterWhere 生成带表别名的 WHERE 子句（参数从 $1 起编）。
func trainingRecordTableFilterWhere(filter *domain.TrainingRecordListFilter, tableAlias string) (where string, args []interface{}, nextArg int) {
	var conds []string
	args = []interface{}{}
	nextArg = 1
	prefix := tableAlias + "."
	if filter != nil {
		if filter.StudentID != nil {
			conds = append(conds, fmt.Sprintf("%sstudent_id = $%d::uuid", prefix, nextArg))
			args = append(args, *filter.StudentID)
			nextArg++
		}
		if filter.InstructorID != nil {
			conds = append(conds, fmt.Sprintf("%sinstructor_id = $%d::uuid", prefix, nextArg))
			args = append(args, *filter.InstructorID)
			nextArg++
		}
		if filter.DateFrom != nil {
			conds = append(conds, fmt.Sprintf("%straining_date >= $%d", prefix, nextArg))
			args = append(args, *filter.DateFrom)
			nextArg++
		}
		if filter.DateTo != nil {
			conds = append(conds, fmt.Sprintf("%straining_date <= $%d", prefix, nextArg))
			args = append(args, *filter.DateTo)
			nextArg++
		}
		if filter.Status != nil {
			conds = append(conds, fmt.Sprintf("%sstatus = $%d", prefix, nextArg))
			args = append(args, string(*filter.Status))
			nextArg++
		}
		if filter.ImportSource != nil {
			conds = append(conds, fmt.Sprintf("%simport_source = $%d", prefix, nextArg))
			args = append(args, string(*filter.ImportSource))
			nextArg++
		}
	}
	if len(conds) == 0 {
		return "", args, nextArg
	}
	return " WHERE " + strings.Join(conds, " AND "), args, nextArg
}

func trainingRecordExportOrderBy(sort *domain.TrainingRecordListSort) string {
	allowed := map[string]string{
		"training_date":           "tr.training_date",
		"created_at":              "tr.created_at",
		"updated_at":              "tr.updated_at",
		"flight_duration_minutes": "tr.flight_duration_minutes",
		"engine_on_time":          "tr.engine_on_time",
		"engine_off_time":         "tr.engine_off_time",
		"id":                      "tr.id",
	}
	field := "tr.training_date"
	asc := false
	if sort != nil && sort.Field != "" {
		if col, ok := allowed[sort.Field]; ok {
			field = col
		}
		asc = sort.Asc
	}
	dir := "DESC"
	if asc {
		dir = "ASC"
	}
	return " ORDER BY " + field + " " + dir + ", tr.id " + dir
}

// CountForExport 与 CSV 导出相同的 JOIN 条件下的行数。
func (r *TrainingRecordRepo) CountForExport(ctx context.Context, filter *domain.TrainingRecordListFilter) (int, error) {
	where, args, _ := trainingRecordTableFilterWhere(filter, "tr")
	q := `SELECT COUNT(*) FROM training_records tr
		INNER JOIN students s ON s.id = tr.student_id
		INNER JOIN instructors i ON i.id = tr.instructor_id
		INNER JOIN aircraft a ON a.id = tr.aircraft_id` + where
	var n int
	err := r.db.Pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

// ListForCSVExport 返回与样稿表头顺序一致的列值（不含表头行）。
func (r *TrainingRecordRepo) ListForCSVExport(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit int, includePriorExperience bool) ([][]string, error) {
	where, args, nextArg := trainingRecordTableFilterWhere(filter, "tr")
	includePh := nextArg
	args = append(args, includePriorExperience)
	nextArg++
	limitPh := nextArg
	args = append(args, limit)

	priorExpr := fmt.Sprintf(`CASE WHEN $%d::boolean THEN ROUND((
			SUM(COALESCE(tr.flight_duration_minutes, 0)) OVER (
				PARTITION BY tr.student_id
				ORDER BY tr.training_date ASC, tr.engine_on_time ASC NULLS LAST, tr.id ASC
			)
		)::numeric / 60.0, 1) ELSE NULL::numeric END`, includePh)

	q := fmt.Sprintf(`
		SELECT
			COALESCE(NULLIF(tr.archive_display_no, ''), '') AS archive_display_no,
			tr.training_date,
			tr.engine_on_time,
			tr.engine_off_time,
			a.code AS aircraft_code,
			COALESCE(tr.departure, '') AS departure,
			COALESCE(tr.destination, '') AS destination,
			tr.is_local,
			COALESCE(tr.task_type, '') AS task_type,
			i.name AS instructor_name,
			s.name AS student_name,
			COALESCE(ta.stage, '') AS archive_stage,
			s.student_no,
			COALESCE(ta.progress_note, '') AS progress_note,
			tr.flight_duration_minutes,
			tr.landings,
			to_char(tp.start_time, 'HH24:MI') AS plan_start,
			to_char(tp.end_time, 'HH24:MI') AS plan_end,
			%s AS prior_cum_hours,
			COALESCE(ta.instructor_review, '') AS instructor_review
		FROM training_records tr
		INNER JOIN students s ON s.id = tr.student_id
		INNER JOIN instructors i ON i.id = tr.instructor_id
		INNER JOIN aircraft a ON a.id = tr.aircraft_id
		LEFT JOIN training_plans tp ON tp.id = tr.training_plan_id
		LEFT JOIN training_archives ta ON ta.training_record_id = tr.id
		%s
		%s
		LIMIT $%d`, priorExpr, where, trainingRecordExportOrderBy(sort), limitPh)

	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out [][]string
	for rows.Next() {
		var archiveNo string
		var trainingDate time.Time
		var engineOn, engineOff sql.NullTime
		var aircraftCode, departure, destination, taskType, instructorName, studentName, archiveStage, studentNo, progressNote string
		var flightMin sql.NullInt32
		var landings sql.NullInt32
		var planStart, planEnd sql.NullString
		var priorCum sql.NullFloat64
		var instructorReview string

		var isLocal bool
		if err := rows.Scan(
			&archiveNo, &trainingDate, &engineOn, &engineOff, &aircraftCode, &departure, &destination,
			&isLocal, &taskType, &instructorName, &studentName, &archiveStage, &studentNo, &progressNote,
			&flightMin, &landings, &planStart, &planEnd, &priorCum, &instructorReview,
		); err != nil {
			return nil, err
		}
		out = append(out, formatTrainingRecordCSVRow(
			archiveNo, trainingDate, engineOn, engineOff, aircraftCode, departure, destination,
			isLocal, taskType, instructorName, studentName, archiveStage, studentNo, progressNote,
			flightMin, landings, planStart, planEnd, priorCum, instructorReview,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

var chineseWeekdays = []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

func formatTrainingRecordCSVRow(
	archiveNo string,
	trainingDate time.Time,
	engineOn, engineOff sql.NullTime,
	aircraftCode, departure, destination string,
	isLocal bool,
	taskType, instructorName, studentName, archiveStage, studentNo, progressNote string,
	flightMin sql.NullInt32,
	landings sql.NullInt32,
	planStart, planEnd sql.NullString,
	priorCum sql.NullFloat64,
	instructorReview string,
) []string {
	lt := trainingDate.In(time.Local)
	flightDateStr := fmt.Sprintf("%d/%d/%d", lt.Year(), int(lt.Month()), lt.Day())
	weekday := chineseWeekdays[int(lt.Weekday())]
	yearMonth := lt.Format("Jan-06")

	localStr := "转场"
	if isLocal {
		localStr = "本场"
	}

	var planS, planE string
	if planStart.Valid {
		planS = planStart.String
	}
	if planEnd.Valid {
		planE = planEnd.String
	}

	var minsStr, landStr, durationHMM, hoursStr, priorStr string
	if flightMin.Valid {
		m := int(flightMin.Int32)
		minsStr = strconv.Itoa(m)
		durationHMM = fmt.Sprintf("%d:%02d", m/60, m%60)
		h := float64(m) / 60.0
		hoursStr = strconv.FormatFloat(math.Round(h*10)/10, 'f', 1, 64)
	}
	if landings.Valid {
		landStr = strconv.Itoa(int(landings.Int32))
	}
	if priorCum.Valid {
		p := priorCum.Float64
		priorStr = strconv.FormatFloat(math.Round(p*10)/10, 'f', 1, 64)
	}

	n := len(domain.TrainingRecordCSVHeaders)
	row := make([]string, n)
	row[0] = archiveNo
	row[1] = flightDateStr
	row[2] = weekday
	row[3] = planS
	row[4] = planE
	row[5] = formatEngineCSVFromNull(engineOn)
	row[6] = formatEngineCSVFromNull(engineOff)
	row[7] = aircraftCode
	row[8] = departure
	row[9] = destination
	row[10] = localStr
	row[11] = taskType
	row[12] = ""
	row[13] = ""
	row[14] = instructorName
	row[15] = studentName
	row[16] = archiveStage
	row[17] = studentNo
	row[18] = progressNote
	row[19] = durationHMM
	row[20] = minsStr
	row[21] = hoursStr
	row[22] = landStr
	row[23] = priorStr
	row[24] = ""
	row[25] = ""
	row[26] = yearMonth
	row[27] = ""
	row[28] = ""
	row[29] = ""
	row[30] = ""
	row[31] = ""
	row[32] = ""
	row[33] = ""
	row[34] = ""
	row[35] = instructorReview
	if n > 36 {
		row[36] = ""
	}
	return row
}

func formatEngineCSVFromNull(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return formatEngineCSVTime(nt.Time)
}

func formatEngineCSVTime(t time.Time) string {
	tt := t.In(time.Local)
	return fmt.Sprintf("%d/%d/%d %d:%02d", tt.Year(), int(tt.Month()), tt.Day(), tt.Hour(), tt.Minute())
}

func (r *TrainingRecordRepo) Create(ctx context.Context, rec *domain.TrainingRecord) error {
	status := rec.Status
	if status == "" {
		status = domain.RecordStatusNormal
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_records (student_id, instructor_id, aircraft_id, training_date, engine_on_time, engine_off_time,
			flight_duration_minutes, is_local, import_source, raw_import_id, status,
			planned_duration_minutes, theory_duration_minutes, sim_duration_minutes, syllabus_outline, archive_display_no,
			departure, destination, flight_type, task_type, landings, training_plan_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10::uuid, $11::training_record_status, $12, $13, $14, NULLIF($15,''), NULLIF($16,''), NULLIF($17,''), NULLIF($18,''), NULLIF($19,''), NULLIF($20,''), $21, $22::uuid)
		RETURNING id::text, created_at, updated_at`,
		rec.StudentID, rec.InstructorID, rec.AircraftID, rec.TrainingDate, rec.EngineOnTime, rec.EngineOffTime,
		rec.FlightDurationMinutes, rec.IsLocal, rec.ImportSource, nullEmpty(rec.RawImportID), status,
		rec.PlannedDurationMinutes, rec.TheoryDurationMinutes, rec.SimDurationMinutes, rec.SyllabusOutline, rec.ArchiveDisplayNo,
		rec.Departure, rec.Destination, rec.FlightType, rec.TaskType, rec.Landings, nullEmpty(rec.TrainingPlanID),
	).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt)
}

func nullEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *TrainingRecordRepo) Update(ctx context.Context, rec *domain.TrainingRecord) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE training_records SET student_id=$2::uuid, instructor_id=$3::uuid, aircraft_id=$4::uuid, training_date=$5,
			engine_on_time=$6, engine_off_time=$7, flight_duration_minutes=$8, is_local=$9, import_source=$10, raw_import_id=$11::uuid,
			status=$12::training_record_status, planned_duration_minutes=$13, theory_duration_minutes=$14, sim_duration_minutes=$15,
			syllabus_outline=NULLIF($16,''), archive_display_no=NULLIF($17,''),
			departure=NULLIF($18,''), destination=NULLIF($19,''), flight_type=NULLIF($20,''), task_type=NULLIF($21,''), landings=$22,
			training_plan_id=$23::uuid, updated_at=NOW()
		WHERE id=$1::uuid`,
		rec.ID, rec.StudentID, rec.InstructorID, rec.AircraftID, rec.TrainingDate, rec.EngineOnTime, rec.EngineOffTime,
		rec.FlightDurationMinutes, rec.IsLocal, rec.ImportSource, nullEmpty(rec.RawImportID),
		rec.Status, rec.PlannedDurationMinutes, rec.TheoryDurationMinutes, rec.SimDurationMinutes, rec.SyllabusOutline, rec.ArchiveDisplayNo,
		rec.Departure, rec.Destination, rec.FlightType, rec.TaskType, rec.Landings, nullEmpty(rec.TrainingPlanID),
	)
	return err
}

func (r *TrainingRecordRepo) UpdateStatus(ctx context.Context, id string, status domain.TrainingRecordStatus) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE training_records SET status = $2::training_record_status, updated_at = NOW() WHERE id = $1::uuid`, id, string(status))
	return err
}
