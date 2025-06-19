-- Create form_submissions table
CREATE TABLE IF NOT EXISTS form_submissions (
    uuid VARCHAR(36) PRIMARY KEY,
    form_id VARCHAR(36) NOT NULL,
    data JSON NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (form_id) REFERENCES forms (uuid) ON DELETE CASCADE
);

-- Create index on form_id
CREATE INDEX IF NOT EXISTS idx_form_submissions_form_id ON form_submissions (form_id);