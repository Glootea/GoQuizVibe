-- name: GrantStudentAccess :one
INSERT INTO student_access (id, asset_type, asset_id, recipient_type, recipient_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (asset_type, asset_id, recipient_type, recipient_id) DO UPDATE SET recipient_id = $5
RETURNING *;

-- name: RevokeStudentAccess :exec
DELETE FROM student_access
WHERE asset_type = $1 AND asset_id = $2 AND recipient_type = $3 AND recipient_id = $4;

-- name: GetStudentAccessList :many
SELECT sa.id, sa.asset_type, sa.asset_id, sa.recipient_type, sa.recipient_id, sa.created_at,
       CASE WHEN sa.recipient_type = 'user' THEN u.name ELSE g.name END as recipient_name,
       CASE WHEN sa.recipient_type = 'user' THEN u.email ELSE NULL END as recipient_email
FROM student_access sa
LEFT JOIN users u ON sa.recipient_type = 'user' AND sa.recipient_id = u.id
LEFT JOIN user_groups g ON sa.recipient_type = 'group' AND sa.recipient_id = g.id
WHERE sa.asset_type = $1 AND sa.asset_id = $2
ORDER BY sa.created_at ASC;

-- name: HasStudentAccess :one
SELECT EXISTS(
    SELECT 1 FROM student_access
    WHERE asset_type = $1 AND asset_id = $2
    AND ((recipient_type = 'user' AND recipient_id = $3)
         OR (recipient_type = 'group' AND recipient_id = ANY($4::uuid[])))
);

-- name: GetStudentAccessAssetIDs :many
SELECT DISTINCT sa.asset_id FROM student_access sa
WHERE sa.asset_type = $1
AND ((sa.recipient_type = 'user' AND sa.recipient_id = $2)
     OR (sa.recipient_type = 'group' AND sa.recipient_id = ANY($3::uuid[])));

-- name: DeleteStudentAccessByAsset :exec
DELETE FROM student_access WHERE asset_type = $1 AND asset_id = $2;
