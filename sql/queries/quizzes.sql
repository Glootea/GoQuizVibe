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