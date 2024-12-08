-- Create test database
CREATE DATABASE IF NOT EXISTS goforms_test;

-- Grant privileges to application user for test database
GRANT ALL PRIVILEGES ON goforms_test.* TO '${MARIADB_USER}'@'%';

-- Create test-specific user (optional, can use same user as main app)
CREATE USER IF NOT EXISTS 'goforms_test'@'%' IDENTIFIED BY 'goforms_test';
GRANT ALL PRIVILEGES ON goforms_test.* TO 'goforms_test'@'%';

-- Make sure privileges are applied
FLUSH PRIVILEGES; 