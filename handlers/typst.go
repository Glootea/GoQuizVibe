package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/services"
)

type TypstHandler struct {
	materialService *services.LearningMaterialService
	adminService    *services.AdminService
}

func NewTypstHandler(ms *services.LearningMaterialService, as *services.AdminService) *TypstHandler {
	return &TypstHandler{
		materialService: ms,
		adminService:    as,
	}
}

func (h *TypstHandler) Compile(w http.ResponseWriter, r *http.Request) error {
	_, err := h.adminService.GetUserFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	var req struct {
		MaterialID string `json:"material_id"`
		Source     string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if req.MaterialID == "" || req.Source == "" {
		return ce.WithHTTPStatus(errors.New("material_id and source required"), http.StatusBadRequest)
	}

	materialID, err := uuid.Parse(req.MaterialID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	url, err := h.materialService.CompileAndGetURL(r.Context(), materialID, []byte(req.Source))
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return json.NewEncoder(w).Encode(map[string]string{"url": url})
}