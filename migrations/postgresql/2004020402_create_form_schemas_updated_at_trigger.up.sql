-- Create trigger to automatically update updated_at
CREATE TRIGGER update_form_schemas_updated_at
    BEFORE UPDATE ON form_schemas
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();