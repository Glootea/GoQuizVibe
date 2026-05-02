-- ENUMs
CREATE TYPE role AS ENUM ('teacher', 'student');
CREATE TYPE quiz_status AS ENUM ('available', 'assigned', 'completed', 'archived');
CREATE TYPE question_type AS ENUM ('choice', 'open', 'fill');

-- Tables
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE quizzes (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    subject TEXT,
    grade INT,
    status quiz_status NOT NULL DEFAULT 'available',
    time_limit INT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE questions (
    id UUID PRIMARY KEY,
    quiz_id UUID REFERENCES quizzes(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    type question_type NOT NULL,
    options JSONB,
    correct_answer TEXT NOT NULL,
    explanation TEXT,
    points INT DEFAULT 10,
    order_index INT DEFAULT 0
);

CREATE TABLE quiz_attempts (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    quiz_id UUID REFERENCES quizzes(id),
    score INT DEFAULT 0,
    max_score INT DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE user_answers (
    id UUID PRIMARY KEY,
    attempt_id UUID REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id UUID REFERENCES questions(id),
    user_answer TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT false
);

CREATE TABLE quiz_sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    quiz_id UUID REFERENCES quizzes(id),
    attempt_id UUID REFERENCES quiz_attempts(id),
    current_index INT DEFAULT 0,
    answers JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_questions_quiz_id ON questions(quiz_id);
CREATE INDEX idx_quiz_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_user_answers_attempt_id ON user_answers(attempt_id);
CREATE INDEX idx_user_answers_question_id ON user_answers(question_id);
CREATE INDEX idx_quiz_sessions_user_id ON quiz_sessions(user_id);
CREATE INDEX idx_quiz_sessions_quiz_id ON quiz_sessions(quiz_id);
CREATE INDEX idx_quiz_sessions_attempt_id ON quiz_sessions(attempt_id);