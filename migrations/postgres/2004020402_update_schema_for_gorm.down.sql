-- Revert form_submissions table changes
ALTER TABLE form_submissions
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS user_id,
    ALTER COLUMN id TYPE BIGSERIAL USING id::bigint,
    ALTER COLUMN form_uuid TYPE VARCHAR(36) USING form_uuid::varchar,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

DROP INDEX IF EXISTS idx_form_submissions_deleted_at;
DROP INDEX IF EXISTS idx_form_submissions_user_id;

-- Revert forms table changes
ALTER TABLE forms
    ALTER COLUMN uuid TYPE VARCHAR(36) USING uuid::varchar,
    DROP COLUMN IF EXISTS deleted_at,
    ALTER COLUMN title TYPE VARCHAR(255),
    ALTER COLUMN description TYPE TEXT,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

DROP INDEX IF EXISTS idx_forms_deleted_at;

-- Revert users table changes
ALTER TABLE users
    DROP COLUMN IF EXISTS deleted_at,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

DROP INDEX IF EXISTS idx_users_deleted_at; 