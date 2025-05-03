CREATE TABLE IF NOT EXISTS form_submissions (
    id VARCHAR(255) PRIMARY KEY,
    form_id BIGINT NOT NULL,
    data JSON NOT NULL,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    metadata JSON NOT NULL,
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE,
    INDEX idx_form_submissions_form_id (form_id),
    INDEX idx_form_submissions_submitted_at (submitted_at),
    INDEX idx_form_submissions_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 