-- Create form_schemas table
CREATE TABLE IF NOT EXISTS form_schemas (
    uuid VARCHAR(36) PRIMARY KEY,
    form_id VARCHAR(36) NOT NULL,
    schema JSON NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (form_id) REFERENCES forms (uuid) ON DELETE CASCADE
);

-- Create index on form_id
CREATE INDEX IF NOT EXISTS idx_form_schemas_form_id ON form_schemas (form_id);