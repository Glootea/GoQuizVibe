-- name: CreateAttempt :one
INSERT INTO quiz_attempts (id, user_id, quiz_id, started_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAttemptByID :one
SELECT * FROM quiz_attempts WHERE id = $1;

-- name: UpdateAttempt :one
UPDATE quiz_attempts SET score = $2, max_score = $3, completed_at = $4
WHERE id = $1
RETURNING *;

-- name: GetAttemptsByUser :many
SELECT * FROM quiz_attempts WHERE user_id = $1 ORDER BY started_at DESC;

-- name: CreateUserAnswer :one
INSERT INTO user_answers (id, attempt_id, question_id, user_answer, is_correct)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAnswersByAttempt :many
SELECT * FROM user_answers WHERE attempt_id = $1;

-- name: GetUserStats :one
SELECT
    COALESCE(SUM(q.points), 0) as total_xp,
    COUNT(CASE WHEN ua.is_correct = true THEN 1 END) as correct_cnt,
    COUNT(CASE WHEN ua.is_correct = false THEN 1 END) as wrong_cnt
FROM user_answers ua
JOIN questions q ON q.id = ua.question_id
JOIN quiz_attempts a ON a.id = ua.attempt_id
WHERE a.user_id = $1;

-- name: GetCompletedAttemptBySessionID :one
SELECT a.* FROM quiz_attempts a
JOIN quiz_sessions s ON s.attempt_id = a.id
WHERE s.id = $1 AND a.completed_at IS NOT NULL;

-- name: GetLastActiveDate :one
SELECT MAX(completed_at) FROM quiz_attempts
WHERE user_id = $1 AND completed_at IS NOT NULL;

-- name: GetCompletedAttemptsCount :one
SELECT COUNT(*) FROM quiz_attempts
WHERE user_id = $1 AND completed_at IS NOT NULL;

-- name: GetQuizErrors :many
SELECT a.* FROM quiz_attempts a
WHERE a.user_id = $1 AND a.completed_at IS NOT NULL;

-- name: GetWrongAnswersByAttempt :many
SELECT ua.* FROM user_answers ua
WHERE ua.attempt_id = $1 AND ua.is_correct = false;