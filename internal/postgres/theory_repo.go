package postgres

import (
	"context"
	"database/sql"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TheoryDocumentRepo struct{ db *DB }

func (db *DB) TheoryDocument() repository.TheoryDocumentRepository { return &TheoryDocumentRepo{db: db} }

var _ repository.TheoryDocumentRepository = (*TheoryDocumentRepo)(nil)

func pgStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func pgInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func pgInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func scanTheoryDocument(
	storage sql.NullString,
	objectID sql.NullString,
	ctype sql.NullString,
	size sql.NullInt64,
	dur sql.NullInt32,
	d *domain.TheoryDocument,
) {
	if storage.Valid {
		d.VideoStorage = storage.String
	}
	if objectID.Valid {
		d.VideoObjectID = objectID.String
	}
	if ctype.Valid {
		d.VideoContentType = ctype.String
	}
	if size.Valid {
		d.VideoSizeBytes = size.Int64
	}
	if dur.Valid {
		d.VideoDurationSeconds = int(dur.Int32)
	}
}

func (r *TheoryDocumentRepo) GetByID(ctx context.Context, id string) (*domain.TheoryDocument, error) {
	var d domain.TheoryDocument
	var storage, objectID, ctype sql.NullString
	var size sql.NullInt64
	var dur sql.NullInt32
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, title, COALESCE(slug,''), sort_order, COALESCE(content_ref,''), COALESCE(video_url,''),
		       video_storage, video_object_id, video_content_type, video_size_bytes, video_duration_seconds,
		       COALESCE(category,''), created_at, updated_at
		FROM theory_documents WHERE id = $1::uuid`, id).Scan(
		&d.ID, &d.Title, &d.Slug, &d.SortOrder, &d.ContentRef, &d.VideoURL,
		&storage, &objectID, &ctype, &size, &dur,
		&d.Category, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	scanTheoryDocument(storage, objectID, ctype, size, dur, &d)
	return &d, nil
}

func (r *TheoryDocumentRepo) List(ctx context.Context) ([]*domain.TheoryDocument, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, title, COALESCE(slug,''), sort_order, COALESCE(content_ref,''), COALESCE(video_url,''),
		       video_storage, video_object_id, video_content_type, video_size_bytes, video_duration_seconds,
		       COALESCE(category,''), created_at, updated_at
		FROM theory_documents ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TheoryDocument
	for rows.Next() {
		var d domain.TheoryDocument
		var storage, objectID, ctype sql.NullString
		var size sql.NullInt64
		var dur sql.NullInt32
		if err := rows.Scan(&d.ID, &d.Title, &d.Slug, &d.SortOrder, &d.ContentRef, &d.VideoURL,
			&storage, &objectID, &ctype, &size, &dur,
			&d.Category, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		scanTheoryDocument(storage, objectID, ctype, size, dur, &d)
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *TheoryDocumentRepo) Create(ctx context.Context, d *domain.TheoryDocument) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO theory_documents (title, slug, sort_order, content_ref, video_url, category,
			video_storage, video_object_id, video_content_type, video_size_bytes, video_duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id::text, created_at, updated_at`,
		d.Title, d.Slug, d.SortOrder, d.ContentRef, d.VideoURL, d.Category,
		pgStr(d.VideoStorage), pgStr(d.VideoObjectID), pgStr(d.VideoContentType), pgInt64(d.VideoSizeBytes), pgInt(d.VideoDurationSeconds),
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *TheoryDocumentRepo) Update(ctx context.Context, d *domain.TheoryDocument) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE theory_documents
		SET title = $2, slug = $3, sort_order = $4, content_ref = $5, video_url = $6, category = $7,
		    video_storage = $8, video_object_id = $9, video_content_type = $10, video_size_bytes = $11, video_duration_seconds = $12,
		    updated_at = NOW()
		WHERE id = $1::uuid`,
		d.ID, d.Title, d.Slug, d.SortOrder, d.ContentRef, d.VideoURL, d.Category,
		pgStr(d.VideoStorage), pgStr(d.VideoObjectID), pgStr(d.VideoContentType), pgInt64(d.VideoSizeBytes), pgInt(d.VideoDurationSeconds),
	)
	return err
}

func (r *TheoryDocumentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM theory_documents WHERE id = $1::uuid`, id)
	return err
}
