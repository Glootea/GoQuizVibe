package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/goquizvibe/microservices/typst/internal/server/proto"
	"github.com/goquizvibe/microservices/typst/internal/service"
	"github.com/goquizvibe/pkg/storage"
)

type Server struct {
	proto.UnimplementedTypstCompilerServer
	compiler *service.Compiler
	storage  storage.Storage
}

func NewServer(compiler *service.Compiler, storage storage.Storage) *Server {
	return &Server{
		compiler: compiler,
		storage:  storage,
	}
}

func (s *Server) Compile(ctx context.Context, req *proto.CompileRequest) (*proto.CompileResponse, error) {
	source, err := s.storage.GetObject(ctx, req.SourcePath)
	if err != nil {
		return &proto.CompileResponse{
			Success: false,
			Errors:  fmt.Sprintf("failed to get source from MinIO: %v", err),
		}, nil
	}

	deps := make(map[string][]byte)
	basePath := req.SourcePath[:strings.LastIndex(req.SourcePath, "/")+1]
	depKeys, err := s.storage.ListObjects(ctx, basePath)
	if err == nil {
		for _, key := range depKeys {
			if key != req.SourcePath && !strings.HasSuffix(key, "/") {
				if data, err := s.storage.GetObject(ctx, key); err == nil {
					depName := strings.TrimPrefix(key, basePath)
					deps[depName] = data
				}
			}
		}
	}

	pdfContent, err := s.compiler.Compile(ctx, source, deps)
	if err != nil {
		return &proto.CompileResponse{
			Success: false,
			Errors:  err.Error(),
		}, nil
	}

	pdfPath := fmt.Sprintf("compiled/%s.pdf", req.MaterialId)
	if err := s.storage.PutObject(ctx, pdfPath, pdfContent, "application/pdf"); err != nil {
		return &proto.CompileResponse{
			Success: false,
			Errors:  fmt.Sprintf("failed to upload PDF: %v", err),
		}, nil
	}

	return &proto.CompileResponse{
		Success:  true,
		PdfPath:  pdfPath,
		Errors:   "",
	}, nil
}