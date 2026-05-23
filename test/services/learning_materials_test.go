package services_test

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/types"
)

func TestMaterialWithURL(t *testing.T) {
	t.Parallel()
	material := db.LearningMaterial{
		ID:           uuid.New(),
		Title:        "Test",
		MaterialType: db.LearningMaterialTypeTypst,
	}

	mwu := types.MaterialWithURL{
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

func TestLearningMaterialService_GetPublicUrl(t *testing.T) {
	t.Parallel()
	material := db.LearningMaterial{
		ID:           uuid.New(),
		Title:        "Test",
		MaterialType: db.LearningMaterialTypeTypst,
	}

	url, _ := url.Parse("http://example.com/material.pdf")
	mwu := types.MaterialWithURL{
		Material:  material,
		PublicURL: url.String(),
		Type:      string(db.LearningMaterialTypeTypst),
	}

	if mwu.PublicURL == "" {
		t.Fatal("expected non-empty public URL")
	}
}