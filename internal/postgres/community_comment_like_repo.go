package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type CommunityLikeRepo struct{ db *DB }

func (db *DB) CommunityLike() repository.CommunityLikeRepository {
	return &CommunityLikeRepo{db: db}
}

var _ repository.CommunityLikeRepository = (*CommunityLikeRepo)(nil)

func (r *CommunityLikeRepo) Like(ctx context.Context, postID, userID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO community_post_likes (post_id, user_id) VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (post_id, user_id) DO NOTHING`,
		postID, userID)
	return err
}

func (r *CommunityLikeRepo) Unlike(ctx context.Context, postID, userID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		DELETE FROM community_post_likes WHERE post_id = $1::uuid AND user_id = $2::uuid`,
		postID, userID)
	return err
}

type CommunityCommentRepo struct{ db *DB }

func (db *DB) CommunityComment() repository.CommunityCommentRepository {
	return &CommunityCommentRepo{db: db}
}

var _ repository.CommunityCommentRepository = (*CommunityCommentRepo)(nil)

func (r *CommunityCommentRepo) Create(ctx context.Context, c *domain.CommunityComment) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO community_comments (post_id, author_user_id, body, parent_id)
		VALUES ($1::uuid, $2::uuid, $3, $4::uuid)
		RETURNING id::text, created_at`,
		c.PostID, c.AuthorUserID, c.Body, nullUUID(c.ParentID),
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *CommunityCommentRepo) ListByPost(ctx context.Context, postID string, limit, offset int) ([]*domain.CommunityComment, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, post_id::text, author_user_id::text, body, parent_id::text, created_at
		FROM community_comments WHERE post_id = $1::uuid
		ORDER BY created_at ASC, id ASC LIMIT $2 OFFSET $3`,
		postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityComment
	for rows.Next() {
		var c domain.CommunityComment
		var pid *string
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorUserID, &c.Body, &pid, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.ParentID = pid
		out = append(out, &c)
	}
	return out, rows.Err()
}
