package mongodb

import (
	"context"
	"time"

	"magua-fly-golang/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.TheoryContentRepository = (*TheoryContentRepo)(nil)

// TheoryContentRepo implements repository.TheoryContentRepository.
type TheoryContentRepo struct {
	c *Client
}

func (c *Client) TheoryContent() repository.TheoryContentRepository {
	return &TheoryContentRepo{c: c}
}

func (r *TheoryContentRepo) GetByDocumentID(ctx context.Context, documentID string) (*repository.TheoryContentDoc, error) {
	coll := r.c.DB.Collection(CollTheoryContent)
	var doc repository.TheoryContentDoc
	err := coll.FindOne(ctx, bson.M{"document_id": documentID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *TheoryContentRepo) Upsert(ctx context.Context, doc *repository.TheoryContentDoc) error {
	coll := r.c.DB.Collection(CollTheoryContent)
	doc.UpdatedAt = time.Now().Unix()
	_, err := coll.UpdateOne(ctx,
		bson.M{"document_id": doc.DocumentID},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	return err
}
