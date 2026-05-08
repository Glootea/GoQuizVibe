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
SELECT * FROM quiz_attempts WHERE user_id = $1 AND completed_at IS NOT NULL ORDER BY started_at DESC;

-- name: CreateUserAnswer :one
INSERT INTO user_answers (id, attempt_id, question_id, user_answer, is_correct)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAnswersByAttempt :many
SELECT * FROM user_answers WHERE attempt_id = $1;

-- name: GetUserStats :one
SELECT
    COALESCE((SELECT SUM(score) FROM quiz_attempts qa WHERE qa.user_id = $1 AND qa.completed_at IS NOT NULL), 0) as total_xp,
    (SELECT COUNT(*) FROM user_answers ua JOIN quiz_attempts a ON a.id = ua.attempt_id WHERE a.user_id = $1 AND a.completed_at IS NOT NULL AND ua.is_correct = true) as correct_cnt,
    (SELECT COUNT(*) FROM user_answers ua JOIN quiz_attempts a ON a.id = ua.attempt_id WHERE a.user_id = $1 AND a.completed_at IS NOT NULL AND ua.is_correct = false) as wrong_cnt;

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

-- name: GetRecentAttempts :many
SELECT a.id, a.user_id, a.quiz_id, a.score, a.max_score, a.started_at, a.completed_at,
       u.name as user_name, q.title as quiz_title
FROM quiz_attempts a
JOIN users u ON u.id = a.user_id
JOIN quizzes q ON q.id = a.quiz_id
WHERE a.completed_at IS NOT NULL
ORDER BY a.completed_at DESC LIMIT $1;

-- name: GetAllAttempts :many
SELECT a.id, a.user_id, a.quiz_id, a.score, a.max_score, a.started_at, a.completed_at,
       u.name as user_name, q.title as quiz_title
FROM quiz_attempts a
JOIN users u ON u.id = a.user_id
JOIN quizzes q ON q.id = a.quiz_id
WHERE a.completed_at IS NOT NULL
ORDER BY a.completed_at DESC;

-- name: GetAttemptsByQuiz :many
SELECT a.id, a.user_id, a.quiz_id, a.score, a.max_score, a.started_at, a.completed_at,
       u.name as user_name
FROM quiz_attempts a
JOIN users u ON u.id = a.user_id
WHERE a.quiz_id = $1 AND a.completed_at IS NOT NULL
ORDER BY a.completed_at DESC;

-- name: GetQuizStats :many
SELECT
    q.id as quiz_id, q.title, q.subject,
    COUNT(a.id) as attempt_count,
    COALESCE(AVG(a.score * 100.0 / NULLIF(a.max_score, 0)), 0) as avg_score,
    CASE WHEN COUNT(a.id) > 0
         THEN (COUNT(*) FILTER (WHERE a.score * 100.0 / NULLIF(a.max_score, 0) >= 60))::float / COUNT(*) * 100
         ELSE 0
    END as pass_rate
FROM quizzes q
LEFT JOIN quiz_attempts a ON a.quiz_id = q.id AND a.completed_at IS NOT NULL AND a.max_score > 0
WHERE q.status != 'archived'
GROUP BY q.id, q.title, q.subject
ORDER BY q.created_at DESC;

-- name: GetGradeDistribution :one
SELECT json_object_agg(grade_level, count) as grade_dist FROM (
    SELECT COALESCE(grade::text, 'unknown') as grade_level, COUNT(*) as count 
    FROM quizzes WHERE status != 'archived' AND grade IS NOT NULL 
    GROUP BY grade
) sub;

-- name: GetSubjectDistribution :one
SELECT json_object_agg(subject_name, count) as subject_dist FROM (
    SELECT subject as subject_name, COUNT(*) as count FROM quizzes WHERE status != 'archived' AND subject IS NOT NULL GROUP BY subject
) sub;