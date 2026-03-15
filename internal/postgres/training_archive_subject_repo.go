package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TrainingArchiveSubjectRepo struct{ db *DB }

func (db *DB) TrainingArchiveSubject() repository.TrainingArchiveSubjectRepository {
	return &TrainingArchiveSubjectRepo{db: db}
}

var _ repository.TrainingArchiveSubjectRepository = (*TrainingArchiveSubjectRepo)(nil)

func (r *TrainingArchiveSubjectRepo) ListByArchiveID(ctx context.Context, archiveID int64) ([]*domain.TrainingArchiveSubject, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, training_archive_id, syllabus_subject_id, evaluation_score, is_graduated, sort_order,
		       COALESCE(subject_content,''), COALESCE(type,''), created_at, updated_at
		FROM training_archive_subjects WHERE training_archive_id = $1 ORDER BY sort_order, id`, archiveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.TrainingArchiveSubject
	for rows.Next() {
		var s domain.TrainingArchiveSubject
		if err := rows.Scan(&s.ID, &s.TrainingArchiveID, &s.SyllabusSubjectID, &s.EvaluationScore, &s.IsGraduated, &s.SortOrder, &s.SubjectContent, &s.Type, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *TrainingArchiveSubjectRepo) ReplaceForArchive(ctx context.Context, archiveID int64, subjects []*domain.TrainingArchiveSubject) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM training_archive_subjects WHERE training_archive_id = $1`, archiveID); err != nil {
		return err
	}
	for i, s := range subjects {
		_, err := tx.Exec(ctx, `
			INSERT INTO training_archive_subjects (training_archive_id, syllabus_subject_id, evaluation_score, is_graduated, sort_order, subject_content, type)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''))`,
			archiveID, s.SyllabusSubjectID, s.EvaluationScore, s.IsGraduated, i, s.SubjectContent, s.Type,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
