package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TaskRepo struct{ db *DB }

func (db *DB) Task() repository.TaskRepository { return &TaskRepo{db: db} }

var _ repository.TaskRepository = (*TaskRepo)(nil)

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	var t domain.Task
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, student_id, assigner_id, type, ref_id, title, COALESCE(description,''), due_at, status, created_at, updated_at
		FROM tasks WHERE id = $1`, id).Scan(
		&t.ID, &t.StudentID, &t.AssignerID, &t.Type, &t.RefID, &t.Title, &t.Description, &t.DueAt, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) ListByStudent(ctx context.Context, studentID int64, status *domain.TaskStatus) ([]*domain.Task, error) {
	query := `
		SELECT id, student_id, assigner_id, type, ref_id, title, COALESCE(description,''), due_at, status, created_at, updated_at
		FROM tasks WHERE student_id = $1`
	args := []interface{}{studentID}
	if status != nil {
		query += ` AND status = $2 ORDER BY created_at DESC`
		args = append(args, *status)
	} else {
		query += ` ORDER BY created_at DESC`
	}
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.StudentID, &t.AssignerID, &t.Type, &t.RefID, &t.Title, &t.Description, &t.DueAt, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *TaskRepo) Create(ctx context.Context, t *domain.Task) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO tasks (student_id, assigner_id, type, ref_id, title, description, due_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		t.StudentID, t.AssignerID, t.Type, t.RefID, t.Title, t.Description, t.DueAt, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id int64, status domain.TaskStatus) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE tasks SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}
