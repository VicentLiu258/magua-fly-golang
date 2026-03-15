package repository

import "context"

// TheoryContentDoc is stored in MongoDB (rich content for theory document).
type TheoryContentDoc struct {
	DocumentID string `bson:"document_id"`
	Body       string `bson:"body"`
	VideoEmbed string `bson:"video_embed,omitempty"`
	UpdatedAt  int64  `bson:"updated_at"`
}

// TheoryContentRepository defines theory content (MongoDB) access.
type TheoryContentRepository interface {
	GetByDocumentID(ctx context.Context, documentID string) (*TheoryContentDoc, error)
	Upsert(ctx context.Context, doc *TheoryContentDoc) error
}
