package usecase

import (
	"context"
	"io"
)

// TheoryVideoStorage 抽象理论课视频对象存储（GridFS / 后续 COS），便于切换实现。
type TheoryVideoStorage interface {
	UploadStream(ctx context.Context, filename, contentType string, metadata map[string]interface{}, src io.Reader) (objectID string, size int64, err error)
	Delete(ctx context.Context, objectID string) error
	FileLength(ctx context.Context, objectID string) (int64, error)
	// OpenPartial 读取 [offset, offset+maxLen) 与文件末尾的交集；maxLen<=0 表示读到文件末尾。
	OpenPartial(ctx context.Context, objectID string, offset, maxLen int64) (rc io.ReadCloser, total int64, err error)
}
