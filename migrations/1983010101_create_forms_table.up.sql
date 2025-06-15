-- Create forms table
CREATE TABLE IF NOT EXISTS forms (
    uuid VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSON NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users (uuid) ON DELETE CASCADE
);

-- Create index on user_id
CREATE INDEX IF NOT EXISTS idx_forms_user_id ON forms (user_id);

-- PostgreSQL specific: Create trigger to automatically update updated_at
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_forms_updated_at') THEN
        CREATE TRIGGER update_forms_updated_at
            BEFORE UPDATE ON forms
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();

END IF;

END $$;

-- MariaDB specific: Add ON UPDATE CURRENT_TIMESTAMP to updated_at
ALTER TABLE forms
MODIFY COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;