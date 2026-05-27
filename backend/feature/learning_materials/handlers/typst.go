package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	adminSvc "github.com/goquizvibe/backend/feature/admin/services"
	lmSvc "github.com/goquizvibe/backend/feature/learning_materials/services"
	ce "github.com/goquizvibe/backend/shared/custom_errors"
)

type TypstHandler struct {
	materialService *lmSvc.LearningMaterialService
	adminService    *adminSvc.AdminService
}

func NewTypstHandler(ms *lmSvc.LearningMaterialService, as *adminSvc.AdminService) *TypstHandler {
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
