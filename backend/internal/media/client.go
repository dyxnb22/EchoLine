package media

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/echoline/echoline/backend/internal/config"
)

// Client generates presigned upload URLs for object storage.
type Client struct {
	minio  *minio.Client
	bucket string
}

// NewClient creates a MinIO/S3-compatible client.
func NewClient(cfg config.Config) (*Client, error) {
	if cfg.S3Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT not configured")
	}
	endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.S3Endpoint, "https://"), "http://")
	useSSL := strings.HasPrefix(cfg.S3Endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	return &Client{minio: client, bucket: cfg.S3Bucket}, nil
}

// PresignPutURL returns a short-lived upload URL and object key.
func (c *Client) PresignPutURL(ctx context.Context, ownerID uuid.UUID, mimeType string) (string, string, error) {
	objectKey := fmt.Sprintf("uploads/%s/%s", ownerID, uuid.New())
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	u, err := c.minio.PresignedPutObject(ctx, c.bucket, objectKey, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("presign put: %w", err)
	}
	return u.String(), objectKey, nil
}

// PresignGetURL returns a short-lived download URL for an object key.
func (c *Client) PresignGetURL(ctx context.Context, objectKey string) (string, error) {
	u, err := c.minio.PresignedGetObject(ctx, c.bucket, objectKey, 5*time.Minute, nil)
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}
	return u.String(), nil
}

// Bucket returns configured bucket name.
func (c *Client) Bucket() string {
	return c.bucket
}

// ParseEndpointHost extracts host for health checks.
func ParseEndpointHost(raw string) (string, error) {
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		return u.Host, nil
	}
	return raw, nil
}
