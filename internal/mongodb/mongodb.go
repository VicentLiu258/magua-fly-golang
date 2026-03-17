package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps MongoDB client and database name for DI.
type Client struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// New creates a new MongoDB client and pings the server.
func New(ctx context.Context, uri, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping: %w", err)
	}

	name := dbName
	if name == "" {
		name = "maguafly"
	}
	return &Client{Client: client, DB: client.Database(name)}, nil
}

// Close disconnects the MongoDB client.
func (c *Client) Close(ctx context.Context) error {
	if c.Client != nil {
		return c.Client.Disconnect(ctx)
	}
	return nil
}

// Health pings the server.
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Client.Ping(ctx, nil)
}

// Collection names used by the app.
const (
	CollTheoryContent        = "theory_content"
	CollSubjectContent       = "subject_content"
	CollComments             = "comments"
	CollFlightImportRaw      = "flight_import_raw"
	CollDashboardSnap        = "dashboard_snapshots"
	CollTheoryQuestionSets   = "theory_question_sets"
	CollTheoryAttemptDetails = "theory_attempt_details"
	// CollContractFilesMeta 存储合同文档元数据（关联 GridFS 文件）
	CollContractFilesMeta = "contract_files_meta"
)
