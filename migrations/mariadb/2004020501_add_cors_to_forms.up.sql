-- Add CORS settings to forms table
ALTER TABLE forms
ADD COLUMN cors_origins JSON DEFAULT('[]'),
ADD COLUMN cors_methods JSON DEFAULT('["GET", "POST", "OPTIONS"]'),
ADD COLUMN cors_headers JSON DEFAULT(
    '["Content-Type", "Accept", "Origin"]'
);

-- Update existing forms to have default CORS settings
UPDATE forms
SET
    cors_origins = '[]',
    cors_methods = '["GET", "POST", "OPTIONS"]',
    cors_headers = '["Content-Type", "Accept", "Origin"]'
WHERE
    cors_origins IS NULL;