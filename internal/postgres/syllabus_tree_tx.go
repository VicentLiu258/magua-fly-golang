package postgres

import (
	"context"
	"fmt"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type SyllabusTreeRepo struct{ db *DB }

func (db *DB) SyllabusTree() repository.SyllabusTreeRepository {
	return &SyllabusTreeRepo{db: db}
}

var _ repository.SyllabusTreeRepository = (*SyllabusTreeRepo)(nil)

func (r *SyllabusTreeRepo) ImportOutlineTree(ctx context.Context, outline *domain.SyllabusOutline, subjects []domain.SyllabusImportSubjectRow) (*domain.SyllabusOutline, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO syllabus_outlines (category, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id::text, created_at, updated_at`,
		outline.Category, outline.Name, outline.SortOrder,
	).Scan(&outline.ID, &outline.CreatedAt, &outline.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for _, sub := range subjects {
		var sid string
		err = tx.QueryRow(ctx, `
			INSERT INTO syllabus_subjects (syllabus_outline_id, code, name, sort_order)
			VALUES ($1::uuid, $2, $3, $4)
			RETURNING id::text`,
			outline.ID, sub.Code, sub.Name, sub.SortOrder,
		).Scan(&sid)
		if err != nil {
			return nil, err
		}
		for _, sec := range sub.Secondaries {
			_, err = tx.Exec(ctx, `
				INSERT INTO syllabus_subject_secondaries (syllabus_subject_id, seq_no, name, description, completion_standard, sort_order)
				VALUES ($1::uuid, $2, $3, $4, $5, $6)`,
				sid, sec.SeqNo, sec.Name, nullStrPtr(sec.Description), nullStrPtr(sec.CompletionStandard), sec.SortOrder,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return outline, nil
}

func nullStrPtr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *SyllabusTreeRepo) CreateSubjectUnderOutline(ctx context.Context, outlineID *string, outline *domain.SyllabusOutline, subject *domain.SyllabusSubject, secondaries []*domain.SyllabusSubjectSecondary) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oid string
	if outlineID != nil && *outlineID != "" {
		oid = *outlineID
	} else if outline != nil {
		err = tx.QueryRow(ctx, `
			INSERT INTO syllabus_outlines (category, name, sort_order)
			VALUES ($1, $2, $3)
			RETURNING id::text`,
			outline.Category, outline.Name, outline.SortOrder,
		).Scan(&oid)
		if err != nil {
			return err
		}
		outline.ID = oid
	} else {
		return fmt.Errorf("syllabus_outline_id or outline (category+name) required")
	}

	subject.SyllabusOutlineID = oid
	err = tx.QueryRow(ctx, `
		INSERT INTO syllabus_subjects (syllabus_outline_id, code, name, sort_order)
		VALUES ($1::uuid, $2, $3, $4)
		RETURNING id::text, created_at, updated_at`,
		subject.SyllabusOutlineID, subject.Code, subject.Name, subject.SortOrder,
	).Scan(&subject.ID, &subject.CreatedAt, &subject.UpdatedAt)
	if err != nil {
		return err
	}

	for _, sec := range secondaries {
		sec.SyllabusSubjectID = subject.ID
		_, err = tx.Exec(ctx, `
			INSERT INTO syllabus_subject_secondaries (syllabus_subject_id, seq_no, name, description, completion_standard, sort_order)
			VALUES ($1::uuid, $2, $3, $4, $5, $6)`,
			sec.SyllabusSubjectID, sec.SeqNo, sec.Name, nullStrPtr(sec.Description), nullStrPtr(sec.CompletionStandard), sec.SortOrder,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
