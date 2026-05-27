DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'quiz_status') THEN
        IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumtypid = 'quiz_status'::regtype AND enumlabel = 'archived') THEN
            ALTER TYPE quiz_status ADD VALUE 'archived';
        END IF;
    ELSE
        CREATE TYPE quiz_status AS ENUM ('available', 'assigned', 'completed', 'archived');
    END IF;
END
$$;