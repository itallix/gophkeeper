package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/caarlos0/env/v11"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage is the wrapper for S3 compatible client using minio.
type ObjectStorage struct {
	client *minio.Client
}

type config struct {
	AccessKey string `env:"ACCESS_KEY" envDefault:"superadmin"`
	SecretKey string `env:"SECRET_KEY" envDefault:"superpassword"`
	Endpoint  string `env:"ENDPOINT" envDefault:"localhost:9000"`
}

func NewObjectStorage() (*ObjectStorage, error) {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("cannot parse minio env vars %w", err)
	}
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &ObjectStorage{
		client: minioClient,
	}, nil
}

func (s *ObjectStorage) Upload(ctx context.Context, bucket, name string, size int64, reader io.Reader) (int64, error) {
	info, err := s.client.PutObject(
		ctx, bucket, name, reader, size,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (s *ObjectStorage) GetObject(ctx context.Context, bucket, name string) (io.ReadCloser, int64, error) {
	object, err := s.client.GetObject(ctx, bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching object from storage: %w", err)
	}
	stat, err := object.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting metadata from object: %w", err)
	}
	return object, stat.Size, nil
}

func (s *ObjectStorage) DeleteChunks(ctx context.Context, bucket, name string) error {
	objectsCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    name,
		Recursive: true,
	})

	for obj := range objectsCh {
		if err := s.client.RemoveObject(ctx, bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
			return err
		}
	}
	return nil
}
