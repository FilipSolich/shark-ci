package objectstore

import (
	"context"
	"io"
)

type MocObjectStore struct{}

var _ ObjectStorer = &MocObjectStore{}

func NewMocObjectStore() *MocObjectStore {
	return &MocObjectStore{}
}

func (s *MocObjectStore) CreateBucket(ctx context.Context, bucket string) error {
	return nil
}

func (s *MocObjectStore) UploadLogs(ctx context.Context, id int64, logs io.Reader, size int64) error {
	return nil
}

func (s *MocObjectStore) DownloadLogs(ctx context.Context, id int64) (io.ReadCloser, error) {
	return nil, nil
}
