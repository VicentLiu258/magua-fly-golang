package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TheoryDocumentRepo struct{ db *DB }

func (db *DB) TheoryDocument() repository.TheoryDocumentRepository { return &TheoryDocumentRepo{db: db} }

var _ repository.TheoryDocumentRepository = (*TheoryDocumentRepo)(nil)

func (r *TheoryDocumentRepo) GetByID(ctx context.Context, id int64) (*domain.TheoryDocument, error) {
	var d domain.TheoryDocument
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, title, COALESCE(slug,''), sort_order, COALESCE(content_ref,''), COALESCE(video_url,''), created_at, updated_at
		FROM theory_documents WHERE id = $1`, id).Scan(
		&d.ID, &d.Title, &d.Slug, &d.SortOrder, &d.ContentRef, &d.VideoURL, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *TheoryDocumentRepo) List(ctx context.Context) ([]*domain.TheoryDocument, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, title, COALESCE(slug,''), sort_order, COALESCE(content_ref,''), COALESCE(video_url,''), created_at, updated_at
		FROM theory_documents ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TheoryDocument
	for rows.Next() {
		var d domain.TheoryDocument
		if err := rows.Scan(&d.ID, &d.Title, &d.Slug, &d.SortOrder, &d.ContentRef, &d.VideoURL, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *TheoryDocumentRepo) Create(ctx context.Context, d *domain.TheoryDocument) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO theory_documents (title, slug, sort_order, content_ref, video_url) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`, d.Title, d.Slug, d.SortOrder, d.ContentRef, d.VideoURL,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}
