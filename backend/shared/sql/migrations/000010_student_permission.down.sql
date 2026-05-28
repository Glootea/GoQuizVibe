DROP INDEX IF EXISTS idx_student_access_recipient;
DROP INDEX IF EXISTS idx_student_access_asset;
DROP TABLE IF EXISTS student_access;
ALTER TABLE learning_materials DROP COLUMN IF EXISTS student_permission;
ALTER TABLE quizzes DROP COLUMN IF EXISTS student_permission;
DROP TYPE IF EXISTS student_permission;
