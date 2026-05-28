DROP INDEX IF EXISTS idx_asset_permissions_asset;
DROP INDEX IF EXISTS idx_asset_permissions_recipient;
DROP INDEX IF EXISTS idx_group_members_user_id;
DROP INDEX IF EXISTS idx_user_groups_owner_id;

DROP TABLE IF EXISTS asset_permissions;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS user_groups;

DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS recipient_type;
DROP TYPE IF EXISTS permission_type;
DROP TYPE IF EXISTS group_role;
