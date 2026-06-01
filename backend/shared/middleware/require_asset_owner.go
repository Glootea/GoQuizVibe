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
	"github.com/goquizvibe/backend/shared/types"
	"github.com/goquizvibe/backend/shared/ui"
)

func NewRequireAssetOwnerMiddleware(
	auth interfaces.Authenticator,
	permissions *permissionsSvc.PermissionsService,
	localeService *locales.Service,
) RequireAssetOwnerMiddleware {
	return RequireAssetOwnerMiddleware{
		auth:          auth,
		permissions:   permissions,
		localeService: localeService,
	}
}

type RequireAssetOwnerMiddleware struct {
	auth          interfaces.Authenticator
	permissions   *permissionsSvc.PermissionsService
	localeService *locales.Service
}

func (m RequireAssetOwnerMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		assetType := r.PathValue("type")
		assetIDStr := r.PathValue("id")
		if assetType == "" || assetIDStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		assetID, err := uuid.Parse(assetIDStr)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		isOwner, err := m.permissions.IsOwner(r.Context(), db.AssetType(assetType), assetID, claims.UserID)
		if err != nil {
			m.renderError(w, r, http.StatusInternalServerError, ce.UserMessage(errors.Join(ce.ErrInternal, err)))
			return
		}
		if !isOwner {
			m.renderError(w, r, http.StatusForbidden, m.translator(r).YouDoNotHavePermissionToManageThisAsset())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m RequireAssetOwnerMiddleware) translator(r *http.Request) locales.Translator {
	if t := GetTranslator(r.Context()); t != nil {
		return t
	}
	return m.localeService.Get(parseAcceptLanguage(r.Header.Get("Accept-Language")))
}

func (m RequireAssetOwnerMiddleware) renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	translator := m.translator(r)
	w.WriteHeader(status)
	if IsHTMXRequest(r) {
		ui.ErrorAlert(message, "/dashboard", translator).Render(r.Context(), w)
		return
	}
	ui.ErrorPage(types.ErrorData{Message: message, RedirectTo: "/dashboard"}, translator).Render(r.Context(), w)
}
