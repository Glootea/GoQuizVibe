-- name: CreateUserGroup :one
INSERT INTO user_groups (id, name, description, owner_id, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserGroupByID :one
SELECT * FROM user_groups WHERE id = $1;

-- name: GetUserGroupsByAdmin :many
SELECT g.* FROM user_groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = $1 AND gm.role = 'admin'
ORDER BY g.created_at DESC;

-- name: UpdateUserGroup :one
UPDATE user_groups SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUserGroup :exec
DELETE FROM user_groups WHERE id = $1 AND owner_id = $2;

-- name: AddUserToGroup :one
INSERT INTO group_members (group_id, user_id, role, joined_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (group_id, user_id) DO UPDATE SET role = $3
RETURNING *;

-- name: RemoveUserFromGroup :exec
DELETE FROM group_members WHERE group_id = $1 AND user_id = $2;

-- name: GetGroupMembers :many
SELECT u.id, u.name, u.email, gm.role, gm.joined_at
FROM users u
JOIN group_members gm ON u.id = gm.user_id
WHERE gm.group_id = $1
ORDER BY gm.role DESC, gm.joined_at ASC;

-- name: GetUserRoleInGroup :one
SELECT role FROM group_members WHERE group_id = $1 AND user_id = $2;

-- name: GetGroupMemberCount :one
SELECT COUNT(*) FROM group_members WHERE group_id = $1;

-- name: IsUserMemberOfGroup :one
SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2);
