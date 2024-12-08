-- Create database first
CREATE DATABASE IF NOT EXISTS goforms_test;

-- Create user if not exists and set password
CREATE USER IF NOT EXISTS 'goforms_test'@'%' IDENTIFIED BY 'goforms_test';

-- Grant privileges
GRANT ALL PRIVILEGES ON goforms_test.* TO 'goforms_test'@'%';

-- Make sure privileges are applied
FLUSH PRIVILEGES; 