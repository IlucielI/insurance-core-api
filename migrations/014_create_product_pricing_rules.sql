-- 014_create_product_pricing_rules.sql
-- Creates dynamic product pricing rules table and links questions via foreign key.

CREATE TABLE IF NOT EXISTS product_pricing_rules (
    id VARCHAR(64) PRIMARY KEY,
    product_id VARCHAR(64) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    rule_code VARCHAR(64) NOT NULL,
    rule_name VARCHAR(120) NOT NULL,
    rule_type VARCHAR(32) NOT NULL, -- 'base_rate', 'bracket', 'multiplier_map', 'term_factor', 'frequency_loading'
    factors JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_pricing_rules_product_id ON product_pricing_rules(product_id);
CREATE INDEX IF NOT EXISTS idx_product_pricing_rules_code ON product_pricing_rules(product_id, rule_code);

-- Link underwriting questions to product pricing rules
ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS pricing_rule_id VARCHAR(64) REFERENCES product_pricing_rules(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_questions_pricing_rule_id ON questions(pricing_rule_id);
