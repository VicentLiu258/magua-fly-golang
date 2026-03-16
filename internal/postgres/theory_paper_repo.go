package postgres

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

type TheoryPaperRepo struct{ db *DB }

func (db *DB) TheoryPaper() repository.TheoryPaperRepository { return &TheoryPaperRepo{db: db} }

var _ repository.TheoryPaperRepository = (*TheoryPaperRepo)(nil)

func (r *TheoryPaperRepo) Create(ctx context.Context, p *domain.TheoryPaper) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO theory_papers (theory_document_id, name, description, total_score, enabled, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6::uuid)
		RETURNING id::text, created_at, updated_at`,
		p.TheoryDocumentID, p.Name, p.Description, p.TotalScore, p.Enabled, p.CreatedBy,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *TheoryPaperRepo) Update(ctx context.Context, p *domain.TheoryPaper) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE theory_papers
		SET name = $2, description = $3, total_score = $4, enabled = $5, updated_at = NOW()
		WHERE id = $1::uuid`,
		p.ID, p.Name, p.Description, p.TotalScore, p.Enabled,
	)
	return err
}

func (r *TheoryPaperRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM theory_papers WHERE id = $1::uuid`, id)
	return err
}

func (r *TheoryPaperRepo) GetByID(ctx context.Context, id string) (*domain.TheoryPaper, error) {
	var p domain.TheoryPaper
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id::text, theory_document_id::text, name, COALESCE(description,''), COALESCE(total_score,0),
		       enabled, created_by::text, created_at, updated_at
		FROM theory_papers WHERE id = $1::uuid`, id,
	).Scan(&p.ID, &p.TheoryDocumentID, &p.Name, &p.Description, &p.TotalScore,
		&p.Enabled, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *TheoryPaperRepo) ListByDocumentID(ctx context.Context, documentID string, limit, offset int) ([]*domain.TheoryPaper, int, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, theory_document_id::text, name, COALESCE(description,''), COALESCE(total_score,0),
		       enabled, created_by::text, created_at, updated_at
		FROM theory_papers
		WHERE theory_document_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		documentID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*domain.TheoryPaper
	for rows.Next() {
		var p domain.TheoryPaper
		if err := rows.Scan(&p.ID, &p.TheoryDocumentID, &p.Name, &p.Description, &p.TotalScore,
			&p.Enabled, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM theory_papers WHERE theory_document_id = $1::uuid`,
		documentID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

type TheoryPaperQuestionRepo struct{ db *DB }

func (db *DB) TheoryPaperQuestion() repository.TheoryPaperQuestionRepository {
	return &TheoryPaperQuestionRepo{db: db}
}

var _ repository.TheoryPaperQuestionRepository = (*TheoryPaperQuestionRepo)(nil)

func (r *TheoryPaperQuestionRepo) ReplaceForPaper(ctx context.Context, paperID string, questions []*domain.TheoryPaperQuestion) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM theory_paper_questions WHERE paper_id = $1::uuid`, paperID); err != nil {
		return err
	}

	for _, q := range questions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO theory_paper_questions (paper_id, question_id, score, sort_order)
			VALUES ($1::uuid, $2, $3, $4)`,
			paperID, q.QuestionID, q.Score, q.SortOrder,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *TheoryPaperQuestionRepo) ListByPaperID(ctx context.Context, paperID string) ([]*domain.TheoryPaperQuestion, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id::text, paper_id::text, question_id, score, sort_order
		FROM theory_paper_questions
		WHERE paper_id = $1::uuid
		ORDER BY sort_order, id`,
		paperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.TheoryPaperQuestion
	for rows.Next() {
		var q domain.TheoryPaperQuestion
		if err := rows.Scan(&q.ID, &q.PaperID, &q.QuestionID, &q.Score, &q.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, &q)
	}
	return list, rows.Err()
}

