-- name: GetQuestionsByQuizID :many
SELECT * FROM questions WHERE quiz_id = $1 ORDER BY order_index ASC;

-- name: CreateQuestion :one
INSERT INTO questions (id, quiz_id, text, type, options, correct_answer, explanation, points, order_index)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: UpdateQuestion :one
UPDATE questions SET text = $2, type = $3, options = $4, correct_answer = $5, explanation = $6, points = $7, order_index = $8
WHERE id = $1 RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;