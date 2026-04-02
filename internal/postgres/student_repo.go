package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type StudentRepo struct{ db *DB }

func (db *DB) Student() repository.StudentRepository { return &StudentRepo{db: db} }

var _ repository.StudentRepository = (*StudentRepo)(nil)

func studentSelectColumns(alias string) string {
	p := ""
	if alias != "" {
		p = alias + "."
	}
	return fmt.Sprintf(`
		%[1]sid::text,
		%[1]suser_id::text,
		%[1]sname,
		%[1]sstudent_no,
		COALESCE(%[1]sresident_city,''),
		COALESCE(%[1]scontact_phone,''),
		COALESCE(%[1]sgender,''),
		%[1]sbirth_date,
		COALESCE(%[1]sid_document_no,''),
		COALESCE(%[1]sbirth_place,''),
		COALESCE(%[1]slicense_archive_no,''),
		COALESCE(%[1]seducation_level,''),
		COALESCE(%[1]spermanent_address,''),
		COALESCE(%[1]smailing_address,''),
		COALESCE(%[1]sregion,''),
		COALESCE(%[1]sservice_status_raw,''),
		COALESCE(%[1]spause_reason,''),
		COALESCE(%[1]stransfer_base,''),
		%[1]sregistry_extra,
		%[1]scontract_info,
		%[1]scontract_duration_minutes,
		%[1]sflown_minutes,
		COALESCE(%[1]straining_progress,''),
		%[1]straining_progress_json,
		%[1]sresponsible_instructor_id::text,
		%[1]sflight_training_status::text,
		COALESCE(%[1]straining_course,''),
		%[1]screated_at,
		%[1]supdated_at`, p)
}

func scanStudentFromScanner(scan func(dest ...any) error, s *domain.Student) error {
	var nd sql.NullTime
	err := scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ResidentCity,
		&s.ContactPhone,
		&s.Gender,
		&nd,
		&s.IDDocumentNo,
		&s.BirthPlace,
		&s.LicenseArchiveNo,
		&s.EducationLevel,
		&s.PermanentAddress,
		&s.MailingAddress,
		&s.Region,
		&s.ServiceStatusRaw,
		&s.PauseReason,
		&s.TransferBase,
		&s.RegistryExtra,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
		&s.FlightTrainingStatus,
		&s.TrainingCourse,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if nd.Valid {
		t := nd.Time
		s.BirthDate = &t
	} else {
		s.BirthDate = nil
	}
	return nil
}

func nullTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullBytesJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

