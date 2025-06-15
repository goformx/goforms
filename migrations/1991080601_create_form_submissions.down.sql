-- Drop PostgreSQL specific objects
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_form_submissions_updated_at') THEN
        DROP TRIGGER IF EXISTS update_form_submissions_updated_at ON form_submissions;
    END IF;
END $$;

-- Drop the form submissions table
DROP TABLE IF EXISTS form_submissions;