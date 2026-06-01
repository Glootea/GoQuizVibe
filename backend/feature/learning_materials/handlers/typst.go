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

	materialIDStr := r.PathValue("id")
	if materialIDStr == "" {
		return ce.WithHTTPStatus(errors.New("material id required"), http.StatusBadRequest)
	}

	materialID, err := uuid.Parse(materialIDStr)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	var req struct {
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if req.Source == "" {
		return ce.WithHTTPStatus(errors.New("source required"), http.StatusBadRequest)
	}

	url, err := h.materialService.CompileAndGetURL(r.Context(), materialID, []byte(req.Source))
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return json.NewEncoder(w).Encode(map[string]string{"url": url})
}
