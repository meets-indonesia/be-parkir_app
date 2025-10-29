//go:build !minio
// +build !minio

package storage

import (
	"be-parkir/internal/config"
	"context"
	"fmt"
	"io"
)

type MinIOClient struct{}

func NewMinIOClient(cfg config.MinIOConfig) (*MinIOClient, error) { return &MinIOClient{}, nil }

func (m *MinIOClient) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// For stub implementation, return a URL that would work with actual MinIO
	// In production, this generates a presigned URL
	return fmt.Sprintf("http://localhost:9000/be-parkir/%s?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...&X-Amz-Date=...&X-Amz-Expires=604800&X-Amz-SignedHeaders=host", objectName), nil
}

func (m *MinIOClient) PresignedGetURL(ctx context.Context, objectName string, expiresSeconds int64) (string, error) {
	return "", nil
}

// GetObject stub implementation
func (m *MinIOClient) GetObject(ctx context.Context, objectName string) (io.Reader, int64, error) {
	return nil, 0, fmt.Errorf("MinIO stub: file not found. Build with -tags=minio to enable")
}
