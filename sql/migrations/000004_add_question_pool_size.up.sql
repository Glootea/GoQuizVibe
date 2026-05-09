-- +migrate Up
ALTER TABLE quizzes ADD COLUMN question_pool_size INT NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE quizzes DROP COLUMN question_pool_size;