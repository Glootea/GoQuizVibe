package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	permissionsSvc "github.com/goquizvibe/backend/feature/permissions/services"
	"github.com/goquizvibe/backend/shared/db"
)

func TestPermissionsService_SetOwner(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	assetID := uuid.New()

	t.Run("successful owner permission set", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Perms.EXPECT().SetOwnerPermission(ctx, gomockAny()).Return(db.AssetPermission{}, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.SetOwner(ctx, db.AssetTypeLearningMaterial, assetID, ownerID)
		if err != nil {
			t.Fatalf("SetOwner() error = %v, want nil", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Perms.EXPECT().SetOwnerPermission(ctx, gomockAny()).Return(db.AssetPermission{}, errors.New("db error"))

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.SetOwner(ctx, db.AssetTypeLearningMaterial, assetID, ownerID)
		if err == nil {
			t.Fatal("SetOwner() error = nil, want error")
		}
	})
}

func TestPermissionsService_Grant(t *testing.T) {
	ctx := context.Background()
	assetID := uuid.New()
	grantorID := uuid.New()
	recipientID := uuid.New()

	t.Run("grant read permission to user", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Perms.EXPECT().GrantPermission(ctx, gomockAny()).Return(db.AssetPermission{}, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Grant(ctx, db.AssetTypeLearningMaterial, assetID, db.PermissionTypeRead, db.RecipientTypeUser, recipientID, grantorID)
		if err != nil {
			t.Fatalf("Grant() error = %v, want nil", err)
		}
	})

	t.Run("grant write permission to group", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)
		groupID := uuid.New()

		m.Perms.EXPECT().GrantPermission(ctx, gomockAny()).Return(db.AssetPermission{}, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Grant(ctx, db.AssetTypeQuiz, assetID, db.PermissionTypeWrite, db.RecipientTypeGroup, groupID, grantorID)
		if err != nil {
			t.Fatalf("Grant() error = %v, want nil", err)
		}
	})

	t.Run("cannot grant owner permission to user", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Grant(ctx, db.AssetTypeLearningMaterial, assetID, db.PermissionTypeOwner, db.RecipientTypeUser, recipientID, grantorID)
		if err == nil {
			t.Fatal("Grant() error = nil, want error (cannot grant owner to user)")
		}
	})

	t.Run("grant owner permission to group succeeds", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)
		groupID := uuid.New()

		m.Perms.EXPECT().GrantPermission(ctx, gomockAny()).Return(db.AssetPermission{}, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Grant(ctx, db.AssetTypeQuiz, assetID, db.PermissionTypeOwner, db.RecipientTypeGroup, groupID, grantorID)
		if err != nil {
			t.Fatalf("Grant() error = %v, want nil", err)
		}
	})
}

func TestPermissionsService_Revoke(t *testing.T) {
	ctx := context.Background()
	assetID := uuid.New()
	recipientID := uuid.New()

	t.Run("revoke read permission from user", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Perms.EXPECT().RevokePermission(ctx, db.RevokePermissionParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			Permission:    db.PermissionTypeRead,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   recipientID,
		}).Return(nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Revoke(ctx, db.AssetTypeLearningMaterial, assetID, db.PermissionTypeRead, db.RecipientTypeUser, recipientID)
		if err != nil {
			t.Fatalf("Revoke() error = %v, want nil", err)
		}
	})

	t.Run("cannot revoke owner permission from user", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Revoke(ctx, db.AssetTypeLearningMaterial, assetID, db.PermissionTypeOwner, db.RecipientTypeUser, recipientID)
		if err == nil {
			t.Fatal("Revoke() error = nil, want error (cannot revoke owner from user)")
		}
		if err != permissionsSvc.ErrCannotRevokeOwner {
			t.Errorf("Revoke() error = %v, want ErrCannotRevokeOwner", err)
		}
	})

	t.Run("revoke owner permission from group succeeds", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)
		groupID := uuid.New()

		m.Perms.EXPECT().RevokePermission(ctx, db.RevokePermissionParams{
			AssetType:     db.AssetTypeQuiz,
			AssetID:       assetID,
			Permission:    db.PermissionTypeOwner,
			RecipientType: db.RecipientTypeGroup,
			RecipientID:   groupID,
		}).Return(nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.Revoke(ctx, db.AssetTypeQuiz, assetID, db.PermissionTypeOwner, db.RecipientTypeGroup, groupID)
		if err != nil {
			t.Fatalf("Revoke() error = %v, want nil", err)
		}
	})
}

