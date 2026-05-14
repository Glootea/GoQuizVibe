ALTER TABLE user_answers ADD COLUMN id UUID DEFAULT uuid_generate_v4();
ALTER TABLE user_answers ADD PRIMARY KEY (id);