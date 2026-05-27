-- up: Add unique constraint on user_answers for ON CONFLICT support
DROP INDEX IF EXISTS idx_user_answers_unique;
CREATE UNIQUE INDEX idx_user_answers_unique ON user_answers(attempt_id, question_id);

-- down: Drop the unique index
DROP INDEX IF EXISTS idx_user_answers_unique;