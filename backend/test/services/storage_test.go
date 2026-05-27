package services_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mime/multipart"

	"github.com/goquizvibe/backend/shared/config"
	storage "github.com/goquizvibe/backend/shared/infrastructure/storage"
	servicestest "github.com/goquizvibe/backend/shared/mocks/servicestest"
	"go.uber.org/mock/gomock"
)

func TestStorageService_UploadImage(t *testing.T) {
	ctx := context.Background()

	t.Run("successful upload", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		header := createTestFileHeaderForStorage(t, "test.jpg", "image/jpeg")
		mockClient.EXPECT().PutObject(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockClient.EXPECT().GetPresignedURL(ctx, gomock.Any(), 24*time.Hour).Return("http://localhost:9000/bucket/object", nil)

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
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		header := createTestFileHeaderForStorage(t, "test.jpg", "image/jpeg")
		mockClient.EXPECT().PutObject(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("upload failed"))

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
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		header := createTestFileHeaderForStorage(t, "test.jpg", "")
		mockClient.EXPECT().PutObject(ctx, gomock.Any(), gomock.Any(), "application/octet-stream").Return(nil)
		mockClient.EXPECT().GetPresignedURL(ctx, gomock.Any(), 24*time.Hour).Return("http://localhost:9000/bucket/object", nil)

		_, err := svc.UploadImage(ctx, header)
		if err != nil {
			t.Fatalf("UploadImage() error = %v, want nil", err)
		}
	})
}

func TestStorageService_DeleteImage(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().RemoveObject(ctx, "object-name").Return(nil)

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
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().RemoveObject(ctx, "object-name").Return(errors.New("delete failed"))

		err := svc.DeleteImage(ctx, "object-name")
		if err == nil {
			t.Fatal("DeleteImage() error = nil, want error")
		}
	})
}

func TestStorageService_GetPresignedURL(t *testing.T) {
	t.Run("presigned URL success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().GetPresignedURL(context.Background(), "object-name", 24*time.Hour).Return("http://localhost:9000/bucket/object-name", nil)

		result := svc.GetPresignedURL("object-name")
		if result == "" {
			t.Error("GetPresignedURL() returned empty string")
		}
	})

	t.Run("presigned URL error fallback", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().GetPresignedURL(context.Background(), "object-name", 24*time.Hour).Return("", errors.New("presign error"))
		mockClient.EXPECT().Endpoint().Return("localhost:9000")

		result := svc.GetPresignedURL("object-name")
		expected := "http://localhost:9000/test-bucket/object-name"
		if result != expected {
			t.Errorf("GetPresignedURL() = %v, want %v", result, expected)
		}
	})
}

func TestStorageService_EnsureBucket(t *testing.T) {
	ctx := context.Background()

	t.Run("bucket already exists", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().EnsureBucket(ctx).Return(nil)
		mockClient.EXPECT().SetBucketPolicy(ctx, "test-bucket", gomock.Any()).Return(nil)

		err := svc.EnsureBucket(ctx)
		if err != nil {
			t.Fatalf("EnsureBucket() error = %v, want nil", err)
		}
	})

	t.Run("EnsureBucket error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := servicestest.NewMockMinioClient(ctrl)
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().EnsureBucket(ctx).Return(errors.New("bucket check failed"))

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
		mockClient.EXPECT().Bucket().Return("test-bucket")
		svc := storage.NewStorageService(mockClient)

		mockClient.EXPECT().EnsureBucket(ctx).Return(nil)
		mockClient.EXPECT().SetBucketPolicy(ctx, "test-bucket", gomock.Any()).Return(errors.New("policy failed"))

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

		_, err := storage.NewStorageServiceFromConfig(cfg)
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

		svc, err := storage.NewStorageServiceFromConfig(cfg)
		if err != nil {
			t.Fatalf("NewStorageServiceFromConfig() error = %v, want nil", err)
		}
		if svc == nil {
			t.Fatal("NewStorageServiceFromConfig() returned nil")
		}
	})
}
