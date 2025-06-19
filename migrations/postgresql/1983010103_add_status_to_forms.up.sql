-- Add status column to forms table
ALTER TABLE forms
ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'draft';