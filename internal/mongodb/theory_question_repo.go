package mongodb

import (
	"context"
	"time"

	"magua-fly-golang/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.TheoryQuestionSetRepository = (*TheoryQuestionSetRepo)(nil)

// TheoryQuestionSetRepo implements repository.TheoryQuestionSetRepository.
type TheoryQuestionSetRepo struct {
	c *Client
}

// TheoryQuestionSet returns a theory question set repository.
func (c *Client) TheoryQuestionSet() repository.TheoryQuestionSetRepository {
	return &TheoryQuestionSetRepo{c: c}
}

func (r *TheoryQuestionSetRepo) GetByDocumentID(ctx context.Context, documentID string) (*repository.TheoryQuestionSetDoc, error) {
	coll := r.c.DB.Collection(CollTheoryQuestionSets)
	var doc repository.TheoryQuestionSetDoc
	err := coll.FindOne(ctx, bson.M{"document_id": documentID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *TheoryQuestionSetRepo) Upsert(ctx context.Context, doc *repository.TheoryQuestionSetDoc) error {
	coll := r.c.DB.Collection(CollTheoryQuestionSets)
	doc.UpdatedAt = time.Now().Unix()
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"document_id": doc.DocumentID},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	return err
}

