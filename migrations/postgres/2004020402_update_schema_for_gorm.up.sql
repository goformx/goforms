-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Update users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Update forms table
ALTER TABLE forms
    ALTER COLUMN uuid TYPE UUID USING uuid::uuid,
    ALTER COLUMN uuid SET DEFAULT gen_random_uuid(),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN title TYPE VARCHAR(100),
    ALTER COLUMN description TYPE VARCHAR(500);

CREATE INDEX IF NOT EXISTS idx_forms_deleted_at ON forms(deleted_at);

-- Update form_submissions table
ALTER TABLE form_submissions
    ALTER COLUMN id TYPE UUID USING id::uuid,
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN form_uuid TYPE UUID USING form_uuid::uuid,
    ADD COLUMN IF NOT EXISTS user_id BIGINT NOT NULL REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_form_submissions_user_id ON form_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_form_submissions_deleted_at ON form_submissions(deleted_at); 