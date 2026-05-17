-- name: GetLearningMaterials :many
SELECT id, title, description, material_type, owner_id, source_path, compiled_path, resource_path, file_size, mime_type, created_at, updated_at FROM learning_materials ORDER BY created_at DESC;

-- name: GetLearningMaterialsByOwner :many
SELECT id, title, description, material_type, owner_id, source_path, compiled_path, resource_path, file_size, mime_type, created_at, updated_at FROM learning_materials WHERE owner_id = $1 ORDER BY created_at DESC;

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