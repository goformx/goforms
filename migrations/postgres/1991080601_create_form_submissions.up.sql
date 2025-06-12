CREATE TABLE IF NOT EXISTS form_submissions (
    id BIGSERIAL PRIMARY KEY,
    form_uuid UUID NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_uuid) REFERENCES forms(uuid) ON DELETE CASCADE
);

-- Create index on form_uuid
CREATE INDEX IF NOT EXISTS idx_form_submissions_form_uuid ON form_submissions(form_uuid);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_form_submissions_updated_at
    BEFORE UPDATE ON form_submissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 