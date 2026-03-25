package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SyllabusSubjectSecondaryRepo struct{ db *DB }

func (db *DB) SyllabusSubjectSecondary() repository.SyllabusSubjectSecondaryRepository {
	return &SyllabusSubjectSecondaryRepo{db: db}
}

var _ repository.SyllabusSubjectSecondaryRepository = (*SyllabusSubjectSecondaryRepo)(nil)

func (r *SyllabusSubjectSecondaryRepo) GetByID(ctx context.Context, id string) (*domain.SyllabusSubjectSecondary, error) {
	var sec domain.SyllabusSubjectSecondary
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, syllabus_subject_id::text, seq_no, name, COALESCE(description,''), COALESCE(completion_standard,''), sort_order, created_at, updated_at
		FROM syllabus_subject_secondaries
		WHERE id = $1::uuid`, id,
	).Scan(&sec.ID, &sec.SyllabusSubjectID, &sec.SeqNo, &sec.Name, &sec.Description, &sec.CompletionStandard, &sec.SortOrder, &sec.CreatedAt, &sec.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

func (r *SyllabusSubjectSecondaryRepo) ListBySubjectID(ctx context.Context, syllabusSubjectID string) ([]*domain.SyllabusSubjectSecondary, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, syllabus_subject_id::text, seq_no, name, COALESCE(description,''), COALESCE(completion_standard,''), sort_order, created_at, updated_at
		FROM syllabus_subject_secondaries
		WHERE syllabus_subject_id = $1::uuid
		ORDER BY sort_order, id`, syllabusSubjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSecondaries(rows)
}

func (r *SyllabusSubjectSecondaryRepo) ListGroupedBySubjectID(ctx context.Context, syllabusSubjectIDs []string) (map[string][]*domain.SyllabusSubjectSecondary, error) {
	out := make(map[string][]*domain.SyllabusSubjectSecondary)
	if len(syllabusSubjectIDs) == 0 {
		return out, nil
	}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, syllabus_subject_id::text, seq_no, name, COALESCE(description,''), COALESCE(completion_standard,''), sort_order, created_at, updated_at
		FROM syllabus_subject_secondaries
		WHERE syllabus_subject_id = ANY($1::uuid[])
		ORDER BY syllabus_subject_id, sort_order, id`, syllabusSubjectIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sec domain.SyllabusSubjectSecondary
		if err := rows.Scan(&sec.ID, &sec.SyllabusSubjectID, &sec.SeqNo, &sec.Name, &sec.Description, &sec.CompletionStandard, &sec.SortOrder, &sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, err
		}
		out[sec.SyllabusSubjectID] = append(out[sec.SyllabusSubjectID], &sec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func scanSecondaries(rows pgx.Rows) ([]*domain.SyllabusSubjectSecondary, error) {
	var list []*domain.SyllabusSubjectSecondary
	for rows.Next() {
		var sec domain.SyllabusSubjectSecondary
		if err := rows.Scan(&sec.ID, &sec.SyllabusSubjectID, &sec.SeqNo, &sec.Name, &sec.Description, &sec.CompletionStandard, &sec.SortOrder, &sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &sec)
	}
	return list, rows.Err()
}

func (r *SyllabusSubjectSecondaryRepo) Create(ctx context.Context, row *domain.SyllabusSubjectSecondary) error {
	if row.SyllabusSubjectID == "" {
		return fmt.Errorf("syllabus_subject_id required")
	}
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO syllabus_subject_secondaries (syllabus_subject_id, seq_no, name, description, completion_standard, sort_order)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at, updated_at`,
		row.SyllabusSubjectID, row.SeqNo, row.Name, secondaryStrOrNil(row.Description), secondaryStrOrNil(row.CompletionStandard), row.SortOrder,
	).Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt)
}

func secondaryStrOrNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *SyllabusSubjectSecondaryRepo) Update(ctx context.Context, row *domain.SyllabusSubjectSecondary) error {
	ct, err := r.db.Pool.Exec(ctx, `
		UPDATE syllabus_subject_secondaries
		SET seq_no = $2, name = $3, description = $4, completion_standard = $5, sort_order = $6, updated_at = NOW()
		WHERE id = $1::uuid`,
		row.ID, row.SeqNo, row.Name, secondaryStrOrNil(row.Description), secondaryStrOrNil(row.CompletionStandard), row.SortOrder,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SyllabusSubjectSecondaryRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM syllabus_subject_secondaries WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SyllabusSubjectSecondaryRepo) Reorder(ctx context.Context, syllabusSubjectID string, orderedIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var n int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM syllabus_subject_secondaries WHERE syllabus_subject_id = $1::uuid`, syllabusSubjectID).Scan(&n)
	if err != nil {
		return err
	}
	if n != len(orderedIDs) {
		return fmt.Errorf("reorder: expected %d ids, got %d", n, len(orderedIDs))
	}

	for i, id := range orderedIDs {
		ct, err := tx.Exec(ctx, `
			UPDATE syllabus_subject_secondaries
			SET sort_order = $3, updated_at = NOW()
			WHERE id = $1::uuid AND syllabus_subject_id = $2::uuid`,
			id, syllabusSubjectID, i)
		if err != nil {
			return err
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("reorder: invalid id or subject mismatch: %s", id)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
