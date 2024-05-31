package objectstore

import (
	"context"
	"io"
)

type ObjectStorer interface {
	CreateBucket(ctx context.Context, bucket string) error
	UploadLogs(ctx context.Context, id int64, logs io.Reader, size int64) error
	DownloadLogs(ctx context.Context, id int64) (io.ReadCloser, error)
}
