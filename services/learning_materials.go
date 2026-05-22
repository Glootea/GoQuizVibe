package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/internal/grpc/proto"
	"github.com/goquizvibe/pkg/storage"
	r "github.com/goquizvibe/repositories"
	"github.com/jackc/pgx/v5/pgtype"
)

type LearningMaterialService struct {
	repo           r.LearningMaterialRepository
	storageService *StorageService
	typstClient    *TypstGRPCClient
}

func NewLearningMaterialService(
	repo r.LearningMaterialRepository,
	storageService *StorageService,
	typstClient *TypstGRPCClient,
) *LearningMaterialService {
	return &LearningMaterialService{
		repo:           repo,
		storageService: storageService,
		typstClient:    typstClient,
	}
}

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
			mainTypst = make([]byte, len(fileData))
			copy(mainTypst, fileData)
		} else {
			depData := make([]byte, len(fileData))
			copy(depData, fileData)
			dependencies[filename] = depData
		}
	}

	if mainTypst == nil {
		return nil, fmt.Errorf("main.typ not found")
	}

	mainTypstPath := storage.TypstSourcePath(materialID.String())
	if err := s.storageService.client.PutObject(ctx, mainTypstPath, mainTypst, "application/typst"); err != nil {
		return nil, fmt.Errorf("upload main.typ: %w", err)
	}

	for name, data := range dependencies {
		depPath := fmt.Sprintf("%s/%s", storage.TypstSourceDir(materialID.String()), name)
		_ = s.storageService.client.PutObject(ctx, depPath, data, "application/octet-stream")
	}

	resp, err := s.typstClient.Compile(ctx, &proto.CompileRequest{
		MaterialId: materialID.String(),
		SourcePath: mainTypstPath,
	})
	if err != nil {
		return nil, fmt.Errorf("compile typst: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("compile typst failed: %s", resp.Errors)
	}

	pdfPath := resp.PdfPath

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
		SourcePath:      storage.TypstSourceDir(materialID.String()),
		CompiledPath:    pdfPath,
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
	resourcePath := storage.ResourcePath(materialID.String(), ext)

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

	if err := s.storageService.client.PutObject(ctx, resourcePath, fileData, contentType); err != nil {
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
		CompiledPath:    "",
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
	if material.CompiledPath != "" {
		_ = s.storageService.client.RemoveObject(ctx, material.CompiledPath)
	}
	if material.ResourcePath != "" {
		_ = s.storageService.client.RemoveObject(ctx, material.ResourcePath)
	}

	if err := s.repo.DeleteLearningMaterial(ctx, materialID); err != nil {
		return fmt.Errorf("delete material: %w", err)
	}

	return nil
}

func (s *LearningMaterialService) deleteFolder(ctx context.Context, folderPath string) {
	keys, err := s.storageService.client.ListObjects(ctx, folderPath+"/")
	if err != nil {
		return
	}
	for _, key := range keys {
		_ = s.storageService.client.RemoveObject(ctx, key)
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
		if material.CompiledPath != "" {
			objectPath = material.CompiledPath
		} else if material.SourcePath != "" {
			objectPath = material.SourcePath + "/main.typ"
		}
	} else if material.ResourcePath != "" {
		objectPath = material.ResourcePath
	}

	if objectPath == "" {
		return "", fmt.Errorf("no path available")
	}

	return s.storageService.GetPresignedURL(objectPath), nil
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

	resp, err := s.typstClient.Compile(ctx, &proto.CompileRequest{
		MaterialId: materialID.String(),
		SourcePath: material.SourcePath + "/main.typ",
	})
	if err != nil {
		return nil, fmt.Errorf("compile typst: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("compile typst failed: %s", resp.Errors)
	}

	pdfPath := resp.PdfPath

	now := time.Now()
	updated, err := s.repo.UpdateLearningMaterial(ctx, db.UpdateLearningMaterialParams{
		ID:              materialID,
		Title:           material.Title,
		Description:     material.Description,
		SourcePath:      material.SourcePath,
		CompiledPath:    pdfPath,
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

func (s *LearningMaterialService) GetSource(ctx context.Context, materialID uuid.UUID) ([]byte, error) {
	material, err := s.repo.GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		return nil, fmt.Errorf("get material: %w", err)
	}
	return s.storageService.client.GetObject(ctx, material.SourcePath+"/main.typ")
}

func (s *LearningMaterialService) CompileAndGetURL(ctx context.Context, materialID uuid.UUID, source []byte) (string, error) {
	material, err := s.repo.GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		return "", fmt.Errorf("get material: %w", err)
	}

	sourcePath := material.SourcePath + "/main.typ"
	if err := s.storageService.client.PutObject(ctx, sourcePath, source, "application/typst"); err != nil {
		return "", fmt.Errorf("upload source: %w", err)
	}

	resp, err := s.typstClient.Compile(ctx, &proto.CompileRequest{
		MaterialId: materialID.String(),
		SourcePath: sourcePath,
	})
	if err != nil {
		return "", fmt.Errorf("compile: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("compile failed: %s", resp.Errors)
	}

	return s.storageService.GetPresignedURL(resp.PdfPath), nil
}