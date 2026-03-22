package postgres

import (
	"context"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SubjectAttemptRepo struct{ db *DB }

func (db *DB) SubjectAttempt() repository.SubjectAttemptRepository {
	return &SubjectAttemptRepo{db: db}
}

var _ repository.SubjectAttemptRepository = (*SubjectAttemptRepo)(nil)

func (r *SubjectAttemptRepo) Create(ctx context.Context, a *domain.SubjectAttempt) error {
	if a.AnswersJSON == "" {
		a.AnswersJSON = "[]"
	}
	if a.SubmittedAt.IsZero() {
		a.SubmittedAt = time.Now()
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO subject_attempts (user_id, subject_learning_id, answers, score, submitted_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		RETURNING id::text`,
		a.UserID, a.SubjectLearningID, a.AnswersJSON, a.Score, a.SubmittedAt,
	).Scan(&a.ID)
}

func (r *SubjectAttemptRepo) UpdateOnSubmit(ctx context.Context, a *domain.SubjectAttempt) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE subject_attempts
		SET answers = $3, score = $4, submitted_at = $5,
		    max_score = $6, correct_count = $7, total_count = $8, passed = $9
		WHERE id = $1::uuid AND user_id = $2::uuid`,
		a.ID, a.UserID, a.AnswersJSON, a.Score, a.SubmittedAt,
		a.MaxScore, a.CorrectCount, a.TotalCount, a.Passed,
	)
	return err
}

func (r *SubjectAttemptRepo) ListByUserAndSubjectLearning(ctx context.Context, userID, subjectLearningID string, limit, offset int) ([]*domain.SubjectAttempt, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, user_id::text, subject_learning_id::text,
		       score, submitted_at,
		       COALESCE(duration_seconds, 0),
		       COALESCE(max_score, 0), COALESCE(correct_count, 0), COALESCE(total_count, 0),
		       COALESCE(passed, false)
		FROM subject_attempts
		WHERE user_id = $1::uuid AND subject_learning_id = $2::uuid
		ORDER BY submitted_at DESC
		LIMIT $3 OFFSET $4`,
		userID, subjectLearningID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.SubjectAttempt
	for rows.Next() {
		var a domain.SubjectAttempt
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.SubjectLearningID,
			&a.Score, &a.SubmittedAt,
			&a.DurationSeconds,
			&a.MaxScore, &a.CorrectCount, &a.TotalCount,
			&a.Passed,
		); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}
