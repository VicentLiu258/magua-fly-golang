package mongodb

import (
	"context"
	"time"

	"magua-fly-golang/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.TheoryAttemptDetailRepository = (*TheoryAttemptDetailRepo)(nil)

// TheoryAttemptDetailRepo implements repository.TheoryAttemptDetailRepository.
type TheoryAttemptDetailRepo struct {
	c *Client
}

// TheoryAttemptDetail returns a repository for theory attempt details.
func (c *Client) TheoryAttemptDetail() repository.TheoryAttemptDetailRepository {
	return &TheoryAttemptDetailRepo{c: c}
}

func (r *TheoryAttemptDetailRepo) GetByAttemptID(ctx context.Context, attemptID string) (*repository.TheoryAttemptDetailDoc, error) {
	coll := r.c.DB.Collection(CollTheoryAttemptDetails)
	var doc repository.TheoryAttemptDetailDoc
	err := coll.FindOne(ctx, bson.M{"theory_attempt_id": attemptID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *TheoryAttemptDetailRepo) Upsert(ctx context.Context, doc *repository.TheoryAttemptDetailDoc) error {
	coll := r.c.DB.Collection(CollTheoryAttemptDetails)
	if doc.CreatedAt == 0 {
		doc.CreatedAt = time.Now().Unix()
	}
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"theory_attempt_id": doc.TheoryAttemptID},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	return err
}

