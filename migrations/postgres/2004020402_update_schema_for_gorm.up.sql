-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop all foreign key constraints first
ALTER TABLE form_submissions
    DROP CONSTRAINT IF EXISTS form_submissions_form_uuid_fkey;

ALTER TABLE form_submissions
    DROP CONSTRAINT IF EXISTS form_submissions_user_id_fkey;

ALTER TABLE forms
    DROP CONSTRAINT IF EXISTS forms_user_id_fkey;

ALTER TABLE form_schemas
    DROP CONSTRAINT IF EXISTS form_schemas_form_uuid_fkey;

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
-- First add a new UUID column
ALTER TABLE form_submissions
    ADD COLUMN new_id UUID DEFAULT gen_random_uuid();

-- Update the new_id column with unique values
UPDATE form_submissions
SET new_id = gen_random_uuid();

-- Make the new_id column NOT NULL
ALTER TABLE form_submissions
    ALTER COLUMN new_id SET NOT NULL;

-- Drop the old primary key constraint
ALTER TABLE form_submissions
    DROP CONSTRAINT form_submissions_pkey;

-- Drop the old id column
ALTER TABLE form_submissions
    DROP COLUMN id;

-- Rename new_id to id
ALTER TABLE form_submissions
    RENAME COLUMN new_id TO id;

-- Add primary key constraint
ALTER TABLE form_submissions
    ADD PRIMARY KEY (id);

-- Update other columns
ALTER TABLE form_submissions
    ALTER COLUMN form_uuid TYPE UUID USING form_uuid::uuid,
    ADD COLUMN IF NOT EXISTS user_id BIGINT NOT NULL,
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Update form_schemas table
-- First add a new UUID column
ALTER TABLE form_schemas
    ADD COLUMN new_id UUID DEFAULT gen_random_uuid();

-- Update the new_id column with unique values
UPDATE form_schemas
SET new_id = gen_random_uuid();

-- Make the new_id column NOT NULL
ALTER TABLE form_schemas
    ALTER COLUMN new_id SET NOT NULL;

-- Drop the old primary key constraint
ALTER TABLE form_schemas
    DROP CONSTRAINT form_schemas_pkey;

-- Drop the old id column
ALTER TABLE form_schemas
    DROP COLUMN id;

-- Rename new_id to id
ALTER TABLE form_schemas
    RENAME COLUMN new_id TO id;

-- Add primary key constraint
ALTER TABLE form_schemas
    ADD PRIMARY KEY (id);

-- Update other columns
ALTER TABLE form_schemas
    ALTER COLUMN form_uuid TYPE UUID USING form_uuid::uuid,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Recreate foreign key constraints
ALTER TABLE forms
    ADD CONSTRAINT forms_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE form_submissions
    ADD CONSTRAINT form_submissions_form_uuid_fkey 
    FOREIGN KEY (form_uuid) 
    REFERENCES forms(uuid) 
    ON DELETE CASCADE;

ALTER TABLE form_submissions
    ADD CONSTRAINT form_submissions_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE form_schemas
    ADD CONSTRAINT form_schemas_form_uuid_fkey
    FOREIGN KEY (form_uuid)
    REFERENCES forms(uuid)
    ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_form_submissions_user_id ON form_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_form_submissions_deleted_at ON form_submissions(deleted_at); 