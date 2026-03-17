package mongodb

import (
	"context"
	"time"

	"magua-fly-golang/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
)

var _ repository.ContractFileMetaRepository = (*ContractFileMetaRepo)(nil)

// ContractFileMetaRepo implements repository.ContractFileMetaRepository.
type ContractFileMetaRepo struct {
	c *Client
}

// ContractFilesMeta returns a repository for contract file metadata.
func (c *Client) ContractFilesMeta() repository.ContractFileMetaRepository {
	return &ContractFileMetaRepo{c: c}
}

func (r *ContractFileMetaRepo) Insert(ctx context.Context, doc *repository.ContractFileMeta) error {
	coll := r.c.DB.Collection(CollContractFilesMeta)
	doc.UploadedAt = time.Now().Unix()
	_, err := coll.InsertOne(ctx, doc)
	return err
}

func (r *ContractFileMetaRepo) ListByContract(ctx context.Context, contractID string) ([]*repository.ContractFileMeta, error) {
	coll := r.c.DB.Collection(CollContractFilesMeta)
	cur, err := coll.Find(ctx, bson.M{"contract_id": contractID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var res []*repository.ContractFileMeta
	for cur.Next(ctx) {
		var doc repository.ContractFileMeta
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		res = append(res, &doc)
	}
	return res, cur.Err()
}

func (r *ContractFileMetaRepo) GetByID(ctx context.Context, id string) (*repository.ContractFileMeta, error) {
	coll := r.c.DB.Collection(CollContractFilesMeta)
	var doc repository.ContractFileMeta
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

