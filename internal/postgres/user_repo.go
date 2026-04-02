package postgres

import (
	"context"
	"fmt"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// UserRepo implements repository.UserRepository.
type UserRepo struct{ db *DB }

func (db *DB) User() repository.UserRepository { return &UserRepo{db: db} }

var _ repository.UserRepository = (*UserRepo)(nil)

func employmentOrActive(e domain.EmploymentStatus) domain.EmploymentStatus {
	if e == "" {
		return domain.EmploymentActive
	}
	return e
}

func scanUser(row interface{ Scan(dest ...any) error }) (*domain.User, error) {
	var u domain.User
	var roleStr, empStr string
	err := row.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Phone, &u.Email, &roleStr, &empStr, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = domain.UserRole(roleStr)
	u.EmploymentStatus = domain.EmploymentStatus(empStr)
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text,
		       employment_status::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE id = $1::uuid AND deleted_at IS NULL`, id))
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text,
		       employment_status::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE username = $1 AND deleted_at IS NULL`, username))
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	emp := employmentOrActive(u.EmploymentStatus)
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, phone, email, role, avatar_url, employment_status)
		VALUES ($1, $2, $3, $4, $5::user_role, $6, $7::employment_status)
		RETURNING id::text, created_at, updated_at`,
		u.Username, u.PasswordHash, u.Phone, u.Email, string(u.Role), u.AvatarURL, string(emp),
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	emp := employmentOrActive(u.EmploymentStatus)
	tag, err := r.db.Pool.Exec(ctx, `
		UPDATE users
		SET username = $2,
		    phone = $3,
		    email = $4,
		    avatar_url = NULLIF(TRIM($5), ''),
		    password_hash = $6,
		    employment_status = $7::employment_status,
		    updated_at = NOW()
		WHERE id = $1::uuid AND deleted_at IS NULL`,
		u.ID, u.Username, u.Phone, u.Email, u.AvatarURL, u.PasswordHash, string(emp),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text,
		       employment_status::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE phone = $1 AND deleted_at IS NULL`, phone))
}

func (r *UserRepo) ListByRoles(ctx context.Context, roles []domain.UserRole) ([]*domain.User, error) {
	if len(roles) == 0 {
		return nil, fmt.Errorf("roles required")
	}
	roleStrs := make([]string, len(roles))
	for i, rl := range roles {
		roleStrs[i] = string(rl)
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, username, password_hash, COALESCE(phone,''), COALESCE(email,''), role::text,
		       employment_status::text, COALESCE(avatar_url,''), created_at, updated_at
		FROM users
		WHERE role::text = ANY($1::text[])
		  AND deleted_at IS NULL
		  AND employment_status = 'active'
		ORDER BY role, username`, roleStrs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}
