package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	authServices "github.com/goquizvibe/backend/feature/auth/services"
	"github.com/goquizvibe/backend/feature/permissions/services"
	permissionsUI "github.com/goquizvibe/backend/feature/permissions/ui"
	ce "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	"github.com/goquizvibe/backend/shared/middleware"
)

type PermissionsHandler struct {
	permService  *services.PermissionsService
	groupService *services.UserGroupService
	auth         interfaces.Authenticator
}

func NewPermissionsHandler(permService *services.PermissionsService, groupService *services.UserGroupService, auth interfaces.Authenticator) *PermissionsHandler {
	return &PermissionsHandler{
		permService:  permService,
		groupService: groupService,
		auth:         auth,
	}
}

func (h *PermissionsHandler) getUserID(r *http.Request) (uuid.UUID, error) {
	return authServices.GetUserIDFromRequest(r, h.auth)
}

func (h *PermissionsHandler) GetPermissionsModal(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	perms, err := h.permService.GetAssetPermissions(r.Context(), db.AssetType(assetType), assetID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	groups, err := h.groupService.GetGroupsWhereAdmin(r.Context(), userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.PermissionsModal(assetType, assetID.String(), perms, groups, t).Render(r.Context(), w)
}

func (h *PermissionsHandler) GrantPermission(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	recipientTypeStr := r.FormValue("recipient_type")
	email := strings.TrimSpace(r.FormValue("email"))
	groupIDStr := strings.TrimSpace(r.FormValue("group_id"))
	permissionStr := r.FormValue("permission")

	var permission db.PermissionType
	if permissionStr == "write" {
		permission = db.PermissionTypeWrite
	} else {
		permission = db.PermissionTypeRead
	}

	var recipientType db.RecipientType
	var recipientID uuid.UUID

	if recipientTypeStr == "group" && groupIDStr != "" {
		recipientType = db.RecipientTypeGroup
		recipientID, err = uuid.Parse(groupIDStr)
		if err != nil {
			return ce.WithHTTPStatus(err, http.StatusBadRequest)
		}
	} else if email != "" {
		recipientType = db.RecipientTypeUser
	} else {
		return ce.WithHTTPStatus(errors.New("invalid recipient"), http.StatusBadRequest)
	}

	err = h.permService.Grant(r.Context(), db.AssetType(assetType), assetID, permission, recipientType, recipientID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *PermissionsHandler) RevokePermission(w http.ResponseWriter, r *http.Request) error {
	_, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	recipientTypeStr := r.FormValue("recipient_type")
	recipientIDStr := r.FormValue("recipient_id")
	permissionStr := r.FormValue("permission")

	recipientID, err := uuid.Parse(recipientIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	var recipientType db.RecipientType
	if recipientTypeStr == "group" {
		recipientType = db.RecipientTypeGroup
	} else {
		recipientType = db.RecipientTypeUser
	}

	var permission db.PermissionType
	switch permissionStr {
	case "owner":
		permission = db.PermissionTypeOwner
	case "write":
		permission = db.PermissionTypeWrite
	default:
		permission = db.PermissionTypeRead
	}

	err = h.permService.Revoke(r.Context(), db.AssetType(assetType), assetID, permission, recipientType, recipientID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
