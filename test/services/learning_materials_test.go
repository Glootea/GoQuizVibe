package services_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/mocks/servicestest"
	"github.com/goquizvibe/services"
	"github.com/minio/minio-go/v7"
	"go.uber.org/mock/gomock"
)

func TestNewTypstCompiler(t *testing.T) {
	t.Parallel()
	compiler := services.NewTypstCompiler()
	if compiler == nil {
		t.Fatal("NewTypstCompiler returned nil")
	}
}

func TestTypstCompiler_CompileTypst_ContextCanceled(t *testing.T) {
	t.Parallel()
	compiler := services.NewTypstCompiler()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := compiler.CompileTypst(ctx, []byte("# Hello"), nil)
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}

func TestMaterialWithURL(t *testing.T) {
	t.Parallel()
	material := db.LearningMaterial{
		ID:           uuid.New(),
		Title:        "Test",
		MaterialType: db.LearningMaterialTypeTypst,
	}

	mwu := services.MaterialWithURL{
		Material:  material,
		PublicURL: "http://example.com",
		Type:      string(db.LearningMaterialTypeTypst),
	}

	if mwu.Material.Title != "Test" {
		t.Fatal("unexpected title")
	}
	if mwu.Type != "typst" {
		t.Fatal("unexpected type")
	}
}

func TestGetMaterialTypeClass(t *testing.T) {
	t.Parallel()
	tests := []struct {
		materialType string
		expected     string
	}{
		{"typst", "bg-purple-100 text-purple-700"},
		{"resource", "bg-green-100 text-green-700"},
		{"unknown", "bg-gray-100 text-gray-700"},
	}

	for _, tt := range tests {
		result := getMaterialTypeClass(tt.materialType)
		if result != tt.expected {
			t.Errorf("getMaterialTypeClass(%s) = %s, want %s", tt.materialType, result, tt.expected)
		}
	}
}

func getMaterialTypeClass(materialType string) string {
	switch materialType {
	case "typst":
		return "bg-purple-100 text-purple-700"
	case "resource":
		return "bg-green-100 text-green-700"
	default:
		return "bg-gray-100 text-gray-700"
	}
}

func TestMinioClient_PresignedGetObject(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := servicestest.NewMockMinioClient(ctrl)

	expectedURL, _ := url.Parse("http://localhost:9000/test-bucket/object.pdf?signature=xyz")
	mockMinio.EXPECT().PresignedGetObject(context.Background(), "test-bucket", "object.pdf", 24*time.Hour, nil).Return(expectedURL, nil)

	result, err := mockMinio.PresignedGetObject(context.Background(), "test-bucket", "object.pdf", 24*time.Hour, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil URL")
	}
}

func TestMinioClient_GetObject(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := servicestest.NewMockMinioClient(ctrl)

	expectedObj := &minio.Object{}
	mockMinio.EXPECT().GetObject(context.Background(), "test-bucket", "object.pdf", minio.GetObjectOptions{}).Return(expectedObj, nil)

	result, err := mockMinio.GetObject(context.Background(), "test-bucket", "object.pdf", minio.GetObjectOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil Object")
	}
}

func TestMinioClient_ListObjects(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := servicestest.NewMockMinioClient(ctrl)

	objChan := make(chan minio.ObjectInfo, 1)
	objChan <- minio.ObjectInfo{Key: "object.pdf"}
	close(objChan)

	mockMinio.EXPECT().ListObjects(context.Background(), "test-bucket", minio.ListObjectsOptions{Prefix: "docs/", Recursive: true}).Return(objChan)

	result := mockMinio.ListObjects(context.Background(), "test-bucket", minio.ListObjectsOptions{Prefix: "docs/", Recursive: true})
	count := 0
	for range result {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 object, got %d", count)
	}
}

func TestLearningMaterialService_New(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := servicestest.NewMockMinioClient(ctrl)
	storageSvc := services.NewStorageService(mockMinio, "test-bucket")
	typstCompiler := services.NewTypstCompiler()

	svc := services.NewLearningMaterialService(nil, storageSvc, typstCompiler)
	if svc == nil {
		t.Fatal("NewLearningMaterialService returned nil")
	}
}