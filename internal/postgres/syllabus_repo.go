package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SyllabusSubjectRepo struct{ db *DB }

func (db *DB) SyllabusSubject() repository.SyllabusSubjectRepository { return &SyllabusSubjectRepo{db: db} }

var _ repository.SyllabusSubjectRepository = (*SyllabusSubjectRepo)(nil)

func (r *SyllabusSubjectRepo) GetByID(ctx context.Context, id int64) (*domain.SyllabusSubject, error) {
	var s domain.SyllabusSubject
	err := r.db.Pool.QueryRow(ctx, `SELECT id, code, name, sort_order, created_at, updated_at FROM syllabus_subjects WHERE id = $1`, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SyllabusSubjectRepo) List(ctx context.Context) ([]*domain.SyllabusSubject, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id, code, name, sort_order, created_at, updated_at FROM syllabus_subjects ORDER BY sort_order, id`)
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
		RETURNING id, created_at, updated_at`, s.Code, s.Name, s.SortOrder,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}
