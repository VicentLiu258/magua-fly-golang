package postgres

import (
	"context"
	"fmt"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"

	"github.com/jackc/pgx/v5"
)

type CommunityChallengeRepo struct{ db *DB }

func (db *DB) CommunityChallenge() repository.CommunityChallengeRepository {
	return &CommunityChallengeRepo{db: db}
}

var _ repository.CommunityChallengeRepository = (*CommunityChallengeRepo)(nil)

func scanChallenge(row rowScanner) (*domain.CommunityChallenge, error) {
	var ch domain.CommunityChallenge
	var desc, rules, cover *string
	var st string
	err := row.Scan(
		&ch.ID, &ch.Title, &desc, &rules, &cover, &ch.StartsAt, &ch.EndsAt, &st, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc != nil {
		ch.Description = *desc
	}
	if rules != nil {
		ch.Rules = *rules
	}
	if cover != nil {
		ch.CoverURL = *cover
	}
	ch.Status = domain.CommunityChallengeStatus(st)
	return &ch, nil
}

func (r *CommunityChallengeRepo) Create(ctx context.Context, ch *domain.CommunityChallenge) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO community_challenges (title, description, rules, cover_url, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, created_at, updated_at`,
		ch.Title, nullStr(ch.Description), nullStr(ch.Rules), nullStr(ch.CoverURL),
		ch.StartsAt, ch.EndsAt, string(ch.Status),
	).Scan(&ch.ID, &ch.CreatedAt, &ch.UpdatedAt)
}

func (r *CommunityChallengeRepo) Update(ctx context.Context, ch *domain.CommunityChallenge) error {
	ct, err := r.db.Pool.Exec(ctx, `
		UPDATE community_challenges SET
			title = $2, description = $3, rules = $4, cover_url = $5,
			starts_at = $6, ends_at = $7, status = $8, updated_at = NOW()
		WHERE id = $1::uuid`,
		ch.ID, ch.Title, nullStr(ch.Description), nullStr(ch.Rules), nullStr(ch.CoverURL),
		ch.StartsAt, ch.EndsAt, string(ch.Status),
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CommunityChallengeRepo) GetByID(ctx context.Context, id string) (*domain.CommunityChallenge, error) {
	return scanChallenge(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, title, description, rules, cover_url, starts_at, ends_at, status::text, created_at, updated_at
		FROM community_challenges WHERE id = $1::uuid`, id))
}

func (r *CommunityChallengeRepo) List(ctx context.Context, status *domain.CommunityChallengeStatus, limit, offset int) ([]*domain.CommunityChallenge, error) {
	q := `SELECT id::text, title, description, rules, cover_url, starts_at, ends_at, status::text, created_at, updated_at
		FROM community_challenges`
	args := []any{}
	n := 1
	if status != nil {
		q += fmt.Sprintf(` WHERE status = $%d`, n)
		args = append(args, string(*status))
		n++
	}
	q += fmt.Sprintf(` ORDER BY starts_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityChallenge
	for rows.Next() {
		ch, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (r *CommunityChallengeRepo) SubmitEntry(ctx context.Context, e *domain.CommunityChallengeEntry) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO community_challenge_entries (challenge_id, user_id, post_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid)
		ON CONFLICT (challenge_id, user_id) DO UPDATE SET post_id = EXCLUDED.post_id, submitted_at = NOW()`,
		e.ChallengeID, e.UserID, e.PostID)
	return err
}

func (r *CommunityChallengeRepo) GetEntry(ctx context.Context, challengeID, userID string) (*domain.CommunityChallengeEntry, error) {
	var e domain.CommunityChallengeEntry
	err := r.db.Pool.QueryRow(ctx, `
		SELECT challenge_id::text, user_id::text, post_id::text, submitted_at
		FROM community_challenge_entries WHERE challenge_id = $1::uuid AND user_id = $2::uuid`,
		challengeID, userID).Scan(&e.ChallengeID, &e.UserID, &e.PostID, &e.SubmittedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *CommunityChallengeRepo) ListEntries(ctx context.Context, challengeID string, limit, offset int) ([]*domain.CommunityChallengeEntry, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT challenge_id::text, user_id::text, post_id::text, submitted_at
		FROM community_challenge_entries WHERE challenge_id = $1::uuid
		ORDER BY submitted_at DESC LIMIT $2 OFFSET $3`,
		challengeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.CommunityChallengeEntry
	for rows.Next() {
		var e domain.CommunityChallengeEntry
		if err := rows.Scan(&e.ChallengeID, &e.UserID, &e.PostID, &e.SubmittedAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}
