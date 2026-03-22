package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SubjectLearningRepo struct{ db *DB }

func (db *DB) SubjectLearning() repository.SubjectLearningRepository {
	return &SubjectLearningRepo{db: db}
}

var _ repository.SubjectLearningRepository = (*SubjectLearningRepo)(nil)

func (r *SubjectLearningRepo) GetByID(ctx context.Context, id string) (*domain.SubjectLearning, error) {
	var s domain.SubjectLearning
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, syllabus_subject_id::text, title,
		       COALESCE(content_ref,''), COALESCE(video_url,''),
		       COALESCE(sort_order, 0), created_at, updated_at
		FROM subject_learning WHERE id = $1::uuid`, id).Scan(
		&s.ID, &s.SyllabusSubjectID, &s.Title, &s.ContentRef, &s.VideoURL, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubjectLearningRepo) List(ctx context.Context) ([]*domain.SubjectLearning, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, syllabus_subject_id::text, title,
		       COALESCE(content_ref,''), COALESCE(video_url,''),
		       COALESCE(sort_order, 0), created_at, updated_at
		FROM subject_learning ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.SubjectLearning
	for rows.Next() {
		var s domain.SubjectLearning
		if err := rows.Scan(&s.ID, &s.SyllabusSubjectID, &s.Title, &s.ContentRef, &s.VideoURL, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}
