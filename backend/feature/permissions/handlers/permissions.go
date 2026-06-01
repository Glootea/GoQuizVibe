package handlers

import (
	"database/sql"
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
	r "github.com/goquizvibe/backend/shared/repositories"
)

type PermissionsHandler struct {
	permService  *services.PermissionsService
	groupService *services.UserGroupService
	auth         interfaces.Authenticator
	usersRepo    r.UserRepository
	quizRepo     r.QuizRepository
	materialRepo r.LearningMaterialRepository
}

func NewPermissionsHandler(
	permService *services.PermissionsService,
	groupService *services.UserGroupService,
	auth interfaces.Authenticator,
	usersRepo r.UserRepository,
	quizRepo r.QuizRepository,
	materialRepo r.LearningMaterialRepository,
) *PermissionsHandler {
	return &PermissionsHandler{
		permService:  permService,
		groupService: groupService,
		auth:         auth,
		usersRepo:    usersRepo,
		quizRepo:     quizRepo,
		materialRepo: materialRepo,
	}
}

func (h *PermissionsHandler) getUserID(r *http.Request) (uuid.UUID, error) {
	return authServices.GetUserIDFromRequest(r, h.auth)
}

func (h *PermissionsHandler) GetPermissionsPage(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	var assetTitle string
	var studentPermission db.StudentPermission = db.StudentPermissionPrivate
	switch assetType {
	case "quiz":
		quiz, err := h.quizRepo.GetQuizByID(r.Context(), assetID)
		if err == nil {
			assetTitle = quiz.Title
			studentPermission = db.StudentPermission(quiz.StudentPermission)
		}
	case "learning_material":
		material, err := h.materialRepo.GetLearningMaterialByID(r.Context(), assetID)
		if err == nil {
			assetTitle = material.Title
			studentPermission = db.StudentPermission(material.StudentPermission)
		}
	}

	perms, err := h.permService.GetAssetPermissions(r.Context(), db.AssetType(assetType), assetID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	groups, err := h.groupService.GetGroupsWhereAdmin(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	studentAccessList, _ := h.permService.GetStudentAccessList(r.Context(), db.AssetType(assetType), assetID)

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.PermissionsPage(assetType, assetID.String(), assetTitle, studentPermission, perms, studentAccessList, groups, &user, t, "").Render(r.Context(), w)
}

func (h *PermissionsHandler) getUser(r *http.Request) (db.User, error) {
	uid, err := h.getUserID(r)
	if err != nil {
		return db.User{}, err
	}
	return h.usersRepo.GetUserByID(r.Context(), uid)
}

func (h *PermissionsHandler) GrantPermission(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
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
		foundUser, err := h.usersRepo.GetUserByEmail(r.Context(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set") {
				return h.renderPermissionsPageWithError(w, r, user, assetType, assetID, "", "Пользователь с таким email не найден")
			}
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}
		recipientID = foundUser.ID
	} else {
		return h.renderPermissionsPageWithError(w, r, user, assetType, assetID, "", "Введите email пользователя или выберите группу")
	}

	err = h.permService.Grant(r.Context(), db.AssetType(assetType), assetID, permission, recipientType, recipientID, user.ID)
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

func (h *PermissionsHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) error {
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
	recipientIDStr := r.FormValue("recipient_id")
	currentPermissionStr := r.FormValue("current_permission")
	newPermissionStr := r.FormValue("permission")

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

	var currentPermission, newPermission db.PermissionType
	switch currentPermissionStr {
	case "owner":
		currentPermission = db.PermissionTypeOwner
	case "write":
		currentPermission = db.PermissionTypeWrite
	default:
		currentPermission = db.PermissionTypeRead
	}

	switch newPermissionStr {
	case "owner":
		newPermission = db.PermissionTypeOwner
	case "write":
		newPermission = db.PermissionTypeWrite
	default:
		newPermission = db.PermissionTypeRead
	}

	if newPermission == currentPermission {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	err = h.permService.Revoke(r.Context(), db.AssetType(assetType), assetID, currentPermission, recipientType, recipientID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if newPermission != db.PermissionTypeOwner {
		err = h.permService.Grant(r.Context(), db.AssetType(assetType), assetID, newPermission, recipientType, recipientID, userID)
		if err != nil {
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *PermissionsHandler) renderPermissionsPageWithError(w http.ResponseWriter, r *http.Request, user db.User, assetType string, assetID uuid.UUID, assetTitle string, errorMsg string) error {
	groups, err := h.groupService.GetGroupsWhereAdmin(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.GrantAccessForm(assetType, assetID.String(), groups, errorMsg, t).Render(r.Context(), w)
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

	recipientTypeStr := r.URL.Query().Get("recipient_type")
	recipientIDStr := r.URL.Query().Get("recipient_id")
	permissionStr := r.URL.Query().Get("permission")

	if recipientIDStr == "" {
		return ce.WithHTTPStatus(errors.New("recipient_id is required: recipient_type="+recipientTypeStr+", permission="+permissionStr), http.StatusBadRequest)
	}

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

func (h *PermissionsHandler) GetStudentAccessList(w http.ResponseWriter, r *http.Request) error {
	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	accessList, err := h.permService.GetStudentAccessList(r.Context(), db.AssetType(assetType), assetID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.StudentAccessList(accessList, assetType, assetIDStr, t).Render(r.Context(), w)
}

func (h *PermissionsHandler) GrantStudentAccess(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	groups, err := h.groupService.GetGroupsWhereAdmin(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	recipientTypeStr := r.FormValue("recipient_type")
	email := strings.TrimSpace(r.FormValue("email"))
	groupIDStr := strings.TrimSpace(r.FormValue("group_id"))

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
		foundUser, err := h.usersRepo.GetUserByEmail(r.Context(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "no rows in result set") {
				t := middleware.GetTranslator(r.Context())
				return permissionsUI.StudentAccessFormWithError(assetType, assetIDStr, groups, t.UserNotFound(), t).Render(r.Context(), w)
			}
			return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
		}
		recipientID = foundUser.ID
	} else {
		t := middleware.GetTranslator(r.Context())
		return permissionsUI.StudentAccessFormWithError(assetType, assetIDStr, groups, t.EnterEmailOrSelectGroup(), t).Render(r.Context(), w)
	}

	err = h.permService.GrantStudentAccess(r.Context(), db.AssetType(assetType), assetID, recipientType, recipientID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	studentAccessList, err := h.permService.GetStudentAccessList(r.Context(), db.AssetType(assetType), assetID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.StudentAccessContainer(assetType, assetIDStr, studentAccessList, groups, t).Render(r.Context(), w)
}

func (h *PermissionsHandler) RevokeStudentAccess(w http.ResponseWriter, r *http.Request) error {
	assetType := r.PathValue("type")
	assetIDStr := r.PathValue("id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	accessIDStr := r.PathValue("accessId")
	accessID, err := uuid.Parse(accessIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	accessList, err := h.permService.GetStudentAccessList(r.Context(), db.AssetType(assetType), assetID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	var accessToDelete *services.StudentAccessWithRecipient
	for _, a := range accessList {
		if a.ID == accessID {
			accessToDelete = &a
			break
		}
	}

	if accessToDelete == nil {
		return ce.WithHTTPStatus(errors.New("access not found"), http.StatusNotFound)
	}

	err = h.permService.RevokeStudentAccess(r.Context(), db.AssetType(assetType), assetID, accessToDelete.RecipientType, accessToDelete.RecipientID)
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

func (h *PermissionsHandler) UpdateStudentPermission(w http.ResponseWriter, r *http.Request) error {
	user, err := h.getUser(r)
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

	studentPermissionStr := r.FormValue("student_permission")

	var studentPermission db.StudentPermission
	switch studentPermissionStr {
	case "open_to_all":
		studentPermission = db.StudentPermissionOpenToAll
	case "assigned":
		studentPermission = db.StudentPermissionAssigned
	case "private":
		studentPermission = db.StudentPermissionPrivate
	default:
		return ce.WithHTTPStatus(errors.New("invalid student permission"), http.StatusBadRequest)
	}

	var errUpdate error
	switch assetType {
	case "quiz":
		_, err = h.quizRepo.GetQuizByID(r.Context(), assetID)
		if err == nil {
			errUpdate = h.quizRepo.UpdateQuizStudentPermission(r.Context(), db.UpdateQuizStudentPermissionParams{
				ID:                assetID,
				StudentPermission: studentPermission,
			})
		}
	case "learning_material":
		_, err = h.materialRepo.GetLearningMaterialByID(r.Context(), assetID)
		if err == nil {
			errUpdate = h.materialRepo.UpdateLearningMaterialStudentPermission(r.Context(), db.UpdateLearningMaterialStudentPermissionParams{
				ID:                assetID,
				StudentPermission: studentPermission,
			})
		}
	}

	if errUpdate != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, errUpdate), http.StatusInternalServerError)
	}

	groups, err := h.groupService.GetGroupsWhereAdmin(r.Context(), user.ID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	studentAccessList, err := h.permService.GetStudentAccessList(r.Context(), db.AssetType(assetType), assetID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.StudentPermissionSection(assetType, assetIDStr, studentPermission, studentAccessList, groups, t).Render(r.Context(), w)
}
