-- ENUMs
DO $$ BEGIN
    CREATE TYPE group_role AS ENUM ('admin', 'member');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE permission_type AS ENUM ('read', 'write', 'owner');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE recipient_type AS ENUM ('user', 'group');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE asset_type AS ENUM ('quiz', 'learning_material');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Tables
CREATE TABLE IF NOT EXISTS user_groups (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    owner_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id UUID REFERENCES user_groups(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role group_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS asset_permissions (
    id UUID PRIMARY KEY,
    asset_type asset_type NOT NULL,
    asset_id UUID NOT NULL,
    permission permission_type NOT NULL,
    recipient_type recipient_type NOT NULL,
    recipient_id UUID NOT NULL,
    grantor_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(asset_type, asset_id, recipient_type, recipient_id, permission)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_user_groups_owner_id ON user_groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_asset_permissions_recipient ON asset_permissions(recipient_type, recipient_id);
CREATE INDEX IF NOT EXISTS idx_asset_permissions_asset ON asset_permissions(asset_type, asset_id);
