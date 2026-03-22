package mongodb

import (
	"context"
	"time"

	"magua-fly-golang/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.SubjectQuestionSetRepository = (*SubjectQuestionSetRepo)(nil)

// SubjectQuestionSetRepo implements repository.SubjectQuestionSetRepository.
type SubjectQuestionSetRepo struct {
	c *Client
}

// SubjectQuestionSet returns a subject question set repository.
func (c *Client) SubjectQuestionSet() repository.SubjectQuestionSetRepository {
	return &SubjectQuestionSetRepo{c: c}
}

func (r *SubjectQuestionSetRepo) GetBySubjectLearningID(ctx context.Context, subjectLearningID string) (*repository.SubjectQuestionSetDoc, error) {
	coll := r.c.DB.Collection(CollSubjectQuestionSets)
	var doc repository.SubjectQuestionSetDoc
	err := coll.FindOne(ctx, bson.M{"subject_learning_id": subjectLearningID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *SubjectQuestionSetRepo) Upsert(ctx context.Context, doc *repository.SubjectQuestionSetDoc) error {
	coll := r.c.DB.Collection(CollSubjectQuestionSets)
	doc.UpdatedAt = time.Now().Unix()
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"subject_learning_id": doc.SubjectLearningID},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	return err
}
