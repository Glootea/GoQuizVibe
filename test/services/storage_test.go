package services_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"mime/multipart"

	"github.com/goquizvibe/config"
	servicestest "github.com/goquizvibe/mocks/servicestest"
	"github.com/goquizvibe/services"
	"github.com/minio/minio-go/v7"
	"go.uber.org/mock/gomock"
)

func TestStorageService_UploadImage(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("successful upload", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		header := createTestFileHeaderForStorage(t, "test.jpg", "image/jpeg")
		mockClient.EXPECT().PutObject(ctx, bucket, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, nil)
		mockClient.EXPECT().PresignedGetObject(ctx, bucket, gomock.Any(), 24*time.Hour, nil).Return(&url.URL{Scheme: "http", Host: "localhost:9000", Path: "/bucket/object"}, nil)

		result, err := svc.UploadImage(ctx, header)
		if err != nil {
			t.Fatalf("UploadImage() error = %v, want nil", err)
		}
		if result == "" {
			t.Error("UploadImage() returned empty URL")
		}
	})

	t.Run("PutObject error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		header := createTestFileHeaderForStorage(t, "test.jpg", "image/jpeg")
		mockClient.EXPECT().PutObject(ctx, bucket, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, errors.New("upload failed"))

		_, err := svc.UploadImage(ctx, header)
		if err == nil {
			t.Fatal("UploadImage() error = nil, want error")
		}
	})

	t.Run("default content type", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		header := createTestFileHeaderForStorage(t, "test.jpg", "")
		mockClient.EXPECT().PutObject(ctx, bucket, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(minio.UploadInfo{}, nil)
		mockClient.EXPECT().PresignedGetObject(ctx, bucket, gomock.Any(), 24*time.Hour, nil).Return(&url.URL{Scheme: "http", Host: "localhost:9000"}, nil)

		_, err := svc.UploadImage(ctx, header)
		if err != nil {
			t.Fatalf("UploadImage() error = %v, want nil", err)
		}
	})
}

func TestStorageService_DeleteImage(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("successful delete", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().RemoveObject(ctx, bucket, "object-name", gomock.Any()).Return(nil)

		err := svc.DeleteImage(ctx, "object-name")
		if err != nil {
			t.Fatalf("DeleteImage() error = %v, want nil", err)
		}
	})

	t.Run("RemoveObject error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().RemoveObject(ctx, bucket, "object-name", gomock.Any()).Return(errors.New("delete failed"))

		err := svc.DeleteImage(ctx, "object-name")
		if err == nil {
			t.Fatal("DeleteImage() error = nil, want error")
		}
	})
}

func TestStorageService_GetImageURL(t *testing.T) {
	bucket := "test-bucket"

	t.Run("presigned URL success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().PresignedGetObject(context.Background(), bucket, "object-name", 24*time.Hour, nil).Return(&url.URL{Scheme: "http", Host: "localhost:9000", Path: "/bucket/object"}, nil)

		result := svc.GetImageURL("object-name")
		if result == "" {
			t.Error("GetImageURL() returned empty string")
		}
		if !strings.HasPrefix(result, "http") {
			t.Errorf("GetImageURL() = %v, want URL", result)
		}
	})

	t.Run("presigned URL error fallback", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().PresignedGetObject(context.Background(), bucket, "object-name", 24*time.Hour, nil).Return(nil, errors.New("presign error"))
		mockClient.EXPECT().EndpointURL().Return(&url.URL{Host: "localhost:9000"})

		result := svc.GetImageURL("object-name")
		expected := "http://localhost:9000/test-bucket/object-name"
		if result != expected {
			t.Errorf("GetImageURL() = %v, want %v", result, expected)
		}
	})
}

func TestStorageService_EnsureBucket(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"

	t.Run("bucket already exists", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().BucketExists(ctx, bucket).Return(true, nil)
		mockClient.EXPECT().SetBucketPolicy(ctx, bucket, gomock.Any()).Return(nil)

		err := svc.EnsureBucket(ctx)
		if err != nil {
			t.Fatalf("EnsureBucket() error = %v, want nil", err)
		}
	})

	t.Run("create bucket and set policy", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().BucketExists(ctx, bucket).Return(false, nil)
		mockClient.EXPECT().MakeBucket(ctx, bucket, gomock.Any()).Return(nil)
		mockClient.EXPECT().SetBucketPolicy(ctx, bucket, gomock.Any()).Return(nil)

		err := svc.EnsureBucket(ctx)
		if err != nil {
			t.Fatalf("EnsureBucket() error = %v, want nil", err)
		}
	})

	t.Run("BucketExists error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().BucketExists(ctx, bucket).Return(false, errors.New("bucket check failed"))

		err := svc.EnsureBucket(ctx)
		if err == nil {
			t.Fatal("EnsureBucket() error = nil, want error")
		}
	})

	t.Run("MakeBucket error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().BucketExists(ctx, bucket).Return(false, nil)
		mockClient.EXPECT().MakeBucket(ctx, bucket, gomock.Any()).Return(errors.New("make bucket failed"))

		err := svc.EnsureBucket(ctx)
		if err == nil {
			t.Fatal("EnsureBucket() error = nil, want error")
		}
	})

	t.Run("SetBucketPolicy error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		svc := services.NewStorageService(mockClient, bucket)

		mockClient.EXPECT().BucketExists(ctx, bucket).Return(true, nil)
		mockClient.EXPECT().SetBucketPolicy(ctx, bucket, gomock.Any()).Return(errors.New("policy failed"))

		err := svc.EnsureBucket(ctx)
		if err == nil {
			t.Fatal("EnsureBucket() error = nil, want error")
		}
	})
}

func createTestFileHeaderForStorage(t *testing.T, filename, contentType string) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write([]byte("test content")); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := req.ParseMultipartForm(1024); err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}

	form := req.MultipartForm
	if form == nil || len(form.File["file"]) == 0 {
		t.Fatalf("failed to get file from form")
	}

	header := form.File["file"][0]
	if contentType != "" {
		header.Header.Set("Content-Type", contentType)
	}
	return header
}

func TestStorageService_NewStorageServiceFromConfig(t *testing.T) {
	t.Run("empty endpoint", func(t *testing.T) {
		t.Parallel()
		cfg := config.MinioConfig{
			Endpoint:  "",
			AccessKey: "testkey",
			SecretKey: "testsecret",
			Bucket:    "test-bucket",
		}

		_, err := services.NewStorageServiceFromConfig(cfg)
		if err == nil {
			t.Fatal("NewStorageServiceFromConfig() error = nil, want error")
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := config.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "testkey",
			SecretKey: "testsecret",
			Bucket:    "test-bucket",
		}

		svc, err := services.NewStorageServiceFromConfig(cfg)
		if err != nil {
			t.Fatalf("NewStorageServiceFromConfig() error = %v, want nil", err)
		}
		if svc == nil {
			t.Fatal("NewStorageServiceFromConfig() returned nil")
		}
	})
}
