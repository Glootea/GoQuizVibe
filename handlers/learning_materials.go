package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/db"
	"github.com/goquizvibe/locales"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/pages/admin"
	"github.com/goquizvibe/services"
	"github.com/goquizvibe/types"
)

type LearningMaterialsHandler struct {
	materialService *services.LearningMaterialService
	adminService    *services.AdminService
	localeSvc       *locales.Service
}

func NewLearningMaterialsHandler(
	materialService *services.LearningMaterialService,
	adminService *services.AdminService,
	localeSvc *locales.Service,
) *LearningMaterialsHandler {
	return &LearningMaterialsHandler{
		materialService: materialService,
		adminService:    adminService,
		localeSvc:       localeSvc,
	}
}

func (h *LearningMaterialsHandler) getUser(r *http.Request) (*db.User, error) {
	return h.adminService.GetUserFromRequest(r)
}

func (h *LearningMaterialsHandler) List(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	materials, err := h.materialService.GetAllMaterials(r.Context())
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	materialsWithURLs := make([]types.MaterialWithURL, 0, len(materials))
	for _, m := range materials {
		url, _ := h.materialService.GetMaterialURL(r.Context(), m)
		materialsWithURLs = append(materialsWithURLs, types.MaterialWithURL{
			Material:  m,
			PublicURL: url,
			Type:      string(m.MaterialType),
		})
	}

	t := middleware.GetTranslator(r.Context())
	return admin.LearningMaterialsListPage(user, materialsWithURLs, t).Render(r.Context(), w)
}

func (h *LearningMaterialsHandler) New(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	if r.Method == "POST" {
		title := r.FormValue("title")
		description := r.FormValue("description")
		materialType := r.FormValue("type")

		if title == "" {
			return ce.WithHTTPStatus(errors.New("title is required"), http.StatusBadRequest)
		}

		if materialType == "typst" {
			if err := r.ParseMultipartForm(50 << 20); err != nil {
				return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
			}

			files := r.MultipartForm.File["files"]
			if len(files) == 0 {
				return ce.WithHTTPStatus(errors.New("no files uploaded"), http.StatusBadRequest)
			}

			material, err := h.materialService.UploadTypstMaterial(r.Context(), user.ID, title, description, files)
			if err != nil {
				return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
			}

			if IsHTMXRequest(r) {
				w.Header().Set("HX-Redirect", "/admin/learning-materials/"+material.ID.String())
				w.WriteHeader(http.StatusOK)
				return nil
			}

			http.Redirect(w, r, "/admin/learning-materials/"+material.ID.String(), http.StatusFound)
			return nil
		}

		if materialType == "resource" {
			if err := r.ParseMultipartForm(50 << 20); err != nil {
				return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
			}

			_, fileHeader, err := r.FormFile("file")
			if err != nil {
				return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
			}

			material, err := h.materialService.UploadResourceMaterial(r.Context(), user.ID, title, description, fileHeader)
			if err != nil {
				return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
			}

			if IsHTMXRequest(r) {
				w.Header().Set("HX-Redirect", "/admin/learning-materials/"+material.ID.String())
				w.WriteHeader(http.StatusOK)
				return nil
			}

			http.Redirect(w, r, "/admin/learning-materials/"+material.ID.String(), http.StatusFound)
			return nil
		}

		return ce.WithHTTPStatus(errors.New("invalid material type"), http.StatusBadRequest)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.LearningMaterialsNewPage(user, t).Render(r.Context(), w)
}

func (h *LearningMaterialsHandler) View(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	materialID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	material, err := h.materialService.GetMaterialByID(r.Context(), materialID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	url, _ := h.materialService.GetMaterialURL(r.Context(), *material)

	t := middleware.GetTranslator(r.Context())
	return admin.LearningMaterialsViewPage(user, material, url, t).Render(r.Context(), w)
}

func (h *LearningMaterialsHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	materialID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	err = h.materialService.DeleteMaterial(r.Context(), materialID, user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/admin/learning-materials")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/learning-materials", http.StatusFound)
	return nil
}

func (h *LearningMaterialsHandler) Preview(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	materialID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	material, err := h.materialService.GetMaterialByID(r.Context(), materialID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
	}

	url, err := h.materialService.GetMaterialURL(r.Context(), *material)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return admin.LearningMaterialsPreviewPage(user, material, url, t).Render(r.Context(), w)
}

func (h *LearningMaterialsHandler) Compile(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	materialID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInvalidRequest, err), http.StatusBadRequest)
	}

	material, err := h.materialService.CompileTypst(r.Context(), materialID, user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/learning-materials/"+material.ID.String(), http.StatusFound)
	return nil
}
