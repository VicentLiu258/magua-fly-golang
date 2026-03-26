package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type CommunityTopicRepo struct{ db *DB }

func (db *DB) CommunityTopic() repository.CommunityTopicRepository {
	return &CommunityTopicRepo{db: db}
}

var _ repository.CommunityTopicRepository = (*CommunityTopicRepo)(nil)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTopic(row rowScanner) (*domain.CommunityTopic, error) {
	var t domain.CommunityTopic
	var desc, cover *string
	err := row.Scan(
		&t.ID, &t.Slug, &t.Title, &desc, &cover, &t.PostCount, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc != nil {
		t.Description = *desc
	}
	if cover != nil {
		t.CoverURL = *cover
	}
	return &t, nil
}

func (r *CommunityTopicRepo) Create(ctx context.Context, t *domain.CommunityTopic) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO community_topics (slug, title, description, cover_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, post_count, created_at, updated_at`,
		t.Slug, t.Title, nullStr(t.Description), nullStr(t.CoverURL),
	).Scan(&t.ID, &t.PostCount, &t.CreatedAt, &t.UpdatedAt)
}

func (r *CommunityTopicRepo) GetByID(ctx context.Context, id string) (*domain.CommunityTopic, error) {
	return scanTopic(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, slug, title, description, cover_url, post_count, created_at, updated_at
		FROM community_topics WHERE id = $1::uuid`, id))
}

func (r *CommunityTopicRepo) GetBySlug(ctx context.Context, slug string) (*domain.CommunityTopic, error) {
	return scanTopic(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, slug, title, description, cover_url, post_count, created_at, updated_at
		FROM community_topics WHERE slug = $1`, slug))
}

func (r *CommunityTopicRepo) ListByIDs(ctx context.Context, ids []string) ([]*domain.CommunityTopic, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, slug, title, description, cover_url, post_count, created_at, updated_at
		FROM community_topics WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityTopic
	for rows.Next() {
		t, err := scanTopic(rows) // pgx.Rows implements Scan like Row
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *CommunityTopicRepo) ListHot(ctx context.Context, limit, offset int) ([]*domain.CommunityTopic, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, slug, title, description, cover_url, post_count, created_at, updated_at
		FROM community_topics
		ORDER BY post_count DESC, updated_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityTopic
	for rows.Next() {
		t, err := scanTopic(rows) // pgx.Rows implements Scan like Row
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
