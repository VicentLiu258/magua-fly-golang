package postgres

import (
	"context"
	"errors"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"

	"github.com/jackc/pgx/v5"
)

type CommunityGroupRepo struct{ db *DB }

func (db *DB) CommunityGroup() repository.CommunityGroupRepository {
	return &CommunityGroupRepo{db: db}
}

var _ repository.CommunityGroupRepository = (*CommunityGroupRepo)(nil)

func scanGroup(row rowScanner) (*domain.CommunityGroup, error) {
	var g domain.CommunityGroup
	var desc, cover *string
	err := row.Scan(
		&g.ID, &g.Name, &g.Slug, &desc, &cover, &g.OwnerUserID, &g.IsPublic, &g.MemberCount, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc != nil {
		g.Description = *desc
	}
	if cover != nil {
		g.CoverURL = *cover
	}
	return &g, nil
}

func (r *CommunityGroupRepo) Create(ctx context.Context, g *domain.CommunityGroup) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO community_groups (name, slug, description, cover_url, owner_user_id, is_public)
		VALUES ($1, $2, $3, $4, $5::uuid, $6)
		RETURNING id::text, member_count, created_at, updated_at`,
		g.Name, g.Slug, nullStr(g.Description), nullStr(g.CoverURL), g.OwnerUserID, g.IsPublic,
	).Scan(&g.ID, &g.MemberCount, &g.CreatedAt, &g.UpdatedAt)
}

func (r *CommunityGroupRepo) Update(ctx context.Context, g *domain.CommunityGroup) error {
	ct, err := r.db.Pool.Exec(ctx, `
		UPDATE community_groups SET
			name = $2, slug = $3, description = $4, cover_url = $5, is_public = $6, updated_at = NOW()
		WHERE id = $1::uuid`,
		g.ID, g.Name, g.Slug, nullStr(g.Description), nullStr(g.CoverURL), g.IsPublic,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CommunityGroupRepo) GetByID(ctx context.Context, id string) (*domain.CommunityGroup, error) {
	return scanGroup(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, name, slug, description, cover_url, owner_user_id::text, is_public, member_count, created_at, updated_at
		FROM community_groups WHERE id = $1::uuid`, id))
}

func (r *CommunityGroupRepo) GetBySlug(ctx context.Context, slug string) (*domain.CommunityGroup, error) {
	return scanGroup(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, name, slug, description, cover_url, owner_user_id::text, is_public, member_count, created_at, updated_at
		FROM community_groups WHERE slug = $1`, slug))
}

func (r *CommunityGroupRepo) List(ctx context.Context, limit, offset int) ([]*domain.CommunityGroup, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, name, slug, description, cover_url, owner_user_id::text, is_public, member_count, created_at, updated_at
		FROM community_groups ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityGroup
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *CommunityGroupRepo) AddMember(ctx context.Context, groupID, userID string, role domain.CommunityGroupMemberRole) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO community_group_members (group_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, $3)
		ON CONFLICT (group_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		groupID, userID, string(role))
	return err
}

func (r *CommunityGroupRepo) RemoveMember(ctx context.Context, groupID, userID string) error {
	ct, err := r.db.Pool.Exec(ctx, `
		DELETE FROM community_group_members WHERE group_id = $1::uuid AND user_id = $2::uuid`,
		groupID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CommunityGroupRepo) GetMembership(ctx context.Context, groupID, userID string) (*domain.CommunityGroupMember, error) {
	var m domain.CommunityGroupMember
	var role string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT group_id::text, user_id::text, role::text, joined_at
		FROM community_group_members WHERE group_id = $1::uuid AND user_id = $2::uuid`,
		groupID, userID).Scan(&m.GroupID, &m.UserID, &role, &m.JoinedAt)
	if err != nil {
		return nil, err
	}
	m.Role = domain.CommunityGroupMemberRole(role)
	return &m, nil
}

func (r *CommunityGroupRepo) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	m, err := r.GetMembership(ctx, groupID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return m != nil, nil
}

func (r *CommunityGroupRepo) ListMembers(ctx context.Context, groupID string, limit, offset int) ([]*domain.CommunityGroupMember, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT group_id::text, user_id::text, role::text, joined_at
		FROM community_group_members WHERE group_id = $1::uuid
		ORDER BY joined_at ASC LIMIT $2 OFFSET $3`,
		groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityGroupMember
	for rows.Next() {
		var m domain.CommunityGroupMember
		var role string
		if err := rows.Scan(&m.GroupID, &m.UserID, &role, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.Role = domain.CommunityGroupMemberRole(role)
		out = append(out, &m)
	}
	return out, rows.Err()
}
