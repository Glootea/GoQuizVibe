-- name: GetLearningMaterials :many
SELECT * FROM learning_materials ORDER BY created_at DESC;

-- name: GetLearningMaterialsForUser :many
SELECT DISTINCT lm.* FROM learning_materials lm
WHERE EXISTS (
    SELECT 1 FROM asset_permissions ap
    WHERE ap.asset_type = 'learning_material' AND ap.asset_id = lm.id
    AND (
        (ap.recipient_type = 'user' AND ap.recipient_id = $1)
        OR (ap.recipient_type = 'group' AND ap.recipient_id = ANY($2::uuid[]))
    )
    AND ap.permission IN ('read', 'write', 'owner')
)
ORDER BY lm.created_at DESC;

-- name: GetLearningMaterialsForStudent :many
SELECT DISTINCT lm.* FROM learning_materials lm
WHERE (
    lm.student_permission = 'open_to_all'
    OR EXISTS (
        SELECT 1 FROM student_access sa
        WHERE sa.asset_type = 'learning_material' AND sa.asset_id = lm.id
        AND ((sa.recipient_type = 'user' AND sa.recipient_id = $1)
             OR (sa.recipient_type = 'group' AND sa.recipient_id = ANY($2::uuid[])))
    )
    OR EXISTS (
        SELECT 1 FROM asset_permissions ap
        WHERE ap.asset_type = 'learning_material' AND ap.asset_id = lm.id
        AND ((ap.recipient_type = 'user' AND ap.recipient_id = $1)
             OR (ap.recipient_type = 'group' AND ap.recipient_id = ANY($2::uuid[])))
        AND ap.permission IN ('read', 'write', 'owner')
    )
)
ORDER BY lm.created_at DESC;

-- name: HasLearningMaterialAccess :one
SELECT EXISTS(
    SELECT 1 FROM asset_permissions
    WHERE asset_type = 'learning_material' AND asset_id = $1
    AND recipient_type = CASE WHEN $3 = 'group' THEN 'group' ELSE 'user' END
    AND recipient_id = $2
    AND (
        ($4 = 'owner' AND permission = 'owner')
        OR ($4 = 'write' AND permission IN ('owner', 'write'))
        OR ($4 = 'read')
    )
);

-- name: GetLearningMaterialByID :one
SELECT * FROM learning_materials WHERE id = $1;

-- name: CreateLearningMaterial :one
INSERT INTO learning_materials (id, title, description, material_type, owner_id, source_path, compiled_path, resource_path, file_size, mime_type, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateLearningMaterial :one
UPDATE learning_materials
SET title = $2, description = $3, source_path = $4, compiled_path = $5, resource_path = $6, file_size = $7, mime_type = $8, updated_at = $9
WHERE id = $1
RETURNING *;

-- name: DeleteLearningMaterial :exec
DELETE FROM learning_materials WHERE id = $1;

-- name: GetRecentLearningMaterials :many
SELECT * FROM learning_materials ORDER BY created_at DESC LIMIT $1;

-- name: UpdateLearningMaterialStudentPermission :exec
UPDATE learning_materials SET student_permission = $2 WHERE id = $1;