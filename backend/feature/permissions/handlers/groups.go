package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	authServices "github.com/goquizvibe/backend/feature/auth/services"
	permServices "github.com/goquizvibe/backend/feature/permissions/services"
	permissionsUI "github.com/goquizvibe/backend/feature/permissions/ui"
	ce "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	"github.com/goquizvibe/backend/shared/middleware"
	r "github.com/goquizvibe/backend/shared/repositories"
)

type GroupsHandler struct {
	groupService *permServices.UserGroupService
	users        r.UserRepository
	auth         interfaces.Authenticator
}

func NewGroupsHandler(groupService *permServices.UserGroupService, users r.UserRepository, auth interfaces.Authenticator) *GroupsHandler {
	return &GroupsHandler{
		groupService: groupService,
		users:        users,
		auth:         auth,
	}
}

func (h *GroupsHandler) getUserID(r *http.Request) (uuid.UUID, error) {
	return authServices.GetUserIDFromRequest(r, h.auth)
}

func (h *GroupsHandler) ListGroups(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groups, err := h.groupService.GetGroupsWhereAdmin(r.Context(), userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	user, err := h.users.GetUserByID(r.Context(), userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.GroupsListPage(groups, &user, t).Render(r.Context(), w)
}

func (h *GroupsHandler) GetGroup(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	group, err := h.groupService.GetGroupByID(r.Context(), groupID)
	if err != nil {
		return ce.WithHTTPStatus(ce.ErrNotFound, http.StatusNotFound)
	}

	isAdmin, err := h.groupService.IsUserAdmin(r.Context(), groupID, userID)
	if err != nil || !isAdmin {
		return ce.WithHTTPStatus(ce.ErrForbidden, http.StatusForbidden)
	}

	members, err := h.groupService.GetMembersWithUser(r.Context(), groupID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	user, err := h.users.GetUserByID(r.Context(), userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.GroupDetailPage(group, &user, members, t).Render(r.Context(), w)
}

func (h *GroupsHandler) CreateGroupForm(w http.ResponseWriter, r *http.Request) error {
	t := middleware.GetTranslator(r.Context())
	return permissionsUI.CreateGroupModal(t).Render(r.Context(), w)
}

func (h *GroupsHandler) CreateGroup(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))

	if name == "" {
		name = "Untitled Group"
	}

	group, err := h.groupService.CreateGroup(r.Context(), name, description, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/admin/groups/"+group.ID.String())
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups/"+group.ID.String(), http.StatusFound)
	return nil
}

func (h *GroupsHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))

	_, err = h.groupService.UpdateGroup(r.Context(), groupID, userID, name, description)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups/"+groupIDStr, http.StatusFound)
	return nil
}

func (h *GroupsHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	err = h.groupService.DeleteGroup(r.Context(), groupID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/admin/groups")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups", http.StatusFound)
	return nil
}

func (h *GroupsHandler) GetMembers(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	isAdmin, err := h.groupService.IsUserAdmin(r.Context(), groupID, userID)
	if err != nil || !isAdmin {
		return ce.WithHTTPStatus(ce.ErrForbidden, http.StatusForbidden)
	}

	members, err := h.groupService.GetMembersWithUser(r.Context(), groupID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return permissionsUI.MembersList(members, t).Render(r.Context(), w)
}

func (h *GroupsHandler) AddMember(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	email := strings.TrimSpace(r.FormValue("email"))

	err = h.groupService.AddMember(r.Context(), groupID, uuid.Nil, userID, email)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups/"+groupIDStr, http.StatusFound)
	return nil
}

func (h *GroupsHandler) RemoveMember(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	memberIDStr := r.PathValue("memberID")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	err = h.groupService.RemoveMember(r.Context(), groupID, memberID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups/"+groupIDStr, http.StatusFound)
	return nil
}

func (h *GroupsHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	memberIDStr := r.PathValue("memberID")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	roleStr := r.FormValue("role")
	var newRole db.GroupRole
	if roleStr == "admin" {
		newRole = db.GroupRoleAdmin
	} else {
		newRole = db.GroupRoleMember
	}

	err = h.groupService.UpdateMemberRole(r.Context(), groupID, memberID, userID, newRole)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups/"+groupIDStr, http.StatusFound)
	return nil
}

func (h *GroupsHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) error {
	userID, err := h.getUserID(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	groupIDStr := r.PathValue("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return ce.WithHTTPStatus(err, http.StatusBadRequest)
	}

	err = h.groupService.LeaveGroup(r.Context(), groupID, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	if middleware.IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/admin/groups")
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/admin/groups", http.StatusFound)
	return nil
}
