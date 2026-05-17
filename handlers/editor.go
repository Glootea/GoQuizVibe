package handlers

import (
	"net/http"

	"github.com/goquizvibe/db"
	"github.com/goquizvibe/pages/editor"
	"github.com/goquizvibe/services"
	"github.com/google/uuid"
)

type EditorHandler struct {
	materialService *services.LearningMaterialService
}

func NewEditor(ms *services.LearningMaterialService) *EditorHandler {
	return &EditorHandler{materialService: ms}
}

func (h *EditorHandler) EditorPage(w http.ResponseWriter, r *http.Request) error {
	materialIDStr := r.URL.Query().Get("material_id")

	var initialSource string
	var materialID string
	var pdfURL string

	if materialIDStr != "" {
		id, err := uuid.Parse(materialIDStr)
		if err == nil {
			materialID = materialIDStr
			source, err := h.materialService.GetSource(r.Context(), id)
			if err == nil {
				initialSource = string(source)
			}

			material := db.LearningMaterial{
				ID:           id,
				SourcePath:    "typst/" + materialIDStr,
				CompiledPath:  "compiled/" + materialIDStr + ".pdf",
				MaterialType: db.LearningMaterialTypeTypst,
			}
			url, err := h.materialService.GetMaterialURL(r.Context(), material)
			if err == nil {
				pdfURL = url
			}
		}
	}

	return editor.EditorPage(materialID, initialSource, pdfURL).Render(r.Context(), w)
}