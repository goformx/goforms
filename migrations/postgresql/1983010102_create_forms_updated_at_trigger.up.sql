-- Create trigger to automatically update updated_at
CREATE TRIGGER update_forms_updated_at
    BEFORE UPDATE ON forms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();