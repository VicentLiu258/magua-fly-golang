package postgres

import (
	"context"
	"fmt"
	"strings"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type StudentRepo struct{ db *DB }

func (db *DB) Student() repository.StudentRepository { return &StudentRepo{db: db} }

var _ repository.StudentRepository = (*StudentRepo)(nil)

func (r *StudentRepo) GetByID(ctx context.Context, id string) (*domain.Student, error) {
	var s domain.Student
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text,
		       user_id::text,
		       name,
		       student_no,
		       COALESCE(resident_city,''),
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       flight_training_status::text,
		       created_at,
		       updated_at
		FROM students WHERE id = $1::uuid AND deleted_at IS NULL`, id).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ResidentCity,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
		&s.FlightTrainingStatus,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) GetByUserID(ctx context.Context, userID string) (*domain.Student, error) {
	var s domain.Student
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text,
		       user_id::text,
		       name,
		       student_no,
		       COALESCE(resident_city,''),
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       flight_training_status::text,
		       created_at,
		       updated_at
		FROM students WHERE user_id = $1::uuid AND deleted_at IS NULL`, userID).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ResidentCity,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
		&s.FlightTrainingStatus,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) GetByStudentNo(ctx context.Context, studentNo string) (*domain.Student, error) {
	var s domain.Student
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text,
		       user_id::text,
		       name,
		       student_no,
		       COALESCE(resident_city,''),
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       flight_training_status::text,
		       created_at,
		       updated_at
		FROM students WHERE student_no = $1 AND deleted_at IS NULL`, studentNo).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ResidentCity,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
		&s.FlightTrainingStatus,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
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
		SELECT s.id::text,
		       s.user_id::text,
		       s.name,
		       s.student_no,
		       COALESCE(s.resident_city,''),
		       s.contract_info,
		       s.contract_duration_minutes,
		       s.flown_minutes,
		       COALESCE(s.training_progress,''),
		       s.training_progress_json,
		       s.responsible_instructor_id::text,
		       s.flight_training_status::text,
		       s.created_at,
		       s.updated_at
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
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.Name,
			&s.StudentNo,
			&s.ResidentCity,
			&s.ContractInfo,
			&s.ContractDurationMinutes,
			&s.FlownMinutes,
			&s.TrainingProgress,
			&s.TrainingProgressJSON,
			&s.ResponsibleInstructorID,
			&s.FlightTrainingStatus,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
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
			contract_info,
			contract_duration_minutes,
			flown_minutes,
			training_progress,
			training_progress_json,
			responsible_instructor_id,
			flight_training_status
		)
		VALUES ($1::uuid, $2, $3, NULLIF(TRIM($4),''), $5, $6, $7, $8, $9, $10::uuid, $11::student_flight_training_status)
		RETURNING id::text, created_at, updated_at`,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ResidentCity,
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
		flightTrainingStatusOrDefault(s.FlightTrainingStatus),
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *StudentRepo) Update(ctx context.Context, s *domain.Student) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE students
		SET user_id=$2::uuid,
		    name=$3,
		    student_no=$4,
		    resident_city=NULLIF(TRIM($5),''),
		    contract_info=$6,
		    contract_duration_minutes=$7,
		    flown_minutes=$8,
		    training_progress=$9,
		    training_progress_json=$10,
		    responsible_instructor_id=$11::uuid,
		    flight_training_status=$12::student_flight_training_status,
		    updated_at=NOW()
		WHERE id=$1::uuid AND deleted_at IS NULL`,
		s.ID,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ResidentCity,
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
		flightTrainingStatusOrDefault(s.FlightTrainingStatus),
	)
	return err
}

func flightTrainingStatusOrDefault(s domain.FlightTrainingStatus) string {
	if s == "" {
		return string(domain.FlightTrainingInProgress)
	}
	return string(s)
}
