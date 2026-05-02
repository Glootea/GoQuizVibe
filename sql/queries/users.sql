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