package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// UserRepo implements repository.UserRepository.
type UserRepo struct{ db *DB }

func (db *DB) User() repository.UserRepository { return &UserRepo{db: db} }

var _ repository.UserRepository = (*UserRepo)(nil)

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	var roleStr string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE id = $1::uuid`, id).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Phone, &u.Email, &roleStr, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	var roleStr string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE username = $1`, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Phone, &u.Email, &roleStr, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, phone, email, role, avatar_url)
		VALUES ($1, $2, $3, $4, $5::user_role, $6)
		RETURNING id::text, created_at, updated_at`,
		u.Username, u.PasswordHash, u.Phone, u.Email, string(u.Role), u.AvatarURL,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var u domain.User
	var roleStr string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE phone = $1`, phone).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Phone, &u.Email, &roleStr, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	return &u, nil
}