func (r *StudentRepo) GetByID(ctx context.Context, id string) (*domain.Student, error) {
	var s domain.Student
	row := r.db.Pool.QueryRow(ctx, `
		SELECT`+studentSelectColumns("")+`
		FROM students WHERE id = $1::uuid AND deleted_at IS NULL`, id)
	if err := scanStudentFromScanner(row.Scan, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) getByIDForUpdateTx(ctx context.Context, tx pgx.Tx, id string) (*domain.Student, error) {
	var s domain.Student
	row := tx.QueryRow(ctx, `
		SELECT`+studentSelectColumns("")+`
		FROM students WHERE id = $1::uuid AND deleted_at IS NULL
		FOR UPDATE`, id)
	if err := scanStudentFromScanner(row.Scan, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) updateTx(ctx context.Context, tx pgx.Tx, s *domain.Student) error {
	_, err := tx.Exec(ctx, `
		UPDATE students
		SET user_id=$2::uuid,
		    name=$3,
		    student_no=$4,
		    resident_city=NULLIF(TRIM($5),''),
		    contact_phone=NULLIF(TRIM($6),''),
		    gender=NULLIF(TRIM($7),''),
		    birth_date=$8,
		    id_document_no=NULLIF(TRIM($9),''),
		    birth_place=NULLIF(TRIM($10),''),
		    license_archive_no=NULLIF(TRIM($11),''),
		    education_level=NULLIF(TRIM($12),''),
		    permanent_address=NULLIF(TRIM($13),''),
		    mailing_address=NULLIF(TRIM($14),''),
		    region=NULLIF(TRIM($15),''),
		    service_status_raw=NULLIF(TRIM($16),''),
		    pause_reason=NULLIF(TRIM($17),''),
		    transfer_base=NULLIF(TRIM($18),''),
		    registry_extra=$19,
		    contract_info=$20,
		    contract_duration_minutes=$21,
		    flown_minutes=$22,
		    training_progress=$23,
		    training_progress_json=$24,
		    responsible_instructor_id=$25::uuid,
		    flight_training_status=$26::student_flight_training_status,
		    training_course=$27,
		    updated_at=NOW()
		WHERE id=$1::uuid AND deleted_at IS NULL`,
		s.ID,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ResidentCity,
		s.ContactPhone,
		s.Gender,
		nullTimePtr(s.BirthDate),
		s.IDDocumentNo,
		s.BirthPlace,
		s.LicenseArchiveNo,
		s.EducationLevel,
		s.PermanentAddress,
		s.MailingAddress,
		s.Region,
		s.ServiceStatusRaw,
		s.PauseReason,
		s.TransferBase,
		nullBytesJSON(s.RegistryExtra),
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
		flightTrainingStatusOrDefault(s.FlightTrainingStatus),
		nullTrainingCourse(s.TrainingCourse),
	)
	return err
}

func (r *StudentRepo) GetByUserID(ctx context.Context, userID string) (*domain.Student, error) {
	var s domain.Student
	row := r.db.Pool.QueryRow(ctx, `
		SELECT`+studentSelectColumns("")+`
		FROM students WHERE user_id = $1::uuid AND deleted_at IS NULL`, userID)
	if err := scanStudentFromScanner(row.Scan, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) GetByStudentNo(ctx context.Context, studentNo string) (*domain.Student, error) {
	var s domain.Student
	row := r.db.Pool.QueryRow(ctx, `
		SELECT`+studentSelectColumns("")+`
		FROM students WHERE student_no = $1 AND deleted_at IS NULL`, studentNo)
	if err := scanStudentFromScanner(row.Scan, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) ListByNormalizedName(ctx context.Context, name string) ([]*domain.Student, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT`+studentSelectColumns("s")+`
		FROM students s
		WHERE s.deleted_at IS NULL
		  AND lower(trim(s.name)) = lower(trim($1))`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Student
	for rows.Next() {
		var s domain.Student
		if err := scanStudentFromScanner(rows.Scan, &s); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *StudentRepo) List(ctx context.Context, p repository.StudentListParams) ([]*domain.Student, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	q := `
		SELECT` + studentSelectColumns("s") + `
		FROM students s`
	var joins string
	if p.OnlyActiveEmployment {
		joins = `
		LEFT JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL`
	}
	where := []string{"s.deleted_at IS NULL"}
	if p.OnlyActiveEmployment {
		where = append(where, "(s.user_id IS NULL OR u.employment_status = 'active')")
	}
	args := []interface{}{}
	argNum := 1
	if kw := strings.TrimSpace(p.Keyword); kw != "" {
		where = append(where, fmt.Sprintf("(s.name ILIKE $%d OR s.student_no ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+kw+"%")
		argNum++
	}
	if rid := strings.TrimSpace(p.ResponsibleInstructorID); rid != "" {
		where = append(where, fmt.Sprintf("s.responsible_instructor_id = $%d::uuid", argNum))
		args = append(args, rid)
		argNum++
	}
	if fts := strings.TrimSpace(p.FlightTrainingStatus); fts != "" {
		where = append(where, fmt.Sprintf("s.flight_training_status = $%d::student_flight_training_status", argNum))
		args = append(args, fts)
		argNum++
	}
	if tc := strings.TrimSpace(p.TrainingCourse); tc != "" {
		where = append(where, fmt.Sprintf("s.training_course = $%d", argNum))
		args = append(args, strings.ToUpper(tc))
		argNum++
	}
	q += joins
	q += " WHERE " + strings.Join(where, " AND ")
	q += fmt.Sprintf(" ORDER BY s.id LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Student
	for rows.Next() {
		var s domain.Student
		if err := scanStudentFromScanner(rows.Scan, &s); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *StudentRepo) Create(ctx context.Context, s *domain.Student) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO students (
			user_id,
			name,
			student_no,
			resident_city,
			contact_phone,
			gender,
			birth_date,
			id_document_no,
			birth_place,
			license_archive_no,
			education_level,
			permanent_address,
			mailing_address,
			region,
			service_status_raw,
			pause_reason,
			transfer_base,
			registry_extra,
			contract_info,
			contract_duration_minutes,
			flown_minutes,
			training_progress,
			training_progress_json,
			responsible_instructor_id,
			flight_training_status,
			training_course
		)
		VALUES (
			$1::uuid, $2, $3, NULLIF(TRIM($4),''),
			NULLIF(TRIM($5),''), NULLIF(TRIM($6),''), $7,
			NULLIF(TRIM($8),''), NULLIF(TRIM($9),''), NULLIF(TRIM($10),''), NULLIF(TRIM($11),''),
			NULLIF(TRIM($12),''), NULLIF(TRIM($13),''), NULLIF(TRIM($14),''),
			NULLIF(TRIM($15),''), NULLIF(TRIM($16),''), NULLIF(TRIM($17),''), $18,
			$19, $20, $21, $22, $23, $24::uuid, $25::student_flight_training_status,
			NULLIF(TRIM($26),'')
		)
		RETURNING id::text, created_at, updated_at`,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ResidentCity,
		s.ContactPhone,
		s.Gender,
		nullTimePtr(s.BirthDate),
		s.IDDocumentNo,
		s.BirthPlace,
		s.LicenseArchiveNo,
		s.EducationLevel,
		s.PermanentAddress,
		s.MailingAddress,
		s.Region,
		s.ServiceStatusRaw,
		s.PauseReason,
		s.TransferBase,
		nullBytesJSON(s.RegistryExtra),
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
		flightTrainingStatusOrDefault(s.FlightTrainingStatus),
		nullTrainingCourse(s.TrainingCourse),
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *StudentRepo) Update(ctx context.Context, s *domain.Student) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE students
		SET user_id=$2::uuid,
		    name=$3,
		    student_no=$4,
		    resident_city=NULLIF(TRIM($5),''),
		    contact_phone=NULLIF(TRIM($6),''),
		    gender=NULLIF(TRIM($7),''),
		    birth_date=$8,
		    id_document_no=NULLIF(TRIM($9),''),
		    birth_place=NULLIF(TRIM($10),''),
		    license_archive_no=NULLIF(TRIM($11),''),
		    education_level=NULLIF(TRIM($12),''),
		    permanent_address=NULLIF(TRIM($13),''),
		    mailing_address=NULLIF(TRIM($14),''),
		    region=NULLIF(TRIM($15),''),
		    service_status_raw=NULLIF(TRIM($16),''),
		    pause_reason=NULLIF(TRIM($17),''),
		    transfer_base=NULLIF(TRIM($18),''),
		    registry_extra=$19,
		    contract_info=$20,
		    contract_duration_minutes=$21,
		    flown_minutes=$22,
		    training_progress=$23,
		    training_progress_json=$24,
		    responsible_instructor_id=$25::uuid,
		    flight_training_status=$26::student_flight_training_status,
		    training_course=$27,
		    updated_at=NOW()
		WHERE id=$1::uuid AND deleted_at IS NULL`,
		s.ID,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ResidentCity,
		s.ContactPhone,
		s.Gender,
		nullTimePtr(s.BirthDate),
		s.IDDocumentNo,
		s.BirthPlace,
		s.LicenseArchiveNo,
		s.EducationLevel,
		s.PermanentAddress,
		s.MailingAddress,
		s.Region,
		s.ServiceStatusRaw,
		s.PauseReason,
		s.TransferBase,
		nullBytesJSON(s.RegistryExtra),
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
		flightTrainingStatusOrDefault(s.FlightTrainingStatus),
		nullTrainingCourse(s.TrainingCourse),
	)
	return err
}

func nullTrainingCourse(s string) interface{} {
	v := strings.TrimSpace(strings.ToUpper(s))
	if v == "" {
		return nil
	}
	return v
}

func flightTrainingStatusOrDefault(s domain.FlightTrainingStatus) string {
	if s == "" {
		return string(domain.FlightTrainingInProgress)
	}
	return string(s)
}
