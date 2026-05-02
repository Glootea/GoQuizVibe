-- name: GetQuizzes :many
SELECT * FROM quizzes ORDER BY created_at DESC;

-- name: GetAvailableQuizzes :many
SELECT * FROM quizzes WHERE status = 'available' ORDER BY created_at DESC;

-- name: GetQuizByID :one
SELECT * FROM quizzes WHERE id = $1;

-- name: GetQuizzesForUser :many
SELECT * FROM quizzes WHERE status = 'available' OR created_by = $1 ORDER BY created_at DESC;

-- name: CreateQuiz :one
INSERT INTO quizzes (id, title, description, subject, grade, status, time_limit, created_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetNonArchivedQuizzes :many
SELECT * FROM quizzes WHERE status != 'archived' ORDER BY created_at DESC;

-- name: UpdateQuiz :one
UPDATE quizzes SET title = $2, description = $3, subject = $4, grade = $5, status = $6, time_limit = $7
WHERE id = $1 RETURNING *;

-- name: DeleteQuiz :exec
UPDATE quizzes SET status = 'archived' WHERE id = $1;

-- name: UpdateQuizStatus :exec
UPDATE quizzes SET status = $2 WHERE id = $1;