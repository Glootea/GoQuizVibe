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
	ErrNoPermission      = errors.New("no permission")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrAssetNotFound    = errors.New("asset not found")
	ErrCannotRevokeOwner = errors.New("cannot revoke owner permission")
)

type PermissionsService struct {
	perms  r.AssetPermissionRepository
	groups r.UserGroupRepository
}

func NewPermissionsService(perms r.AssetPermissionRepository, groups r.UserGroupRepository) *PermissionsService {
	return &PermissionsService{
		perms:  perms,
		groups: groups,
	}
}

func (s *PermissionsService) SetOwner(ctx context.Context, assetType string, assetID, ownerID uuid.UUID) error {
	_, err := s.perms.SetOwnerPermission(ctx, db.SetOwnerPermissionParams{
		ID:          uuid.New(),
		AssetType:   assetType,
		AssetID:     assetID,
		RecipientID: ownerID,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		return fmt.Errorf("set owner permission: %w", err)
	}
	return nil
}

func (s *PermissionsService) Grant(ctx context.Context, assetType string, assetID uuid.UUID, permission db.PermissionType, recipientType db.RecipientType, recipientID, grantorID uuid.UUID) error {
	if recipientType == db.RecipientTypeUser && permission == db.PermissionTypeOwner {
		return fmt.Errorf("cannot grant owner permission to user")
	}

	_, err := s.perms.GrantPermission(ctx, db.GrantPermissionParams{
		ID:            uuid.New(),
		AssetType:     assetType,
		AssetID:       assetID,
		Permission:    permission,
		RecipientType: recipientType,
		RecipientID:   recipientID,
		GrantorID:     grantorID,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return fmt.Errorf("grant permission: %w", err)
	}
	return nil
}

func (s *PermissionsService) Revoke(ctx context.Context, assetType string, assetID uuid.UUID, permission db.PermissionType, recipientType db.RecipientType, recipientID uuid.UUID) error {
	if recipientType == db.RecipientTypeUser && permission == db.PermissionTypeOwner {
		return ErrCannotRevokeOwner
	}

	err := s.perms.RevokePermission(ctx, db.RevokePermissionParams{
		AssetType:     assetType,
		AssetID:       assetID,
		Permission:    permission,
		RecipientType: recipientType,
		RecipientID:   recipientID,
	})
	if err != nil {
		return fmt.Errorf("revoke permission: %w", err)
	}
	return nil
}

func (s *PermissionsService) GetAssetPermissions(ctx context.Context, assetType string, assetID uuid.UUID) ([]PermissionWithGrantor, error) {
	rows, err := s.perms.GetAssetPermissions(ctx, db.GetAssetPermissionsParams{
		AssetType: assetType,
		AssetID:   assetID,
	})
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}

	result := make([]PermissionWithGrantor, 0, len(rows))
	for _, r := range rows {
		result = append(result, PermissionWithGrantor{
			ID:            r.ID,
			AssetType:     r.AssetType,
			AssetID:       r.AssetID,
			Permission:    db.PermissionType(fmt.Sprintf("%v", r.Permission)),
			RecipientType: db.RecipientType(fmt.Sprintf("%v", r.RecipientType)),
			RecipientID:   r.RecipientID,
			GrantorID:     r.GrantorID,
			CreatedAt:     r.CreatedAt,
			RecipientName: fmt.Sprintf("%v", r.RecipientName),
		})
	}
	return result, nil
}

func (s *PermissionsService) CanAccess(ctx context.Context, assetType string, assetID, userID uuid.UUID, requiredPermission db.PermissionType) (bool, error) {
	userGroups, err := s.groups.GetUserGroupsByAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get user groups: %w", err)
	}
	groupIDs := make([]uuid.UUID, len(userGroups))
	for i, g := range userGroups {
		groupIDs[i] = g.ID
	}

	hasUserPerm, err := s.perms.HasPermissionLevel(ctx, db.HasPermissionLevelParams{
		AssetType:     assetType,
		AssetID:       assetID,
		RecipientType: db.RecipientTypeUser,
		RecipientID:   userID,
		Column5:       requiredPermission,
	})
	if err != nil {
		return false, fmt.Errorf("check user permission: %w", err)
	}
	if hasUserPerm {
		return true, nil
	}

	if len(groupIDs) > 0 {
		hasGroupPerm, err := s.perms.HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     assetType,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeGroup,
			RecipientID:   groupIDs[0],
			Column5:       requiredPermission,
		})
		if err == nil && hasGroupPerm {
			return true, nil
		}
	}

	return false, nil
}

func (s *PermissionsService) CanRead(ctx context.Context, assetType string, assetID, userID uuid.UUID) (bool, error) {
	return s.CanAccess(ctx, assetType, assetID, userID, db.PermissionTypeRead)
}

func (s *PermissionsService) CanWrite(ctx context.Context, assetType string, assetID, userID uuid.UUID) (bool, error) {
	return s.CanAccess(ctx, assetType, assetID, userID, db.PermissionTypeWrite)
}

func (s *PermissionsService) IsOwner(ctx context.Context, assetType string, assetID, userID uuid.UUID) (bool, error) {
	return s.CanAccess(ctx, assetType, assetID, userID, db.PermissionTypeOwner)
}

func (s *PermissionsService) GetAccessibleAssetIDs(ctx context.Context, assetType string, userID uuid.UUID, groupIDs []uuid.UUID) ([]uuid.UUID, error) {
	ids, err := s.perms.GetAccessibleAssetIDs(ctx, db.GetAccessibleAssetIDsParams{
		AssetType:   assetType,
		RecipientID: userID,
		Column3:     groupIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("get accessible assets: %w", err)
	}
	return ids, nil
}

type PermissionWithGrantor struct {
	ID            uuid.UUID
	AssetType     string
	AssetID       uuid.UUID
	Permission    db.PermissionType
	RecipientType db.RecipientType
	RecipientID   uuid.UUID
	GrantorID     uuid.UUID
	CreatedAt     time.Time
	RecipientName string
}
