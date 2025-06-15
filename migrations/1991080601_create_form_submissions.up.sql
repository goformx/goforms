-- Create form submissions table
CREATE TABLE IF NOT EXISTS form_submissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    form_uuid VARCHAR(36) NOT NULL,
    data JSON NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_uuid) REFERENCES forms (uuid) ON DELETE CASCADE
);

-- Create index on form_uuid
CREATE INDEX IF NOT EXISTS idx_form_submissions_form_uuid ON form_submissions (form_uuid);

-- PostgreSQL specific: Create trigger to automatically update updated_at
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_form_submissions_updated_at') THEN
        CREATE TRIGGER update_form_submissions_updated_at
            BEFORE UPDATE ON form_submissions
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();

END IF;

END $$;

-- MariaDB specific: Add ON UPDATE CURRENT_TIMESTAMP to updated_at
ALTER TABLE form_submissions
MODIFY COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;