-- name: GetQuestionsByQuizID :many
SELECT * FROM questions WHERE quiz_id = $1 ORDER BY order_index ASC;

-- name: CreateQuestion :one
INSERT INTO questions (id, quiz_id, text, type, options, correct_answer, explanation, points, order_index)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;