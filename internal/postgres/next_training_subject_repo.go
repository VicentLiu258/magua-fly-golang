package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type NextTrainingSubjectRepo struct{ db *DB }

func (db *DB) NextTrainingSubject() repository.NextTrainingSubjectRepository {
	return &NextTrainingSubjectRepo{db: db}
}

var _ repository.NextTrainingSubjectRepository = (*NextTrainingSubjectRepo)(nil)

func (r *NextTrainingSubjectRepo) ListByArchiveID(ctx context.Context, archiveID string) ([]*domain.NextTrainingSubject, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, training_archive_id::text, syllabus_subject_id::text, assigned_by::text, created_at
		FROM next_training_subjects WHERE training_archive_id = $1::uuid`, archiveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.NextTrainingSubject
	for rows.Next() {
		var s domain.NextTrainingSubject
		if err := rows.Scan(&s.ID, &s.TrainingArchiveID, &s.SyllabusSubjectID, &s.AssignedBy, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *NextTrainingSubjectRepo) ReplaceForArchive(ctx context.Context, archiveID string, assignedBy string, syllabusSubjectIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM next_training_subjects WHERE training_archive_id = $1::uuid`, archiveID); err != nil {
		return err
	}
	for _, sid := range syllabusSubjectIDs {
		_, err := tx.Exec(ctx, `INSERT INTO next_training_subjects (training_archive_id, syllabus_subject_id, assigned_by) VALUES ($1::uuid, $2::uuid, $3::uuid)`, archiveID, sid, assignedBy)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
