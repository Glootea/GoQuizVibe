-- name: GetQuizzes :many
SELECT * FROM quizzes ORDER BY created_at DESC;

-- name: GetAvailableQuizzes :many
SELECT * FROM quizzes WHERE status = 'available' ORDER BY created_at DESC;

-- name: GetQuizByID :one
SELECT * FROM quizzes WHERE id = $1;

-- name: GetQuizzesForUser :many
SELECT DISTINCT q.* FROM quizzes q
WHERE EXISTS (
    SELECT 1 FROM asset_permissions ap
    WHERE ap.asset_type = 'quiz' AND ap.asset_id = q.id
    AND (
        (ap.recipient_type = 'user' AND ap.recipient_id = $1)
        OR (ap.recipient_type = 'group' AND ap.recipient_id = ANY($2::uuid[]))
    )
    AND ap.permission IN ('read', 'write', 'owner')
)
ORDER BY q.created_at DESC;

-- name: HasQuizAccess :one
SELECT EXISTS(
    SELECT 1 FROM asset_permissions
    WHERE asset_type = 'quiz' AND asset_id = $1
    AND recipient_type = CASE WHEN $3 = 'group' THEN 'group' ELSE 'user' END
    AND recipient_id = $2
    AND (
        ($4 = 'owner' AND permission = 'owner')
        OR ($4 = 'write' AND permission IN ('owner', 'write'))
        OR ($4 = 'read')
    )
);

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

-- name: QuizTitleExists :one
SELECT EXISTS(SELECT 1 FROM quizzes WHERE title = $1);