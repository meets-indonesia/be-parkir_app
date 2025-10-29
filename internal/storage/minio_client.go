//go:build minio
// +build minio

package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"be-parkir/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client *minio.Client
	cfg    config.MinIOConfig
}

func NewMinIOClient(cfg config.MinIOConfig) (*MinIOClient, error) {
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	// Ensure bucket exists
	ctx := context.Background()
	exists, err := cli.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := cli.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	return &MinIOClient{client: cli, cfg: cfg}, nil
}

func (m *MinIOClient) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.cfg.Bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	// Generate presigned URL for direct download (valid for 7 days)
	downloadURL, err := m.PresignedGetURL(ctx, objectName, 7*24*time.Hour)
	if err != nil {
		return "", err
	}
	return downloadURL, nil
}

func (m *MinIOClient) PresignedGetURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, m.cfg.Bucket, objectName, expires, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// GetObject downloads object from MinIO
func (m *MinIOClient) GetObject(ctx context.Context, objectName string) (io.Reader, int64, error) {
	obj, err := m.client.GetObject(ctx, m.cfg.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}
	stat, err := obj.Stat()
	if err != nil {
		return nil, 0, err
	}
	return obj, stat.Size, nil
}
