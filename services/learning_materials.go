package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/minio/minio-go/v7"
	r "github.com/goquizvibe/repositories"
	"github.com/jackc/pgx/v5/pgtype"
)

type LearningMaterialService struct {
	repo           r.LearningMaterialRepository
	storageService *StorageService
	typstCompiler  *TypstCompiler
}

func NewLearningMaterialService(
	repo r.LearningMaterialRepository,
	storageService *StorageService,
	typstCompiler *TypstCompiler,
) *LearningMaterialService {
	return &LearningMaterialService{
		repo:           repo,
		storageService: storageService,
		typstCompiler:  typstCompiler,
	}
}

const (
	typstDir                = "typst"
	compiledDir             = "compiled"
	resourceDir             = "resources"
)

func (s *LearningMaterialService) UploadTypstMaterial(ctx context.Context, ownerID uuid.UUID, title, description string, files []*multipart.FileHeader) (*db.LearningMaterial, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	materialID := uuid.New()

	var mainTypst []byte
	dependencies := make(map[string][]byte)

	for _, file := range files {
		content, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open file: %w", err)
		}
		fileData, err := io.ReadAll(content)
		content.Close()
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}

		filename := filepath.Base(file.Filename)
		if filename == "main.typ" {
			mainTypst = fileData
		} else {
			dependencies[filename] = fileData
		}
	}

	if mainTypst == nil {
		return nil, fmt.Errorf("main.typ not found")
	}

	sourcePath := fmt.Sprintf("%s/%s", typstDir, materialID.String())

	svgContent, err := s.typstCompiler.CompileTypst(ctx, mainTypst, dependencies)
	if err != nil {
		return nil, fmt.Errorf("compile typst: %w", err)
	}

	var buf bytes.Buffer
	buf.Write(svgContent)
	svgPath := fmt.Sprintf("%s/%s.svg", compiledDir, materialID.String())
	_, err = s.storageService.client.PutObject(ctx, s.storageService.bucket, svgPath, &buf, int64(buf.Len()), minio.PutObjectOptions{
		ContentType: "image/svg+xml",
	})
	if err != nil {
		return nil, fmt.Errorf("upload svg: %w", err)
	}

	for name, data := range dependencies {
		depPath := fmt.Sprintf("%s/%s/%s", typstDir, materialID.String(), name)
		var depBuf bytes.Buffer
		depBuf.Write(data)
		_, _ = s.storageService.client.PutObject(ctx, s.storageService.bucket, depPath, &depBuf, int64(depBuf.Len()), minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		})
	}

	totalSize := int64(len(mainTypst))
	for _, d := range dependencies {
		totalSize += int64(len(d))
	}

	now := time.Now()
	material, err := s.repo.CreateLearningMaterial(ctx, db.CreateLearningMaterialParams{
		ID:              materialID,
		Title:           title,
		Description:     description,
		MaterialType:    db.LearningMaterialTypeTypst,
		OwnerID:         ownerID,
		SourcePath:      sourcePath,
		CompiledSvgPath: svgPath,
		ResourcePath:    "",
		FileSize:        pgtype.Int8{Int64: totalSize, Valid: true},
		MimeType:        "application/typst",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return nil, fmt.Errorf("create material: %w", err)
	}

	return &material, nil
}

func (s *LearningMaterialService) UploadResourceMaterial(ctx context.Context, ownerID uuid.UUID, title, description string, file *multipart.FileHeader) (*db.LearningMaterial, error) {
	materialID := uuid.New()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	resourcePath := fmt.Sprintf("%s/%s%s", resourceDir, materialID.String(), ext)

	content, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer content.Close()

	fileData, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".pdf":
			contentType = "application/pdf"
		default:
			contentType = "application/octet-stream"
		}
	}

	var buf bytes.Buffer
	buf.Write(fileData)
	_, err = s.storageService.client.PutObject(ctx, s.storageService.bucket, resourcePath, &buf, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload resource: %w", err)
	}

	now := time.Now()
	material, err := s.repo.CreateLearningMaterial(ctx, db.CreateLearningMaterialParams{
		ID:              materialID,
		Title:           title,
		Description:     description,
		MaterialType:    db.LearningMaterialTypeResource,
		OwnerID:         ownerID,
		SourcePath:      "",
		CompiledSvgPath: "",
		ResourcePath:    resourcePath,
		FileSize:        pgtype.Int8{Int64: file.Size, Valid: true},
		MimeType:        contentType,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return nil, fmt.Errorf("create material: %w", err)
	}

	return &material, nil
}

func (s *LearningMaterialService) DeleteMaterial(ctx context.Context, materialID, ownerID uuid.UUID) error {
	material, err := s.repo.GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		return fmt.Errorf("get material: %w", err)
	}

	if material.OwnerID != ownerID {
		return fmt.Errorf("unauthorized")
	}

	if material.SourcePath != "" {
		s.deleteFolder(ctx, material.SourcePath)
	}
	if material.CompiledSvgPath != "" {
		s.storageService.client.RemoveObject(ctx, s.storageService.bucket, material.CompiledSvgPath, minio.RemoveObjectOptions{})
	}
	if material.ResourcePath != "" {
		s.storageService.client.RemoveObject(ctx, s.storageService.bucket, material.ResourcePath, minio.RemoveObjectOptions{})
	}

	if err := s.repo.DeleteLearningMaterial(ctx, materialID); err != nil {
		return fmt.Errorf("delete material: %w", err)
	}

	return nil
}

