package postgres

import (
	"context"
	"errors"

	"magua-fly-golang/internal/repository"

	"github.com/jackc/pgx/v5"
)

// CommunityFollowRepo implements CommunityFollowRepository.
type CommunityFollowRepo struct{ db *DB }

func (db *DB) CommunityFollow() repository.CommunityFollowRepository {
	return &CommunityFollowRepo{db: db}
}

var _ repository.CommunityFollowRepository = (*CommunityFollowRepo)(nil)

func (r *CommunityFollowRepo) Follow(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return errors.New("cannot follow self")
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_follows (follower_id, followee_id) VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (follower_id, followee_id) DO NOTHING`,
		followerID, followeeID)
	return err
}

func (r *CommunityFollowRepo) Unfollow(ctx context.Context, followerID, followeeID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		DELETE FROM user_follows WHERE follower_id = $1::uuid AND followee_id = $2::uuid`,
		followerID, followeeID)
	return err
}

func (r *CommunityFollowRepo) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var n int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT 1 FROM user_follows WHERE follower_id = $1::uuid AND followee_id = $2::uuid LIMIT 1`,
		followerID, followeeID).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CommunityFollowRepo) ListFolloweeIDs(ctx context.Context, followerID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT followee_id::text FROM user_follows WHERE follower_id = $1::uuid`, followerID)
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

func (r *CommunityFollowRepo) ListFollowers(ctx context.Context, followeeID string, limit, offset int) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT follower_id::text FROM user_follows WHERE followee_id = $1::uuid
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		followeeID, limit, offset)
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

func (r *CommunityFollowRepo) ListFollowing(ctx context.Context, followerID string, limit, offset int) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT followee_id::text FROM user_follows WHERE follower_id = $1::uuid
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		followerID, limit, offset)
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

func (r *CommunityFollowRepo) CountFollowers(ctx context.Context, followeeID string) (int, error) {
	var n int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM user_follows WHERE followee_id = $1::uuid`, followeeID).Scan(&n)
	return n, err
}

func (r *CommunityFollowRepo) CountFollowing(ctx context.Context, followerID string) (int, error) {
	var n int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM user_follows WHERE follower_id = $1::uuid`, followerID).Scan(&n)
	return n, err
}

func (r *CommunityFollowRepo) RecommendUserIDs(ctx context.Context, viewerID string, limit int) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT u.id::text
		FROM users u
		WHERE u.deleted_at IS NULL
		  AND u.employment_status = 'active'
		  AND u.id <> $1::uuid
		  AND NOT EXISTS (
			SELECT 1 FROM user_follows f
			WHERE f.follower_id = $1::uuid AND f.followee_id = u.id
		  )
		ORDER BY
		  EXISTS (
			SELECT 1 FROM community_group_members m1
			INNER JOIN community_group_members m2 ON m1.group_id = m2.group_id
			WHERE m1.user_id = $1::uuid AND m2.user_id = u.id
		  ) DESC,
		  (SELECT COUNT(*)::int FROM user_follows WHERE followee_id = u.id) DESC,
		  CASE WHEN u.role::text = 'instructor' THEN 1 ELSE 0 END DESC,
		  u.created_at ASC
		LIMIT $2`,
		viewerID, limit)
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
