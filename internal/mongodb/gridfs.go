package mongodb

import (
	"bytes"
	"context"
	"fmt"

	"magua-fly-golang/internal/usecase"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GridFSBucketImpl 是 usecase.GridFSBucket 的 Mongo GridFS 实现。
type GridFSBucketImpl struct {
	bucket *gridfs.Bucket
}

// NewGridFSBucket 创建一个新的 GridFSBucketImpl。
// bucketName 例如 "contracts_files"。
func (c *Client) NewGridFSBucket(bucketName string) (usecase.GridFSBucket, error) {
	opts := options.GridFSBucket().SetName(bucketName)
	b, err := gridfs.NewBucket(c.DB, opts)
	if err != nil {
		return nil, fmt.Errorf("gridfs.NewBucket: %w", err)
	}
	return &GridFSBucketImpl{bucket: b}, nil
}

// UploadFromStream 将内存中的数据写入 GridFS，并返回文件 ID（字符串形式）。
func (g *GridFSBucketImpl) UploadFromStream(ctx context.Context, filename string, data []byte, metadata map[string]interface{}) (string, error) {
	var metaDoc interface{}
	if metadata != nil {
		metaDoc = metadata
	}
	uploadOpts := options.GridFSUpload().SetMetadata(metaDoc)
	fileID, err := g.bucket.UploadFromStream(filename, bytes.NewReader(data), uploadOpts)
	if err != nil {
		return "", err
	}
	// 直接使用默认格式化为字符串
	return fmt.Sprintf("%v", fileID), nil
}

// DownloadToStream 根据文件 ID 读取数据与部分元信息。
func (g *GridFSBucketImpl) DownloadToStream(ctx context.Context, fileID string) ([]byte, map[string]interface{}, error) {
	// 将 string 转为 ObjectID
	var objID interface{}
	if err := bson.UnmarshalExtJSON([]byte(`"`+fileID+`"`), false, &objID); err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	_, err := g.bucket.DownloadToStream(objID, &buf)
	if err != nil {
		return nil, nil, err
	}

	// 读取 files 集合中的元数据（可选）
	meta := make(map[string]interface{})
	filesColl := g.bucket.GetFilesCollection()
	var fDoc bson.M
	if err := filesColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&fDoc); err == nil {
		if m, ok := fDoc["metadata"].(bson.M); ok {
			for k, v := range m {
				meta[k] = v
			}
		}
	}
	return buf.Bytes(), meta, nil
}

// Delete 按 GridFS 文件 ID（ObjectId 字符串）删除文件。
func (g *GridFSBucketImpl) Delete(ctx context.Context, fileID string) error {
	var objID interface{}
	if err := bson.UnmarshalExtJSON([]byte(`"`+fileID+`"`), false, &objID); err != nil {
		return err
	}
	return g.bucket.DeleteContext(ctx, objID)
}