func TestPermissionsService_CanAccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	assetID := uuid.New()
	groupID := uuid.New()

	t.Run("user has direct permission", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeWrite,
		}).Return(true, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.CanAccess(ctx, db.AssetTypeLearningMaterial, assetID, userID, db.PermissionTypeWrite)
		if err != nil {
			t.Fatalf("CanAccess() error = %v, want nil", err)
		}
		if !result {
			t.Error("CanAccess() = false, want true")
		}
	})

	t.Run("user has group permission", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{{ID: groupID, Name: "Test Group"}}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeWrite,
		}).Return(false, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeGroup,
			RecipientID:   groupID,
			Column5:       db.PermissionTypeWrite,
		}).Return(true, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.CanAccess(ctx, db.AssetTypeLearningMaterial, assetID, userID, db.PermissionTypeWrite)
		if err != nil {
			t.Fatalf("CanAccess() error = %v, want nil", err)
		}
		if !result {
			t.Error("CanAccess() = false, want true (via group)")
		}
	})

	t.Run("user has no permission", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{{ID: groupID}}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeRead,
		}).Return(false, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeGroup,
			RecipientID:   groupID,
			Column5:       db.PermissionTypeRead,
		}).Return(false, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.CanAccess(ctx, db.AssetTypeLearningMaterial, assetID, userID, db.PermissionTypeRead)
		if err != nil {
			t.Fatalf("CanAccess() error = %v, want nil", err)
		}
		if result {
			t.Error("CanAccess() = true, want false")
		}
	})

	t.Run("groups retrieval error", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return(nil, errors.New("groups error"))

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		_, err := svc.CanAccess(ctx, db.AssetTypeLearningMaterial, assetID, userID, db.PermissionTypeRead)
		if err == nil {
			t.Fatal("CanAccess() error = nil, want error")
		}
	})
}

func TestPermissionsService_CanRead(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	assetID := uuid.New()

	t.Run("user has read permission", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeQuiz,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeRead,
		}).Return(true, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.CanRead(ctx, db.AssetTypeQuiz, assetID, userID)
		if err != nil {
			t.Fatalf("CanRead() error = %v, want nil", err)
		}
		if !result {
			t.Error("CanRead() = false, want true")
		}
	})
}

func TestPermissionsService_CanWrite(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	assetID := uuid.New()

	t.Run("user has write permission", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeWrite,
		}).Return(true, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.CanWrite(ctx, db.AssetTypeLearningMaterial, assetID, userID)
		if err != nil {
			t.Fatalf("CanWrite() error = %v, want nil", err)
		}
		if !result {
			t.Error("CanWrite() = false, want true")
		}
	})

	t.Run("user has no write permission", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeWrite,
		}).Return(false, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.CanWrite(ctx, db.AssetTypeLearningMaterial, assetID, userID)
		if err != nil {
			t.Fatalf("CanWrite() error = %v, want nil", err)
		}
		if result {
			t.Error("CanWrite() = true, want false")
		}
	})
}

func TestPermissionsService_IsOwner(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	assetID := uuid.New()

	t.Run("user is owner", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeQuiz,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeOwner,
		}).Return(true, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.IsOwner(ctx, db.AssetTypeQuiz, assetID, userID)
		if err != nil {
			t.Fatalf("IsOwner() error = %v, want nil", err)
		}
		if !result {
			t.Error("IsOwner() = false, want true")
		}
	})

	t.Run("user is not owner", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeQuiz,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeOwner,
		}).Return(false, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.IsOwner(ctx, db.AssetTypeQuiz, assetID, userID)
		if err != nil {
			t.Fatalf("IsOwner() error = %v, want nil", err)
		}
		if result {
			t.Error("IsOwner() = true, want false")
		}
	})
}

func TestPermissionsService_GetUserPermissionLevel(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	assetID := uuid.New()

	t.Run("returns owner when user is owner", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Groups.EXPECT().GetUserGroupsByAdmin(ctx, userID).Return([]db.UserGroup{}, nil)
		m.Perms.EXPECT().HasPermissionLevel(ctx, db.HasPermissionLevelParams{
			AssetType:     db.AssetTypeLearningMaterial,
			AssetID:       assetID,
			RecipientType: db.RecipientTypeUser,
			RecipientID:   userID,
			Column5:       db.PermissionTypeOwner,
		}).Return(true, nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		result, err := svc.GetUserPermissionLevel(ctx, db.AssetTypeLearningMaterial, assetID, userID)
		if err != nil {
			t.Fatalf("GetUserPermissionLevel() error = %v, want nil", err)
		}
		if result != db.PermissionTypeOwner {
			t.Errorf("GetUserPermissionLevel() = %v, want owner", result)
		}
	})
}

func TestPermissionsService_RevokeAllForAsset(t *testing.T) {
	ctx := context.Background()
	assetID := uuid.New()

	t.Run("successful revoke all permissions", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Perms.EXPECT().DeleteAssetPermissionsByAsset(ctx, db.DeleteAssetPermissionsByAssetParams{
			AssetType: db.AssetTypeLearningMaterial,
			AssetID:   assetID,
		}).Return(nil)

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.RevokeAllForAsset(ctx, db.AssetTypeLearningMaterial, assetID)
		if err != nil {
			t.Fatalf("RevokeAllForAsset() error = %v, want nil", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		m := NewPermissionsServiceMocks(t)

		m.Perms.EXPECT().DeleteAssetPermissionsByAsset(ctx, db.DeleteAssetPermissionsByAssetParams{
			AssetType: db.AssetTypeQuiz,
			AssetID:   assetID,
		}).Return(errors.New("db error"))

		svc := permissionsSvc.NewPermissionsService(m.Perms, m.Groups, m.StudentAccess)
		err := svc.RevokeAllForAsset(ctx, db.AssetTypeQuiz, assetID)
		if err == nil {
			t.Fatal("RevokeAllForAsset() error = nil, want error")
		}
	})
}
