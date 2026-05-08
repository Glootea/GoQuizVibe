-- name: GetImagesByQuestionID :many
SELECT * FROM question_images WHERE question_id = $1 ORDER BY order_index ASC;

-- name: GetQuestionImageByID :one
SELECT * FROM question_images WHERE id = $1;

-- name: CreateQuestionImage :one
INSERT INTO question_images (id, question_id, url, order_index, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteQuestionImage :exec
DELETE FROM question_images WHERE id = $1;

-- name: DeleteImagesByQuestionID :exec
DELETE FROM question_images WHERE question_id = $1;

-- name: GetImageCountByQuestionID :one
SELECT COUNT(*) FROM question_images WHERE question_id = $1;
