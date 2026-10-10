// Package minio wraps the MinIO Go SDK for private-bucket file storage.
// All objects are stored in a single private bucket; access is via signed URLs only.
package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/config"
)

const signedURLTTL = 1 * time.Hour

// Client wraps MinIO operations needed by the application.
type Client struct {
	mc     *minio.Client
	signer *minio.Client // подписывает ссылки публичным адресом (см. S3Config.PublicEndpoint)
	bucket string
}

// New connects to MinIO and ensures the private bucket exists.
func New(cfg config.S3Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}

	ctx := context.Background()
	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio BucketExists: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio MakeBucket: %w", err)
		}
	}
	// Bucket remains private — no anonymous policy set.

	// Хост входит в подпись ссылки, поэтому подменить его после подписи нельзя —
	// подписываем отдельным клиентом с публичным адресом. Регион задан явно,
	// чтобы подпись не требовала сетевого запроса к публичному адресу.
	signer := mc
	if cfg.PublicEndpoint != "" && cfg.PublicEndpoint != cfg.Endpoint {
		signer, err = minio.New(cfg.PublicEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.UseSSL,
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("minio.New public signer: %w", err)
		}
	}

	return &Client{mc: mc, signer: signer, bucket: cfg.Bucket}, nil
}

// PresignedGetURL returns a signed GET URL valid for signedURLTTL.
// filename — исходное имя: браузер покажет или сохранит файл под ним, а не
// под техническим ключом объекта. Пусто — без подсказки.
func (c *Client) PresignedGetURL(ctx context.Context, objectKey, filename string) (string, error) {
	var params url.Values
	if filename != "" {
		params = url.Values{}
		params.Set("response-content-disposition",
			"inline; filename*=UTF-8''"+url.PathEscape(filename))
	}
	u, err := c.signer.PresignedGetObject(ctx, c.bucket, objectKey, signedURLTTL, params)
	if err != nil {
		return "", fmt.Errorf("minio PresignedGetObject: %w", err)
	}
	return u.String(), nil
}

// PutObject uploads a file from an io.Reader.
func (c *Client) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := c.mc.PutObject(ctx, c.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("minio PutObject: %w", err)
	}
	return nil
}

// GetObject downloads an object and returns its bytes.
func (c *Client) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio GetObject: %w", err)
	}
	defer obj.Close()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("minio ReadAll: %w", err)
	}
	return data, nil
}

// RemoveObject deletes an object.
func (c *Client) RemoveObject(ctx context.Context, objectKey string) error {
	err := c.mc.RemoveObject(ctx, c.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio RemoveObject: %w", err)
	}
	return nil
}
