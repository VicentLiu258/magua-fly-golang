package postgres

import (
	"context"
	"fmt"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SyllabusSubjectRepo struct{ db *DB }

func (db *DB) SyllabusSubject() repository.SyllabusSubjectRepository { return &SyllabusSubjectRepo{db: db} }

var _ repository.SyllabusSubjectRepository = (*SyllabusSubjectRepo)(nil)

func (r *SyllabusSubjectRepo) GetByID(ctx context.Context, id string) (*domain.SyllabusSubject, error) {
	var s domain.SyllabusSubject
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, syllabus_outline_id::text, code, name, sort_order, created_at, updated_at
		FROM syllabus_subjects WHERE id = $1::uuid`, id).Scan(
		&s.ID, &s.SyllabusOutlineID, &s.Code, &s.Name, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SyllabusSubjectRepo) List(ctx context.Context, filter *domain.SyllabusSubjectListFilter) ([]*domain.SyllabusSubject, error) {
	var (
		q    string
		args []any
	)
	switch {
	case filter != nil && filter.OutlineID != nil && *filter.OutlineID != "":
		q = `
			SELECT id::text, syllabus_outline_id::text, code, name, sort_order, created_at, updated_at
			FROM syllabus_subjects
			WHERE syllabus_outline_id = $1::uuid
			ORDER BY sort_order, id`
		args = []any{*filter.OutlineID}
	case filter != nil && filter.Category != nil && *filter.Category != "":
		q = `
			SELECT s.id::text, s.syllabus_outline_id::text, s.code, s.name, s.sort_order, s.created_at, s.updated_at
			FROM syllabus_subjects s
			JOIN syllabus_outlines o ON o.id = s.syllabus_outline_id
			WHERE o.category = $1
			ORDER BY s.sort_order, s.id`
		args = []any{*filter.Category}
	default:
		q = `
			SELECT id::text, syllabus_outline_id::text, code, name, sort_order, created_at, updated_at
			FROM syllabus_subjects
			ORDER BY sort_order, id`
	}
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.SyllabusSubject
	for rows.Next() {
		var s domain.SyllabusSubject
		if err := rows.Scan(&s.ID, &s.SyllabusOutlineID, &s.Code, &s.Name, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *SyllabusSubjectRepo) Create(ctx context.Context, s *domain.SyllabusSubject) error {
	if s.SyllabusOutlineID == "" {
		return fmt.Errorf("syllabus_outline_id required")
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO syllabus_subjects (syllabus_outline_id, code, name, sort_order)
		VALUES ($1::uuid, $2, $3, $4)
		RETURNING id::text, created_at, updated_at`,
		s.SyllabusOutlineID, s.Code, s.Name, s.SortOrder,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SyllabusSubjectRepo) Update(ctx context.Context, s *domain.SyllabusSubject) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE syllabus_subjects
		SET syllabus_outline_id = $2::uuid, code = $3, name = $4, sort_order = $5, updated_at = NOW()
		WHERE id = $1::uuid`,
		s.ID, s.SyllabusOutlineID, s.Code, s.Name, s.SortOrder,
	)
	return err
}

func (r *SyllabusSubjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM syllabus_subjects WHERE id = $1::uuid`, id)
	return err
}
