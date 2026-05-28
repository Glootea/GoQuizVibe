-- name: GetLearningMaterials :many
SELECT id, title, description, material_type, owner_id, source_path, compiled_path, resource_path, file_size, mime_type, created_at, updated_at FROM learning_materials ORDER BY created_at DESC;

-- name: GetLearningMaterialsForUser :many
SELECT DISTINCT lm.id, lm.title, lm.description, lm.material_type, lm.owner_id, lm.source_path, lm.compiled_path, lm.resource_path, lm.file_size, lm.mime_type, lm.created_at, lm.updated_at FROM learning_materials lm
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
SELECT id, title, description, material_type, owner_id, source_path, compiled_path, resource_path, file_size, mime_type, created_at, updated_at FROM learning_materials WHERE id = $1;

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
SELECT id, title, description, material_type, owner_id, source_path, compiled_path, resource_path, file_size, mime_type, created_at, updated_at FROM learning_materials ORDER BY created_at DESC LIMIT $1;