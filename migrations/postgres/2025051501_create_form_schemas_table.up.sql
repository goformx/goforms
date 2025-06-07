CREATE TABLE IF NOT EXISTS form_schemas (
    id BIGSERIAL PRIMARY KEY,
    form_uuid VARCHAR(36) NOT NULL,
    version INTEGER NOT NULL,
    schema JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_uuid) REFERENCES forms(uuid) ON DELETE CASCADE,
    UNIQUE (form_uuid, version)
);

-- Create index on form_uuid
CREATE INDEX IF NOT EXISTS idx_form_schemas_form_uuid ON form_schemas(form_uuid);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_form_schemas_updated_at
    BEFORE UPDATE ON form_schemas
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 