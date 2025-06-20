-- Remove CORS settings from forms table
ALTER TABLE forms
DROP COLUMN cors_origins,
DROP COLUMN cors_methods,
DROP COLUMN cors_headers;