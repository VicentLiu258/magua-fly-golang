package mongodb

import (
	"context"
	"fmt"
	"io"

	"magua-fly-golang/internal/usecase"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TheoryVideoGridFS 使用 GridFS bucket 存储理论课视频（流式上传、Range 下载）。
type TheoryVideoGridFS struct {
	bucket *gridfs.Bucket
}

// NewTheoryVideoGridFSBucket 创建理论视频 GridFS 存储（例如 bucketName=theory_videos）。
func (c *Client) NewTheoryVideoGridFSBucket(bucketName string) (usecase.TheoryVideoStorage, error) {
	opts := options.GridFSBucket().SetName(bucketName)
	b, err := gridfs.NewBucket(c.DB, opts)
	if err != nil {
		return nil, fmt.Errorf("gridfs.NewBucket %q: %w", bucketName, err)
	}
	return &TheoryVideoGridFS{bucket: b}, nil
}

var _ usecase.TheoryVideoStorage = (*TheoryVideoGridFS)(nil)

func parseGridFSFileID(fileID string) (interface{}, error) {
	var objID interface{}
	if err := bson.UnmarshalExtJSON([]byte(`"`+fileID+`"`), false, &objID); err != nil {
		return nil, err
	}
	return objID, nil
}

func (g *TheoryVideoGridFS) UploadStream(ctx context.Context, filename, contentType string, metadata map[string]interface{}, src io.Reader) (string, int64, error) {
	meta := bson.M{"contentType": contentType}
	for k, v := range metadata {
		meta[k] = v
	}
	uploadOpts := options.GridFSUpload().SetMetadata(meta)
	us, err := g.bucket.OpenUploadStream(filename, uploadOpts)
	if err != nil {
		return "", 0, err
	}
	n, copyErr := io.Copy(us, src)
	if copyErr != nil {
		_ = us.Close()
		return "", 0, copyErr
	}
	if err := us.Close(); err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%v", us.FileID), n, nil
}

func (g *TheoryVideoGridFS) Delete(ctx context.Context, objectID string) error {
	obj, err := parseGridFSFileID(objectID)
	if err != nil {
		return err
	}
	return g.bucket.DeleteContext(ctx, obj)
}

func (g *TheoryVideoGridFS) FileLength(ctx context.Context, objectID string) (int64, error) {
	obj, err := parseGridFSFileID(objectID)
	if err != nil {
		return 0, err
	}
	ds, err := g.bucket.OpenDownloadStream(obj)
	if err != nil {
		return 0, err
	}
	defer ds.Close()
	return ds.GetFile().Length, nil
}

type readCloser struct {
	r io.Reader
	c io.Closer
}

func (rc *readCloser) Read(p []byte) (int, error) { return rc.r.Read(p) }
func (rc *readCloser) Close() error               { return rc.c.Close() }

func (g *TheoryVideoGridFS) OpenPartial(ctx context.Context, objectID string, offset, maxLen int64) (io.ReadCloser, int64, error) {
	obj, err := parseGridFSFileID(objectID)
	if err != nil {
		return nil, 0, err
	}
	ds, err := g.bucket.OpenDownloadStream(obj)
	if err != nil {
		return nil, 0, err
	}
	total := ds.GetFile().Length
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		_ = ds.Close()
		return nil, total, fmt.Errorf("offset past end of file")
	}
	if skipped, err := ds.Skip(offset); err != nil || skipped != offset {
		_ = ds.Close()
		if err != nil {
			return nil, total, err
		}
		return nil, total, fmt.Errorf("gridfs Skip: want %d got %d", offset, skipped)
	}
	remain := total - offset
	readLen := remain
	if maxLen > 0 && maxLen < remain {
		readLen = maxLen
	}
	return &readCloser{r: io.LimitReader(ds, readLen), c: ds}, total, nil
}
