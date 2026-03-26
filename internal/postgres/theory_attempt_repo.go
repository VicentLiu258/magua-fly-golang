package postgres

import (
	"context"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TheoryAttemptRepo struct{ db *DB }

func (db *DB) TheoryAttempt() repository.TheoryAttemptRepository { return &TheoryAttemptRepo{db: db} }

var _ repository.TheoryAttemptRepository = (*TheoryAttemptRepo)(nil)

func (r *TheoryAttemptRepo) Create(ctx context.Context, a *domain.TheoryAttempt) error {
	// Ensure we always persist a valid JSON payload for answers.
	if a.AnswersJSON == "" {
		a.AnswersJSON = "[]"
	}
	if a.SubmittedAt.IsZero() {
		a.SubmittedAt = time.Now()
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO theory_attempts (user_id, theory_document_id, answers, score, submitted_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		RETURNING id::text`,
		a.UserID, a.TheoryDocumentID, a.AnswersJSON, a.Score, a.SubmittedAt,
	).Scan(&a.ID)
}

func (r *TheoryAttemptRepo) UpdateOnSubmit(ctx context.Context, a *domain.TheoryAttempt) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE theory_attempts
		SET answers = $3, score = $4, submitted_at = $5
		WHERE id = $1::uuid AND user_id = $2::uuid`,
		a.ID, a.UserID, a.AnswersJSON, a.Score, a.SubmittedAt,
	)
	return err
}

func (r *TheoryAttemptRepo) GetByIDAndUser(ctx context.Context, attemptID, userID string) (*domain.TheoryAttempt, error) {
	var a domain.TheoryAttempt
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, theory_document_id::text,
		       score, submitted_at
		FROM theory_attempts
		WHERE id = $1::uuid AND user_id = $2::uuid`,
		attemptID, userID,
	).Scan(&a.ID, &a.UserID, &a.TheoryDocumentID, &a.Score, &a.SubmittedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *TheoryAttemptRepo) ListByUserAndDocument(ctx context.Context, userID, documentID string, limit, offset int) ([]*domain.TheoryAttempt, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, user_id::text, theory_document_id::text,
		       score, submitted_at
		FROM theory_attempts
		WHERE user_id = $1::uuid AND theory_document_id = $2::uuid
		ORDER BY submitted_at DESC
		LIMIT $3 OFFSET $4`,
		userID, documentID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.TheoryAttempt
	for rows.Next() {
		var a domain.TheoryAttempt
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.TheoryDocumentID,
			&a.Score, &a.SubmittedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

