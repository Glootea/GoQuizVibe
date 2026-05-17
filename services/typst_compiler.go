package services

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type TypstCompiler struct{}

func NewTypstCompiler() *TypstCompiler {
	return &TypstCompiler{}
}

func (c *TypstCompiler) CompileTypst(ctx context.Context, typstContent []byte, dependencies map[string][]byte) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "typst-compile-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	mainTypstPath := filepath.Join(tempDir, "main.typ")
	if err := os.WriteFile(mainTypstPath, typstContent, 0644); err != nil {
		return nil, fmt.Errorf("write main.typ: %w", err)
	}

	for name, content := range dependencies {
		depPath := filepath.Join(tempDir, name)
		if err := os.MkdirAll(filepath.Dir(depPath), 0755); err != nil {
			continue
		}
		if err := os.WriteFile(depPath, content, 0644); err != nil {
			continue
		}
	}

	outputPath := filepath.Join(tempDir, "output.svg")

	cmd := exec.CommandContext(ctx, "typst", "compile", mainTypstPath, outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("typst compile: %w, stderr: %s", err, stderr.String())
	}

	svgContent, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	return svgContent, nil
}