-- name: SetOwnerPermission :one
INSERT INTO asset_permissions (id, asset_type, asset_id, permission, recipient_type, recipient_id, grantor_id, created_at)
VALUES ($1, $2, $3, 'owner', 'user', $4, $4, $5)
ON CONFLICT (asset_type, asset_id, recipient_type, recipient_id, permission) DO UPDATE SET grantor_id = $4
RETURNING *;

-- name: GrantPermission :one
INSERT INTO asset_permissions (id, asset_type, asset_id, permission, recipient_type, recipient_id, grantor_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (asset_type, asset_id, recipient_type, recipient_id, permission) DO UPDATE SET grantor_id = $7
RETURNING *;

-- name: RevokePermission :exec
DELETE FROM asset_permissions
WHERE asset_type = $1 AND asset_id = $2 AND permission = $3 AND recipient_type = $4 AND recipient_id = $5;

-- name: GetAssetPermissions :many
SELECT ap.id, ap.asset_type, ap.asset_id, ap.permission, ap.recipient_type, ap.recipient_id, ap.grantor_id, ap.created_at,
       CASE WHEN ap.recipient_type = 'user' THEN u.name ELSE g.name END as recipient_name
FROM asset_permissions ap
LEFT JOIN users u ON ap.recipient_type = 'user' AND ap.recipient_id = u.id
LEFT JOIN user_groups g ON ap.recipient_type = 'group' AND ap.recipient_id = g.id
WHERE ap.asset_type = $1 AND ap.asset_id = $2
ORDER BY ap.permission DESC, ap.created_at ASC;

-- name: GetUserAssetPermissions :many
SELECT * FROM asset_permissions
WHERE recipient_type = 'user' AND recipient_id = $1;

-- name: GetGroupAssetPermissions :many
SELECT * FROM asset_permissions
WHERE recipient_type = 'group' AND recipient_id = ANY($1::uuid[]);

-- name: GetAccessibleAssetIDs :many
SELECT DISTINCT ap.asset_id FROM asset_permissions ap
WHERE ap.asset_type = $1
AND (
    (ap.recipient_type = 'user' AND ap.recipient_id = $2)
    OR (ap.recipient_type = 'group' AND ap.recipient_id = ANY($3::uuid[]))
);

-- name: DeleteAssetPermissionsByAsset :exec
DELETE FROM asset_permissions WHERE asset_type = $1 AND asset_id = $2;

-- name: HasPermission :one
SELECT EXISTS(
    SELECT 1 FROM asset_permissions
    WHERE asset_type = $1 AND asset_id = $2
    AND recipient_type = $3 AND recipient_id = $4
    AND permission = $5
);

-- name: HasPermissionLevel :one
SELECT EXISTS(
    SELECT 1 FROM asset_permissions
    WHERE asset_type = $1 AND asset_id = $2
    AND recipient_type = $3 AND recipient_id = $4
    AND (
        ($5 = 'owner' AND asset_permission = 'owner')
        OR ($5 = 'write' AND asset_permission IN ('owner', 'write'))
        OR ($5 = 'read')
    )
);
