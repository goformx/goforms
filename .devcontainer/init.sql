-- Create test database
CREATE DATABASE IF NOT EXISTS goforms_test;

-- Grant privileges to application user
-- Using a direct reference since environment variables aren't expanded in .sql files
GRANT ALL PRIVILEGES ON goforms_test.* TO 'goforms'@'%';

-- Create test-specific user
CREATE USER IF NOT EXISTS 'goforms_test'@'%' IDENTIFIED BY 'goforms_test';
GRANT ALL PRIVILEGES ON goforms_test.* TO 'goforms_test'@'%';

-- Make sure privileges are applied
FLUSH PRIVILEGES; 

-- Set sensible defaults
SET GLOBAL innodb_strict_mode=ON;
SET GLOBAL time_zone='+00:00';

-- Additional settings for development
SET GLOBAL max_connections=1000;
SET GLOBAL connect_timeout=30; 