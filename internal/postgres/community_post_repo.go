package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"

	"github.com/jackc/pgx/v5"
)

type CommunityPostRepo struct{ db *DB }

func (db *DB) CommunityPost() repository.CommunityPostRepository {
	return &CommunityPostRepo{db: db}
}

var _ repository.CommunityPostRepository = (*CommunityPostRepo)(nil)

func mediaJSON(urls []string) []byte {
	if len(urls) == 0 {
		return []byte("[]")
	}
	b, err := json.Marshal(urls)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func scanMedia(b []byte, dest *[]string) error {
	if len(b) == 0 {
		*dest = nil
		return nil
	}
	return json.Unmarshal(b, dest)
}

func scanPost(row rowScanner) (*domain.CommunityPost, error) {
	var p domain.CommunityPost
	var vis, st string
	var title *string
	var gid *string
	var mediaRaw []byte
	err := row.Scan(
		&p.ID, &p.AuthorUserID, &title, &p.Body, &vis, &gid, &st, &mediaRaw,
		&p.LikeCount, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if title != nil {
		p.Title = *title
	}
	p.Visibility = domain.CommunityPostVisibility(vis)
	p.Status = domain.CommunityPostStatus(st)
	p.GroupID = gid
	if err := scanMedia(mediaRaw, &p.MediaURLs); err != nil {
		return nil, err
	}
	return &p, nil
}

const postSelectCols = `
	p.id::text, p.author_user_id::text, p.title, p.body, p.visibility::text, p.group_id::text, p.status::text,
	p.media_urls, p.like_count, p.comment_count, p.created_at, p.updated_at`

func (r *CommunityPostRepo) Create(ctx context.Context, p *domain.CommunityPost) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO community_posts (author_user_id, title, body, visibility, group_id, status, media_urls)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::jsonb)
		RETURNING id::text, like_count, comment_count, created_at, updated_at`,
		p.AuthorUserID, nullStrPtr(p.Title), p.Body, string(p.Visibility), nullUUID(p.GroupID), string(p.Status), mediaJSON(p.MediaURLs),
	).Scan(&p.ID, &p.LikeCount, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt)
}

func (r *CommunityPostRepo) CreateWithTopics(ctx context.Context, p *domain.CommunityPost, topicIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `
		INSERT INTO community_posts (author_user_id, title, body, visibility, group_id, status, media_urls)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::jsonb)
		RETURNING id::text, like_count, comment_count, created_at, updated_at`,
		p.AuthorUserID, nullStrPtr(p.Title), p.Body, string(p.Visibility), nullUUID(p.GroupID), string(p.Status), mediaJSON(p.MediaURLs),
	).Scan(&p.ID, &p.LikeCount, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}
	for _, tid := range topicIDs {
		if tid == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO community_post_topics (post_id, topic_id) VALUES ($1::uuid, $2::uuid)`,
			p.ID, tid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CommunityPostRepo) Update(ctx context.Context, p *domain.CommunityPost) error {
	ct, err := r.db.Pool.Exec(ctx, `
		UPDATE community_posts SET
			title = $1, body = $2, visibility = $3, group_id = $4, status = $5, media_urls = $6::jsonb, updated_at = NOW()
		WHERE id = $7::uuid AND author_user_id = $8::uuid`,
		nullStrPtr(p.Title), p.Body, string(p.Visibility), nullUUID(p.GroupID), string(p.Status), mediaJSON(p.MediaURLs), p.ID, p.AuthorUserID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CommunityPostRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM community_posts WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CommunityPostRepo) GetByID(ctx context.Context, id string) (*domain.CommunityPost, error) {
	return scanPost(r.db.Pool.QueryRow(ctx, `
		SELECT `+postSelectCols+` FROM community_posts p WHERE p.id = $1::uuid`, id))
}

func (r *CommunityPostRepo) ReplaceTopicLinks(ctx context.Context, postID string, topicIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM community_post_topics WHERE post_id = $1::uuid`, postID); err != nil {
		return err
	}
	for _, tid := range topicIDs {
		if tid == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO community_post_topics (post_id, topic_id) VALUES ($1::uuid, $2::uuid)`,
			postID, tid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CommunityPostRepo) ListTopicIDs(ctx context.Context, postID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT topic_id::text FROM community_post_topics WHERE post_id = $1::uuid`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *CommunityPostRepo) ListFeedGlobal(ctx context.Context, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error) {
	q := `SELECT ` + postSelectCols + ` FROM community_posts p
		WHERE p.status = 'published' AND p.visibility = 'public'`
	args, n := []any{}, 1
	if cursorCreatedAt != nil && cursorID != nil && *cursorID != "" {
		q += fmt.Sprintf(` AND (p.created_at, p.id) < ($%d::timestamptz, $%d::uuid)`, n, n+1)
		args = append(args, *cursorCreatedAt, *cursorID)
		n += 2
	}
	q += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanPostRows(rows)
}

func scanPostRows(rows pgx.Rows) ([]*domain.CommunityPost, error) {
	defer rows.Close()
	var out []*domain.CommunityPost
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *CommunityPostRepo) ListFeedFollowing(ctx context.Context, viewerID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error) {
	q := `SELECT ` + postSelectCols + ` FROM community_posts p
		WHERE p.status = 'published'
		  AND p.author_user_id IN (SELECT followee_id FROM user_follows WHERE follower_id = $1::uuid)
		  AND (
			p.visibility = 'public'
			OR p.visibility = 'followers_only'
			OR (
			  p.visibility = 'group_only'
			  AND EXISTS (
				SELECT 1 FROM community_group_members m
				WHERE m.group_id = p.group_id AND m.user_id = $1::uuid
			  )
			)
		  )`
	args := []any{viewerID}
	n := 2
	if cursorCreatedAt != nil && cursorID != nil && *cursorID != "" {
		q += fmt.Sprintf(` AND (p.created_at, p.id) < ($%d::timestamptz, $%d::uuid)`, n, n+1)
		args = append(args, *cursorCreatedAt, *cursorID)
		n += 2
	}
	q += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanPostRows(rows)
}

func (r *CommunityPostRepo) ListFeedGroup(ctx context.Context, groupID, viewerID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error) {
	q := `SELECT ` + postSelectCols + ` FROM community_posts p
		WHERE p.status = 'published' AND p.group_id = $1::uuid
		  AND EXISTS (
			SELECT 1 FROM community_group_members m
			WHERE m.group_id = $1::uuid AND m.user_id = $2::uuid
		  )`
	args := []any{groupID, viewerID}
	n := 3
	if cursorCreatedAt != nil && cursorID != nil && *cursorID != "" {
		q += fmt.Sprintf(` AND (p.created_at, p.id) < ($%d::timestamptz, $%d::uuid)`, n, n+1)
		args = append(args, *cursorCreatedAt, *cursorID)
		n += 2
	}
	q += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanPostRows(rows)
}

func (r *CommunityPostRepo) ListByTopic(ctx context.Context, topicID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error) {
	q := `SELECT ` + postSelectCols + ` FROM community_posts p
		INNER JOIN community_post_topics pt ON pt.post_id = p.id AND pt.topic_id = $1::uuid
		WHERE p.status = 'published' AND p.visibility = 'public'`
	args := []any{topicID}
	n := 2
	if cursorCreatedAt != nil && cursorID != nil && *cursorID != "" {
		q += fmt.Sprintf(` AND (p.created_at, p.id) < ($%d::timestamptz, $%d::uuid)`, n, n+1)
		args = append(args, *cursorCreatedAt, *cursorID)
		n += 2
	}
	q += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanPostRows(rows)
}

func (r *CommunityPostRepo) ListByAuthor(ctx context.Context, authorID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error) {
	q := `SELECT ` + postSelectCols + ` FROM community_posts p
		WHERE p.author_user_id = $1::uuid AND p.status = 'published'`
	args := []any{authorID}
	n := 2
	if cursorCreatedAt != nil && cursorID != nil && *cursorID != "" {
		q += fmt.Sprintf(` AND (p.created_at, p.id) < ($%d::timestamptz, $%d::uuid)`, n, n+1)
		args = append(args, *cursorCreatedAt, *cursorID)
		n += 2
	}
	q += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return scanPostRows(rows)
}

func (r *CommunityPostRepo) BatchLikedByUser(ctx context.Context, userID string, postIDs []string) (map[string]bool, error) {
	out := make(map[string]bool)
	if len(postIDs) == 0 {
		return out, nil
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT post_id::text FROM community_post_likes
		WHERE user_id = $1::uuid AND post_id = ANY($2::uuid[])`,
		userID, postIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		out[pid] = true
	}
	return out, rows.Err()
}
