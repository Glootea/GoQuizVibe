-- name: CreateSession :one
INSERT INTO quiz_sessions (id, user_id, quiz_id, attempt_id, current_index, answers, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSession :one
SELECT * FROM quiz_sessions WHERE id = $1;

-- name: SessionExists :one
SELECT EXISTS(SELECT 1 FROM quiz_sessions WHERE id = $1);

-- name: GetSessionByAttemptID :one
SELECT * FROM quiz_sessions WHERE attempt_id = $1;

-- name: UpdateSession :one
UPDATE quiz_sessions SET current_index = $2, answers = $3
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM quiz_sessions WHERE id = $1;