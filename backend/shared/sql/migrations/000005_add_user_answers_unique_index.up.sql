DROP INDEX IF EXISTS idx_user_answers_unique;
CREATE UNIQUE INDEX idx_user_answers_unique ON user_answers(attempt_id, question_id);