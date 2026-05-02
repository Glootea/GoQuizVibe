package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/goquizvibe/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
	client *minio.Client
	bucket string
}

func NewStorageService(cfg config.MinioConfig) (*StorageService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
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

	_, err = s.client.PutObject(ctx, s.bucket, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	return s.presignedURL(objectName), nil
}

func (s *StorageService) DeleteImage(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	return nil
}

func (s *StorageService) presignedURL(objectName string) string {
	return fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, objectName)
}

func (s *StorageService) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

func (s *StorageService) GetImageURL(objectName string) string {
	return s.presignedURL(objectName)
}