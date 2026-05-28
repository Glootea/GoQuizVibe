-- Student Permission System
-- Enables visibility control for students: open_to_all, assigned, private

DO $$ BEGIN
    CREATE TYPE student_permission AS ENUM ('open_to_all', 'assigned', 'private');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TABLE quizzes ADD COLUMN student_permission student_permission NOT NULL DEFAULT 'private';
ALTER TABLE learning_materials ADD COLUMN student_permission student_permission NOT NULL DEFAULT 'private';

CREATE TABLE IF NOT EXISTS student_access (
    id UUID PRIMARY KEY,
    asset_type asset_type NOT NULL,
    asset_id UUID NOT NULL,
    recipient_type recipient_type NOT NULL,
    recipient_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(asset_type, asset_id, recipient_type, recipient_id)
);

CREATE INDEX IF NOT EXISTS idx_student_access_asset ON student_access(asset_type, asset_id);
CREATE INDEX IF NOT EXISTS idx_student_access_recipient ON student_access(recipient_type, recipient_id);
