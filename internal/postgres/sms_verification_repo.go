package postgres

import (
	"context"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// SmsVerificationRepo implements repository.SmsVerificationRepository.
type SmsVerificationRepo struct{ db *DB }

func (db *DB) SmsVerification() repository.SmsVerificationRepository {
	return &SmsVerificationRepo{db: db}
}

var _ repository.SmsVerificationRepository = (*SmsVerificationRepo)(nil)

func (r *SmsVerificationRepo) Create(ctx context.Context, v *domain.SmsVerificationCode) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO sms_verification_codes (phone, scene, code, ip, expired_at)
		VALUES ($1, $2::sms_scene, $3, $4, $5)
		RETURNING id::text, created_at
	`, v.Phone, string(v.Scene), v.Code, v.IP, v.ExpiredAt).Scan(&v.ID, &v.CreatedAt)
}

func (r *SmsVerificationRepo) FindLatestByPhoneScene(ctx context.Context, phone string, scene domain.SmsScene) (*domain.SmsVerificationCode, error) {
	var v domain.SmsVerificationCode
	var sceneStr string
	var usedAt *time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, phone, scene::text, code, ip, expired_at, used_at, created_at
		FROM sms_verification_codes
		WHERE phone = $1 AND scene = $2::sms_scene
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, string(scene)).Scan(
		&v.ID, &v.Phone, &sceneStr, &v.Code, &v.IP, &v.ExpiredAt, &usedAt, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	v.Scene = domain.SmsScene(sceneStr)
	v.UsedAt = usedAt
	return &v, nil
}

func (r *SmsVerificationRepo) FindValidCode(ctx context.Context, phone string, scene domain.SmsScene, code string, now time.Time) (*domain.SmsVerificationCode, error) {
	var v domain.SmsVerificationCode
	var sceneStr string
	var usedAt *time.Time
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, phone, scene::text, code, ip, expired_at, used_at, created_at
		FROM sms_verification_codes
		WHERE phone = $1
		  AND scene = $2::sms_scene
		  AND code = $3
		  AND used_at IS NULL
		  AND expired_at > $4
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, string(scene), code, now).Scan(
		&v.ID, &v.Phone, &sceneStr, &v.Code, &v.IP, &v.ExpiredAt, &usedAt, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	v.Scene = domain.SmsScene(sceneStr)
	v.UsedAt = usedAt
	return &v, nil
}

func (r *SmsVerificationRepo) MarkUsed(ctx context.Context, id string, usedAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE sms_verification_codes
		SET used_at = $1
		WHERE id = $2::uuid AND used_at IS NULL
	`, usedAt, id)
	return err
}

