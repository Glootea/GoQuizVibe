package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/shared/db"
	r "github.com/goquizvibe/backend/shared/repositories"
)

var (
	ErrGroupNotFound   = errors.New("group not found")
	ErrNotGroupAdmin   = errors.New("not a group admin")
	ErrNotGroupMember  = errors.New("not a group member")
	ErrNotGroupOwner   = errors.New("not a group owner")
	ErrCannotLeaveGroup = errors.New("cannot leave group")
	ErrAlreadyMember   = errors.New("user is already a member")
)

type UserGroupService struct {
	repo  r.UserGroupRepository
	users r.UserRepository
}

func NewUserGroupService(repo r.UserGroupRepository, users r.UserRepository) *UserGroupService {
	return &UserGroupService{
		repo:  repo,
		users: users,
	}
}

func (s *UserGroupService) CreateGroup(ctx context.Context, name, description string, ownerID uuid.UUID) (*db.UserGroup, error) {
	groupID := uuid.New()
	group, err := s.repo.CreateUserGroup(ctx, db.CreateUserGroupParams{
		ID:          groupID,
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	_, err = s.repo.AddUserToGroup(ctx, db.AddUserToGroupParams{
		GroupID:  groupID,
		UserID:   ownerID,
		Role:     db.GroupRoleAdmin,
		JoinedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("add owner as admin: %w", err)
	}

	return &group, nil
}

func (s *UserGroupService) GetGroupsWhereAdmin(ctx context.Context, userID uuid.UUID) ([]db.UserGroup, error) {
	groups, err := s.repo.GetUserGroupsByAdmin(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}
	return groups, nil
}

func (s *UserGroupService) GetGroupByID(ctx context.Context, groupID uuid.UUID) (*db.UserGroup, error) {
	group, err := s.repo.GetUserGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	return &group, nil
}

func (s *UserGroupService) UpdateGroup(ctx context.Context, groupID, requesterID uuid.UUID, name, description string) (*db.UserGroup, error) {
	if !s.isAdmin(ctx, groupID, requesterID) {
		return nil, ErrNotGroupAdmin
	}

	group, err := s.repo.UpdateUserGroup(ctx, db.UpdateUserGroupParams{
		ID:          groupID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	return &group, nil
}

func (s *UserGroupService) DeleteGroup(ctx context.Context, groupID, requesterID uuid.UUID) error {
	group, err := s.repo.GetUserGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	if group.OwnerID != requesterID {
		return ErrNotGroupOwner
	}

	err = s.repo.DeleteUserGroup(ctx, db.DeleteUserGroupParams{
		ID:      groupID,
		OwnerID: requesterID,
	})
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	return nil
}

func (s *UserGroupService) AddMember(ctx context.Context, groupID, userID, addedByID uuid.UUID, email string) error {
	if !s.isAdmin(ctx, groupID, addedByID) {
		return ErrNotGroupAdmin
	}

	if email != "" {
		user, err := s.users.GetUserByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("get user by email: %w", err)
		}
		userID = user.ID
	}

	memberCheck, err := s.repo.IsUserMemberOfGroup(ctx, db.IsUserMemberOfGroupParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if memberCheck {
		return ErrAlreadyMember
	}

	_, err = s.repo.AddUserToGroup(ctx, db.AddUserToGroupParams{
		GroupID:  groupID,
		UserID:   userID,
		Role:     db.GroupRoleMember,
		JoinedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("add user to group: %w", err)
	}
	return nil
}

func (s *UserGroupService) RemoveMember(ctx context.Context, groupID, targetUserID, removedByID uuid.UUID) error {
	if !s.isAdmin(ctx, groupID, removedByID) {
		return ErrNotGroupAdmin
	}

	group, err := s.repo.GetUserGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	if group.OwnerID == targetUserID {
		return fmt.Errorf("cannot remove owner")
	}

	err = s.repo.RemoveUserFromGroup(ctx, db.RemoveUserFromGroupParams{
		GroupID: groupID,
		UserID:  targetUserID,
	})
	if err != nil {
		return fmt.Errorf("remove user from group: %w", err)
	}
	return nil
}

func (s *UserGroupService) UpdateMemberRole(ctx context.Context, groupID, targetUserID, updatedByID uuid.UUID, newRole db.GroupRole) error {
	if !s.isAdmin(ctx, groupID, updatedByID) {
		return ErrNotGroupAdmin
	}

	group, err := s.repo.GetUserGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	if group.OwnerID == targetUserID && newRole != db.GroupRoleAdmin {
		return fmt.Errorf("cannot demote owner to member")
	}

	_, err = s.repo.AddUserToGroup(ctx, db.AddUserToGroupParams{
		GroupID:  groupID,
		UserID:   targetUserID,
		Role:     newRole,
		JoinedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}

func (s *UserGroupService) LeaveGroup(ctx context.Context, groupID, userID uuid.UUID) error {
	group, err := s.repo.GetUserGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	if group.OwnerID == userID {
		return ErrCannotLeaveGroup
	}

	memberCount, err := s.repo.GetGroupMemberCount(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get member count: %w", err)
	}
	if memberCount == 1 {
		return ErrCannotLeaveGroup
	}

	err = s.repo.RemoveUserFromGroup(ctx, db.RemoveUserFromGroupParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("leave group: %w", err)
	}
	return nil
}

func (s *UserGroupService) GetMembers(ctx context.Context, groupID uuid.UUID) ([]db.GetGroupMembersRow, error) {
	members, err := s.repo.GetGroupMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	return members, nil
}

func (s *UserGroupService) GetMembersWithUser(ctx context.Context, groupID uuid.UUID, currentUserID uuid.UUID) ([]MemberWithRole, error) {
	members, err := s.GetMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	group, err := s.repo.GetUserGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	result := make([]MemberWithRole, 0, len(members))
	for _, m := range members {
		roleStr := fmt.Sprintf("%v", m.Role)
		isOwner := group.OwnerID == m.ID
		isCurrentUser := currentUserID == m.ID
		canManage := false
		if currentUserID == group.OwnerID {
			canManage = true
		} else if s.isAdmin(ctx, groupID, currentUserID) {
			canManage = !isOwner && !isCurrentUser
		}

		result = append(result, MemberWithRole{
			ID:           m.ID,
			Name:         m.Name,
			Email:        m.Email,
			Role:         db.GroupRole(roleStr),
			JoinedAt:     m.JoinedAt,
			IsOwner:      isOwner,
			CanManage:    canManage,
		})
	}
	return result, nil
}

func (s *UserGroupService) isAdmin(ctx context.Context, groupID, userID uuid.UUID) bool {
	role, err := s.repo.GetUserRoleInGroup(ctx, db.GetUserRoleInGroupParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return false
	}
	return fmt.Sprintf("%v", role) == "admin"
}

func (s *UserGroupService) IsUserAdmin(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	role, err := s.repo.GetUserRoleInGroup(ctx, db.GetUserRoleInGroupParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return false, fmt.Errorf("get role: %w", err)
	}
	roleStr := fmt.Sprintf("%v", role)
	return roleStr == "admin" || roleStr == string(db.GroupRoleAdmin), nil
}

func (s *UserGroupService) GetUserAdminGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	groups, err := s.repo.GetUserGroupsByAdmin(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}
	ids := make([]uuid.UUID, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	return ids, nil
}

type MemberWithRole struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Role     db.GroupRole
	JoinedAt time.Time
	IsOwner  bool
	CanManage bool
}
