package objectstore

import (
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioObjectStore struct {
	client *minio.Client
}

var _ ObjectStorer = &MinioObjectStore{}

func NewMinioObjectStore() (*MinioObjectStore, error) {
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minio", "minio123", ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioObjectStore{client}, nil
}

func (s *MinioObjectStore) CreateBucket(ctx context.Context, bucket string) error {
	err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := s.client.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			return nil
		} else {
			return errors.Join(errBucketExists, err)
		}
	}
	return nil
}

func (s *MinioObjectStore) UploadLogs(ctx context.Context, id int64, logs io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, "shark-ci-logs", strconv.FormatInt(id, 10)+".log", logs, size, minio.PutObjectOptions{})
	return err
}

func (s *MinioObjectStore) DownloadLogs(ctx context.Context, id int64) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, "shark-ci-logs", strconv.FormatInt(id, 10)+".log", minio.GetObjectOptions{})
}
