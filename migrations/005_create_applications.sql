CREATE TABLE IF NOT EXISTS applications (
    id VARCHAR(64) PRIMARY KEY,
    product_id VARCHAR(64) NOT NULL REFERENCES products(id),
    full_name VARCHAR(120) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(32) NOT NULL,
    age INTEGER NOT NULL CHECK (age BETWEEN 18 AND 60),
    gender VARCHAR(16) NOT NULL,
    sum_assured BIGINT NOT NULL CHECK (sum_assured > 0),
    payment_term INTEGER NOT NULL CHECK (payment_term > 0),
    payment_frequency VARCHAR(16) NOT NULL,
    smoker VARCHAR(8) NOT NULL,
    occupation_class VARCHAR(16) NOT NULL,
    health_risk VARCHAR(16) NOT NULL,
    premium BIGINT NOT NULL CHECK (premium >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_applications_product_id ON applications(product_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
