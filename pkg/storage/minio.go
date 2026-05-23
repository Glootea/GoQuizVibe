package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type Storage interface {
	Bucket() string
	Endpoint() string
	PutObject(ctx context.Context, path string, data []byte, contentType string) error
	GetObject(ctx context.Context, path string) ([]byte, error)
	RemoveObject(ctx context.Context, path string) error
	ListObjects(ctx context.Context, prefix string) ([]string, error)
	EnsureBucket(ctx context.Context) error
	GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)
	SetBucketPolicy(ctx context.Context, bucket, policy string) error
}

type MinioClient struct {
	client *minio.Client
	bucket string
}

func NewMinioClient(cfg MinioConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new: %w", err)
	}
	return &MinioClient{client: client, bucket: cfg.Bucket}, nil
}

func (m *MinioClient) Bucket() string {
	return m.bucket
}

func (m *MinioClient) Endpoint() string {
	return m.client.EndpointURL().Host
}

func (m *MinioClient) PutObject(ctx context.Context, path string, data []byte, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, path, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

func (m *MinioClient) GetObject(ctx context.Context, path string) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (m *MinioClient) RemoveObject(ctx context.Context, path string) error {
	return m.client.RemoveObject(ctx, m.bucket, path, minio.RemoveObjectOptions{})
}

func (m *MinioClient) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	objChan := m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	})
	for obj := range objChan {
		if obj.Err == nil {
			keys = append(keys, obj.Key)
		}
	}
	return keys, nil
}

func (m *MinioClient) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("bucket exists: %w", err)
	}
	if !exists {
		err = m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("make bucket: %w", err)
		}
	}
	return nil
}

func (m *MinioClient) SetBucketPolicy(ctx context.Context, bucket, policy string) error {
	return m.client.SetBucketPolicy(ctx, bucket, policy)
}

func (m *MinioClient) GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, path, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get: %w", err)
	}
	return presignedURL.String(), nil
}

const (
	TypstDir    = "typst"
	CompiledDir = "compiled"
	ResourceDir = "resources"
)

func TypstSourcePath(materialID string) string {
	return fmt.Sprintf("%s/%s/main.typ", TypstDir, materialID)
}

func TypstSourceDir(materialID string) string {
	return fmt.Sprintf("%s/%s", TypstDir, materialID)
}

func CompiledPDFPath(materialID string) string {
	return fmt.Sprintf("%s/%s.pdf", CompiledDir, materialID)
}

func ResourcePath(materialID, ext string) string {
	return fmt.Sprintf("%s/%s%s", ResourceDir, materialID, ext)
}