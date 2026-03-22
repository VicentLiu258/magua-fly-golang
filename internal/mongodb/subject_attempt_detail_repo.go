package mongodb

import (
	"context"
	"time"

	"magua-fly-golang/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.SubjectAttemptDetailRepository = (*SubjectAttemptDetailRepo)(nil)

// SubjectAttemptDetailRepo implements repository.SubjectAttemptDetailRepository.
type SubjectAttemptDetailRepo struct {
	c *Client
}

// SubjectAttemptDetail returns a repository for subject attempt details.
func (c *Client) SubjectAttemptDetail() repository.SubjectAttemptDetailRepository {
	return &SubjectAttemptDetailRepo{c: c}
}

func (r *SubjectAttemptDetailRepo) GetByAttemptID(ctx context.Context, attemptID string) (*repository.SubjectAttemptDetailDoc, error) {
	coll := r.c.DB.Collection(CollSubjectAttemptDetails)
	var doc repository.SubjectAttemptDetailDoc
	err := coll.FindOne(ctx, bson.M{"subject_attempt_id": attemptID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *SubjectAttemptDetailRepo) Upsert(ctx context.Context, doc *repository.SubjectAttemptDetailDoc) error {
	coll := r.c.DB.Collection(CollSubjectAttemptDetails)
	if doc.CreatedAt == 0 {
		doc.CreatedAt = time.Now().Unix()
	}
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"subject_attempt_id": doc.SubjectAttemptID},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	return err
}
