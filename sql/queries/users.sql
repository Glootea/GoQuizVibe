-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, role, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: GetStudentCount :one
select COUNT(*) from users where role = 'student';

-- name: GetAdminStatsData :one
SELECT
    (SELECT COUNT(*) FROM quizzes WHERE status != 'archived') as total_quizzes,
    (SELECT COUNT(*) FROM users WHERE role = 'student') as total_students,
    (SELECT COUNT(*) FROM quiz_attempts WHERE completed_at IS NOT NULL) as total_attempts,
    COALESCE((SELECT AVG(score * 100.0 / NULLIF(max_score, 0))
              FROM quiz_attempts WHERE completed_at IS NOT NULL AND max_score > 0), 0) as avg_score;