CREATE TABLE IF NOT EXISTS application_review_checks (
    id VARCHAR(64) PRIMARY KEY,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    check_type VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    notes VARCHAR(500),
    reviewed_by VARCHAR(120),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (application_id, check_type)
);

CREATE INDEX IF NOT EXISTS idx_application_review_checks_application_id ON application_review_checks(application_id);
CREATE INDEX IF NOT EXISTS idx_application_review_checks_status ON application_review_checks(status);
