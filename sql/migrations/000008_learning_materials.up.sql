-- ENUMs
CREATE TYPE learning_material_type AS ENUM ('typst', 'resource');

-- Tables
CREATE TABLE learning_materials (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    material_type learning_material_type NOT NULL,
    owner_id UUID REFERENCES users(id),
    source_path TEXT,
    compiled_path TEXT,
    resource_path TEXT,
    file_size BIGINT,
    mime_type TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_learning_materials_owner_id ON learning_materials(owner_id);
CREATE INDEX idx_learning_materials_material_type ON learning_materials(material_type);