CREATE TABLE question_images (
    id UUID PRIMARY KEY,
    question_id UUID REFERENCES questions(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    order_index INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_question_images_question_id ON question_images(question_id);