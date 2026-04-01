package postgres

import (
	"context"

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
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       created_at,
		       updated_at
		FROM students WHERE id = $1::uuid AND deleted_at IS NULL`, id).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
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
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       created_at,
		       updated_at
		FROM students WHERE user_id = $1::uuid AND deleted_at IS NULL`, userID).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
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
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       created_at,
		       updated_at
		FROM students WHERE student_no = $1 AND deleted_at IS NULL`, studentNo).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.StudentNo,
		&s.ContractInfo,
		&s.ContractDurationMinutes,
		&s.FlownMinutes,
		&s.TrainingProgress,
		&s.TrainingProgressJSON,
		&s.ResponsibleInstructorID,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepo) List(ctx context.Context, limit, offset int) ([]*domain.Student, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text,
		       user_id::text,
		       name,
		       student_no,
		       contract_info,
		       contract_duration_minutes,
		       flown_minutes,
		       COALESCE(training_progress,''),
		       training_progress_json,
		       responsible_instructor_id::text,
		       created_at,
		       updated_at
		FROM students WHERE deleted_at IS NULL ORDER BY id LIMIT $1 OFFSET $2`, limit, offset)
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
			&s.ContractInfo,
			&s.ContractDurationMinutes,
			&s.FlownMinutes,
			&s.TrainingProgress,
			&s.TrainingProgressJSON,
			&s.ResponsibleInstructorID,
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
			contract_info,
			contract_duration_minutes,
			flown_minutes,
			training_progress,
			training_progress_json,
			responsible_instructor_id
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9::uuid)
		RETURNING id::text, created_at, updated_at`,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *StudentRepo) Update(ctx context.Context, s *domain.Student) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE students
		SET user_id=$2::uuid,
		    name=$3,
		    student_no=$4,
		    contract_info=$5,
		    contract_duration_minutes=$6,
		    flown_minutes=$7,
		    training_progress=$8,
		    training_progress_json=$9,
		    responsible_instructor_id=$10::uuid,
		    updated_at=NOW()
		WHERE id=$1::uuid AND deleted_at IS NULL`,
		s.ID,
		nullUUID(s.UserID),
		s.Name,
		s.StudentNo,
		s.ContractInfo,
		s.ContractDurationMinutes,
		s.FlownMinutes,
		s.TrainingProgress,
		s.TrainingProgressJSON,
		nullUUID(s.ResponsibleInstructorID),
	)
	return err
}

