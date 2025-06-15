-- Create databases
CREATE DATABASE IF NOT EXISTS goforms;
CREATE DATABASE IF NOT EXISTS goforms_test;

-- Create application user with password
CREATE USER IF NOT EXISTS 'goforms'@'%' IDENTIFIED BY 'goforms';
GRANT ALL PRIVILEGES ON goforms.* TO 'goforms'@'%';
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
