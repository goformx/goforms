-- Drop all foreign key constraints first
ALTER TABLE form_submissions
    DROP CONSTRAINT IF EXISTS form_submissions_form_uuid_fkey;

ALTER TABLE form_submissions
    DROP CONSTRAINT IF EXISTS form_submissions_user_id_fkey;

ALTER TABLE forms
    DROP CONSTRAINT IF EXISTS forms_user_id_fkey;

ALTER TABLE form_schemas
    DROP CONSTRAINT IF EXISTS form_schemas_form_uuid_fkey;

-- Revert form_submissions table changes
-- First add a new bigint column
ALTER TABLE form_submissions
    ADD COLUMN new_id BIGSERIAL;

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
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS user_id,
    ALTER COLUMN form_uuid TYPE VARCHAR(36) USING form_uuid::varchar,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Revert forms table changes
ALTER TABLE forms
    ALTER COLUMN uuid TYPE VARCHAR(36) USING uuid::varchar,
    DROP COLUMN IF EXISTS deleted_at,
    ALTER COLUMN title TYPE VARCHAR(255),
    ALTER COLUMN description TYPE TEXT,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Revert form_schemas table changes
-- First add a new bigint column
ALTER TABLE form_schemas
    ADD COLUMN new_id BIGSERIAL;

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
    ALTER COLUMN form_uuid TYPE VARCHAR(36) USING form_uuid::varchar,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Revert users table changes
ALTER TABLE users
    DROP COLUMN IF EXISTS deleted_at,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Recreate original foreign key constraints
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

ALTER TABLE form_schemas
    ADD CONSTRAINT form_schemas_form_uuid_fkey
    FOREIGN KEY (form_uuid)
    REFERENCES forms(uuid)
    ON DELETE CASCADE;

-- Drop indexes
DROP INDEX IF EXISTS idx_form_submissions_deleted_at;
DROP INDEX IF EXISTS idx_form_submissions_user_id;
DROP INDEX IF EXISTS idx_forms_deleted_at;
DROP INDEX IF EXISTS idx_users_deleted_at; 