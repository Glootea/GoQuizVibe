package services_test

import (
	"context"
	"testing"

	"github.com/goquizvibe/internal/grpc/proto"
)

func TestCompileRequest_Fields(t *testing.T) {
	t.Parallel()

	t.Run("material id and source path", func(t *testing.T) {
		t.Parallel()
		req := &proto.CompileRequest{
			MaterialId: "test-material-123",
			SourcePath: "typst/material-123/main.typ",
		}

		if req.GetMaterialId() != "test-material-123" {
			t.Errorf("MaterialId = %s, want 'test-material-123'", req.GetMaterialId())
		}
		if req.GetSourcePath() != "typst/material-123/main.typ" {
			t.Errorf("SourcePath = %s, want 'typst/material-123/main.typ'", req.GetSourcePath())
		}
	})
}

func TestCompileResponse_Success(t *testing.T) {
	t.Parallel()

	t.Run("successful compilation", func(t *testing.T) {
		t.Parallel()
		resp := &proto.CompileResponse{
			PdfPath: "/compiled/material-123.pdf",
			Errors:  "",
			Success: true,
		}

		if resp.GetPdfPath() != "/compiled/material-123.pdf" {
			t.Errorf("PdfPath = %s, want '/compiled/material-123.pdf'", resp.GetPdfPath())
		}
		if resp.GetErrors() != "" {
			t.Errorf("Errors = %s, want empty", resp.GetErrors())
		}
		if !resp.GetSuccess() {
			t.Error("Success should be true")
		}
	})
}

func TestCompileResponse_Error(t *testing.T) {
	t.Parallel()

	t.Run("compilation failed", func(t *testing.T) {
		t.Parallel()
		resp := &proto.CompileResponse{
			PdfPath: "",
			Errors:  "error: #set not found\n  --> test.typ:1:1\n",
			Success: false,
		}

		if resp.GetPdfPath() != "" {
			t.Error("PdfPath should be empty on error")
		}
		if resp.GetErrors() == "" {
			t.Error("Errors should be populated on failure")
		}
		if resp.GetSuccess() {
			t.Error("Success should be false")
		}
	})
}

func TestCompileResponse_EmptyPaths(t *testing.T) {
	t.Parallel()

	t.Run("empty response fields", func(t *testing.T) {
		t.Parallel()
		resp := &proto.CompileResponse{}

		if resp.GetPdfPath() != "" {
			t.Error("Default PdfPath should be empty")
		}
		if resp.GetErrors() != "" {
			t.Error("Default Errors should be empty")
		}
		if resp.GetSuccess() != false {
			t.Error("Default Success should be false")
		}
	})
}

func TestTypstGRPCClient_Interface(t *testing.T) {
	t.Parallel()

	t.Run("typst compiler client interface", func(t *testing.T) {
		t.Parallel()
		var client proto.TypstCompilerClient

		if client != nil {
			t.Error("client should be nil by default")
		}
	})

	t.Run("nil context handling", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		if ctx == nil {
			t.Error("context should not be nil")
		}
	})
}