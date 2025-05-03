ALTER TABLE contact_submissions
    ADD INDEX idx_contact_submissions_email (email),
    ADD INDEX idx_contact_submissions_status (status); 