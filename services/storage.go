package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/goquizvibe/config"
	"github.com/goquizvibe/pkg/storage"
)

type StorageService struct {
	client storage.Storage
	bucket string
}

func NewStorageService(client storage.Storage) *StorageService {
	return &StorageService{
		client: client,
		bucket: client.Bucket(),
	}
}

func NewStorageServiceFromConfig(cfg config.MinioConfig) (*StorageService, error) {
	client, err := storage.NewMinioClient(storage.MinioConfig{
		Endpoint:  cfg.Endpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		Bucket:    cfg.Bucket,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &StorageService{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *StorageService) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if err := s.client.PutObject(ctx, objectName, data, contentType); err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	return s.GetPresignedURL(objectName), nil
}

func (s *StorageService) DeleteImage(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, objectName)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	return nil
}

func (s *StorageService) GetPresignedURL(objectName string) string {
	presignedURL, err := s.client.GetPresignedURL(context.Background(), objectName, 24*time.Hour)
	if err != nil {
		return fmt.Sprintf("http://%s/%s/%s", s.client.Endpoint(), s.bucket, objectName)
	}
	return presignedURL
}

func (s *StorageService) EnsureBucket(ctx context.Context) error {
	if err := s.client.EnsureBucket(ctx); err != nil {
		return fmt.Errorf("failed to ensure bucket: %w", err)
	}

	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, s.bucket)
	if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	return nil
}

func (s *StorageService) GetImageURL(objectName string) string {
	return s.GetPresignedURL(objectName)
}