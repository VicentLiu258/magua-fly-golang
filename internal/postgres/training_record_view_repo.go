package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingRecordViewRepo struct{ db *DB }

func (db *DB) TrainingRecordView() repository.TrainingRecordViewRepository {
	return &TrainingRecordViewRepo{db: db}
}

var _ repository.TrainingRecordViewRepository = (*TrainingRecordViewRepo)(nil)

func (r *TrainingRecordViewRepo) Create(ctx context.Context, v *domain.TrainingRecordView) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO training_record_views (user_id, name, filter_json, sort_json, group_by)
		VALUES ($1::uuid, $2, $3, $4, NULLIF($5,''))
		RETURNING id::text, created_at, updated_at`,
		v.UserID, v.Name, v.FilterJSON, v.SortJSON, v.GroupBy,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

func (r *TrainingRecordViewRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.TrainingRecordView, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, user_id::text, name, filter_json, sort_json, COALESCE(group_by,''), created_at, updated_at
		FROM training_record_views WHERE user_id = $1::uuid ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingRecordView
	for rows.Next() {
		var v domain.TrainingRecordView
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.FilterJSON, &v.SortJSON, &v.GroupBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

func (r *TrainingRecordViewRepo) GetByID(ctx context.Context, id string) (*domain.TrainingRecordView, error) {
	var v domain.TrainingRecordView
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, name, filter_json, sort_json, COALESCE(group_by,''), created_at, updated_at
		FROM training_record_views WHERE id = $1::uuid`, id).Scan(
		&v.ID, &v.UserID, &v.Name, &v.FilterJSON, &v.SortJSON, &v.GroupBy, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *TrainingRecordViewRepo) Update(ctx context.Context, v *domain.TrainingRecordView) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE training_record_views SET name=$2, filter_json=$3, sort_json=$4, group_by=NULLIF($5,''), updated_at=NOW() WHERE id=$1::uuid`,
		v.ID, v.Name, v.FilterJSON, v.SortJSON, v.GroupBy,
	)
	return err
}

func (r *TrainingRecordViewRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM training_record_views WHERE id = $1::uuid`, id)
	return err
}
