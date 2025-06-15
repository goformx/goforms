-- Create form schemas table
CREATE TABLE IF NOT EXISTS form_schemas (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    form_uuid VARCHAR(36) NOT NULL,
    version INTEGER NOT NULL,
    schema JSON NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_uuid) REFERENCES forms (uuid) ON DELETE CASCADE,
    UNIQUE (form_uuid, version)
);

-- Create index on form_uuid
CREATE INDEX IF NOT EXISTS idx_form_schemas_form_uuid ON form_schemas (form_uuid);

-- PostgreSQL specific: Create trigger to automatically update updated_at
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_form_schemas_updated_at') THEN
        CREATE TRIGGER update_form_schemas_updated_at
            BEFORE UPDATE ON form_schemas
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();

END IF;

END $$;

-- MariaDB specific: Add ON UPDATE CURRENT_TIMESTAMP to updated_at
ALTER TABLE form_schemas
MODIFY COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;