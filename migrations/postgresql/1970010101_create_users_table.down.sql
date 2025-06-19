-- Drop PostgreSQL specific objects
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at') THEN
        DROP TRIGGER IF EXISTS update_users_updated_at ON users;
    END IF;
    
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP FUNCTION IF EXISTS update_updated_at_column();
    END IF;
END $$;

-- Drop the users table
DROP TABLE IF EXISTS users;