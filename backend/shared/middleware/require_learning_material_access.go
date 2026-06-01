package middleware

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	permissionsSvc "github.com/goquizvibe/backend/feature/permissions/services"
	ce "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	"github.com/goquizvibe/backend/shared/locales"
	r "github.com/goquizvibe/backend/shared/repositories"
	"github.com/goquizvibe/backend/shared/types"
	"github.com/goquizvibe/backend/shared/ui"
)

var protectedLearningMaterialPatterns = map[string]struct{}{
	"/admin/learning-materials/{id}":         {},
	"/admin/learning-materials/{id}/preview": {},
	"/admin/learning-materials/{id}/compile": {},
	"/editor/{id}":                           {},
	"/api/typst/compile/{id}":                {},
}

func NewRequireLearningMaterialAccessMiddleware(
	auth interfaces.Authenticator,
	permissions *permissionsSvc.PermissionsService,
	materials r.LearningMaterialRepository,
	localeService *locales.Service,
) RequireLearningMaterialAccessMiddleware {
	return RequireLearningMaterialAccessMiddleware{
		auth:          auth,
		permissions:   permissions,
		materials:     materials,
		localeService: localeService,
	}
}

type RequireLearningMaterialAccessMiddleware struct {
	auth          interfaces.Authenticator
	permissions   *permissionsSvc.PermissionsService
	materials     r.LearningMaterialRepository
	localeService *locales.Service
}

func (RequireLearningMaterialAccessMiddleware) IsProtectedPattern(pattern string) bool {
	_, ok := protectedLearningMaterialPatterns[pattern]
	return ok
}

func requiredLevelForMethod(method string) db.PermissionType {
	switch method {
	case http.MethodGet, http.MethodHead:
		return db.PermissionTypeRead
	default:
		return db.PermissionTypeWrite
	}
}

// RequiredLevelForMethod returns the minimum permission level required for
// a given HTTP method. Safe defaults: GET/HEAD → Read, anything else → Write.
func RequiredLevelForMethod(method string) db.PermissionType {
	return requiredLevelForMethod(method)
}

func (m RequireLearningMaterialAccessMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.IsProtectedPattern(r.Pattern) {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(cookieNameToken)
		if err != nil {
			HandleAuthFailure(w, r)
			return
		}
		claims, err := m.auth.ValidateToken(cookie.Value)
		if err != nil {
			HandleAuthFailure(w, r)
			return
		}

		materialIDStr := r.PathValue("id")
		if materialIDStr == "" {
			m.renderError(w, r, http.StatusBadRequest, ce.UserMessage(ce.ErrInvalidRequest))
			return
		}
		materialID, err := uuid.Parse(materialIDStr)
		if err != nil {
			m.renderError(w, r, http.StatusBadRequest, ce.UserMessage(errors.Join(ce.ErrInvalidRequest, err)))
			return
		}

		if _, err := m.materials.GetLearningMaterialByID(r.Context(), materialID); err != nil {
			m.renderError(w, r, http.StatusNotFound, ce.UserMessage(errors.Join(ce.ErrNotFound, err)))
			return
		}

		hasAccess, err := m.permissions.CanAccess(r.Context(), db.AssetTypeLearningMaterial, materialID, claims.UserID, requiredLevelForMethod(r.Method))
		if err != nil {
			m.renderError(w, r, http.StatusInternalServerError, ce.UserMessage(errors.Join(ce.ErrInternal, err)))
			return
		}
		if !hasAccess {
			m.renderError(w, r, http.StatusForbidden, m.translator(r).YouDoNotHavePermissionToAccessThisLearningMaterial())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m RequireLearningMaterialAccessMiddleware) translator(r *http.Request) locales.Translator {
	if t := GetTranslator(r.Context()); t != nil {
		return t
	}
	return m.localeService.Get(parseAcceptLanguage(r.Header.Get("Accept-Language")))
}

func (m RequireLearningMaterialAccessMiddleware) renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	translator := m.translator(r)
	w.WriteHeader(status)
	if IsHTMXRequest(r) {
		ui.ErrorAlert(message, "/admin/learning-materials", translator).Render(r.Context(), w)
		return
	}
	ui.ErrorPage(types.ErrorData{Message: message, RedirectTo: "/admin/learning-materials"}, translator).Render(r.Context(), w)
}
