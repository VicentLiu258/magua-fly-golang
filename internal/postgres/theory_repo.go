package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TheoryDocumentRepo struct{ db *DB }

func (db *DB) TheoryDocument() repository.TheoryDocumentRepository { return &TheoryDocumentRepo{db: db} }

var _ repository.TheoryDocumentRepository = (*TheoryDocumentRepo)(nil)

func (r *TheoryDocumentRepo) GetByID(ctx context.Context, id string) (*domain.TheoryDocument, error) {
	var d domain.TheoryDocument
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, title, COALESCE(slug,''), sort_order, COALESCE(content_ref,''), COALESCE(video_url,''), COALESCE(category,''), created_at, updated_at
		FROM theory_documents WHERE id = $1::uuid`, id).Scan(
		&d.ID, &d.Title, &d.Slug, &d.SortOrder, &d.ContentRef, &d.VideoURL, &d.Category, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *TheoryDocumentRepo) List(ctx context.Context) ([]*domain.TheoryDocument, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, title, COALESCE(slug,''), sort_order, COALESCE(content_ref,''), COALESCE(video_url,''), COALESCE(category,''), created_at, updated_at
		FROM theory_documents ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TheoryDocument
	for rows.Next() {
		var d domain.TheoryDocument
		if err := rows.Scan(&d.ID, &d.Title, &d.Slug, &d.SortOrder, &d.ContentRef, &d.VideoURL, &d.Category, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *TheoryDocumentRepo) Create(ctx context.Context, d *domain.TheoryDocument) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO theory_documents (title, slug, sort_order, content_ref, video_url, category)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at, updated_at`,
		d.Title, d.Slug, d.SortOrder, d.ContentRef, d.VideoURL, d.Category,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *TheoryDocumentRepo) Update(ctx context.Context, d *domain.TheoryDocument) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE theory_documents
		SET title = $2, slug = $3, sort_order = $4, content_ref = $5, video_url = $6, category = $7, updated_at = NOW()
		WHERE id = $1::uuid`,
		d.ID, d.Title, d.Slug, d.SortOrder, d.ContentRef, d.VideoURL, d.Category,
	)
	return err
}

func (r *TheoryDocumentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM theory_documents WHERE id = $1::uuid`, id)
	return err
}
