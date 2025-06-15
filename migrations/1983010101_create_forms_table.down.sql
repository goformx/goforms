-- Drop PostgreSQL specific objects
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_forms_updated_at') THEN
        DROP TRIGGER IF EXISTS update_forms_updated_at ON forms;
    END IF;
END $$;

-- Drop the forms table
DROP TABLE IF EXISTS forms;