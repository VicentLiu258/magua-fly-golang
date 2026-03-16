package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SyllabusSubjectRepo struct{ db *DB }

func (db *DB) SyllabusSubject() repository.SyllabusSubjectRepository { return &SyllabusSubjectRepo{db: db} }

var _ repository.SyllabusSubjectRepository = (*SyllabusSubjectRepo)(nil)

func (r *SyllabusSubjectRepo) GetByID(ctx context.Context, id string) (*domain.SyllabusSubject, error) {
	var s domain.SyllabusSubject
	err := r.db.Pool.QueryRow(ctx, `SELECT id::text, code, name, sort_order, created_at, updated_at FROM syllabus_subjects WHERE id = $1::uuid`, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SyllabusSubjectRepo) List(ctx context.Context) ([]*domain.SyllabusSubject, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id::text, code, name, sort_order, created_at, updated_at FROM syllabus_subjects ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.SyllabusSubject
	for rows.Next() {
		var s domain.SyllabusSubject
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *SyllabusSubjectRepo) Create(ctx context.Context, s *domain.SyllabusSubject) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO syllabus_subjects (code, name, sort_order) VALUES ($1, $2, $3)
		RETURNING id::text, created_at, updated_at`, s.Code, s.Name, s.SortOrder,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SyllabusSubjectRepo) Update(ctx context.Context, s *domain.SyllabusSubject) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE syllabus_subjects
		SET code = $2, name = $3, sort_order = $4, updated_at = NOW()
		WHERE id = $1::uuid`,
		s.ID, s.Code, s.Name, s.SortOrder,
	)
	return err
}

func (r *SyllabusSubjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM syllabus_subjects WHERE id = $1::uuid`, id)
	return err
}
