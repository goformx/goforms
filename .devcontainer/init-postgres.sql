-- Create databases
CREATE DATABASE goforms;
CREATE DATABASE goforms_test;

-- Create application user with password
CREATE USER goforms WITH PASSWORD 'goforms';

-- Grant privileges to application user
GRANT ALL PRIVILEGES ON DATABASE goforms TO goforms;
GRANT ALL PRIVILEGES ON DATABASE goforms_test TO goforms;

-- Create test-specific user
CREATE USER goforms_test WITH PASSWORD 'goforms_test';
GRANT ALL PRIVILEGES ON DATABASE goforms_test TO goforms_test;

-- Connect to goforms database to set up schemas
\c goforms;

-- Create schema and grant privileges
CREATE SCHEMA IF NOT EXISTS public;
GRANT ALL ON SCHEMA public TO goforms;
GRANT ALL ON ALL TABLES IN SCHEMA public TO goforms;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO goforms;

-- Set up default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO goforms;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO goforms;

-- Connect to test database to set up schemas
\c goforms_test;

-- Create schema and grant privileges for test database
CREATE SCHEMA IF NOT EXISTS public;
GRANT ALL ON SCHEMA public TO goforms;
GRANT ALL ON SCHEMA public TO goforms_test;
GRANT ALL ON ALL TABLES IN SCHEMA public TO goforms;
GRANT ALL ON ALL TABLES IN SCHEMA public TO goforms_test;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO goforms;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO goforms_test;

-- Set up default privileges for future objects in test database
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO goforms;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO goforms_test;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO goforms;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO goforms_test;

-- Set timezone
SET timezone = 'UTC';

-- Set connection limits
ALTER SYSTEM SET max_connections = '1000';
ALTER SYSTEM SET statement_timeout = '30s'; 