func (s *LearningMaterialService) deleteFolder(ctx context.Context, folderPath string) {
	objChan := s.storageService.client.ListObjects(ctx, s.storageService.bucket, minio.ListObjectsOptions{
		Prefix:    folderPath + "/",
		Recursive: true,
	})
	for obj := range objChan {
		if obj.Err == nil {
			s.storageService.client.RemoveObject(ctx, s.storageService.bucket, obj.Key, minio.RemoveObjectOptions{})
		}
	}
}

func (s *LearningMaterialService) GetAllMaterials(ctx context.Context) ([]db.LearningMaterial, error) {
	materials, err := s.repo.GetLearningMaterials(ctx)
	if err != nil {
		return nil, fmt.Errorf("get materials: %w", err)
	}
	return materials, nil
}

func (s *LearningMaterialService) GetMaterialByID(ctx context.Context, materialID uuid.UUID) (*db.LearningMaterial, error) {
	material, err := s.repo.GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		return nil, fmt.Errorf("get material: %w", err)
	}
	return &material, nil
}

func (s *LearningMaterialService) GetMaterialURL(ctx context.Context, material db.LearningMaterial) (string, error) {
	var objectPath string
	if material.MaterialType == db.LearningMaterialTypeTypst {
		if material.CompiledSvgPath != "" {
			objectPath = material.CompiledSvgPath
		} else if material.SourcePath != "" {
			objectPath = material.SourcePath + "/main.typ"
		}
	} else if material.ResourcePath != "" {
		objectPath = material.ResourcePath
	}

	if objectPath == "" {
		return "", fmt.Errorf("no path available")
	}

	presignedURL, err := s.storageService.client.PresignedGetObject(ctx, s.storageService.bucket, objectPath, 24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *LearningMaterialService) CompileTypst(ctx context.Context, materialID, ownerID uuid.UUID) (*db.LearningMaterial, error) {
	material, err := s.repo.GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		return nil, fmt.Errorf("get material: %w", err)
	}

	if material.OwnerID != ownerID {
		return nil, fmt.Errorf("unauthorized")
	}

	if material.MaterialType != db.LearningMaterialTypeTypst {
		return nil, fmt.Errorf("material is not typst type")
	}

	if material.SourcePath == "" {
		return nil, fmt.Errorf("no source path")
	}

	mainTypst, err := s.getFileFromMinIO(ctx, material.SourcePath+"/main.typ")
	if err != nil {
		return nil, fmt.Errorf("get main.typ: %w", err)
	}

	dependencies := make(map[string][]byte)
	objChan := s.storageService.client.ListObjects(ctx, s.storageService.bucket, minio.ListObjectsOptions{
		Prefix:    material.SourcePath + "/",
		Recursive: false,
	})
	for obj := range objChan {
		if obj.Err == nil && obj.Key != material.SourcePath+"/main.typ" {
			if data, err := s.getFileFromMinIO(ctx, obj.Key); err == nil {
				depName := strings.TrimPrefix(obj.Key, material.SourcePath+"/")
				dependencies[depName] = data
			}
		}
	}

	svgContent, err := s.typstCompiler.CompileTypst(ctx, mainTypst, dependencies)
	if err != nil {
		return nil, fmt.Errorf("compile typst: %w", err)
	}

	svgPath := fmt.Sprintf("%s/%s.svg", compiledDir, materialID.String())
	var buf bytes.Buffer
	buf.Write(svgContent)
	_, err = s.storageService.client.PutObject(ctx, s.storageService.bucket, svgPath, &buf, int64(buf.Len()), minio.PutObjectOptions{
		ContentType: "image/svg+xml",
	})
	if err != nil {
		return nil, fmt.Errorf("upload svg: %w", err)
	}

	now := time.Now()
	updated, err := s.repo.UpdateLearningMaterial(ctx, db.UpdateLearningMaterialParams{
		ID:              materialID,
		Title:           material.Title,
		Description:     material.Description,
		SourcePath:      material.SourcePath,
		CompiledSvgPath: svgPath,
		ResourcePath:    material.ResourcePath,
		FileSize:        material.FileSize,
		MimeType:        material.MimeType,
		UpdatedAt:       now,
	})
	if err != nil {
		return nil, fmt.Errorf("update material: %w", err)
	}

	return &updated, nil
}

func (s *LearningMaterialService) getFileFromMinIO(ctx context.Context, objectPath string) ([]byte, error) {
	obj, err := s.storageService.client.GetObject(ctx, s.storageService.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

type MaterialWithURL struct {
	Material  db.LearningMaterial
	PublicURL string
	Type      string
}