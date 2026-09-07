CREATE TABLE IF NOT EXISTS questionnaires (
    id VARCHAR(64) PRIMARY KEY,
    product_id VARCHAR(64) REFERENCES products(id) ON DELETE SET NULL,
    category VARCHAR(32) NOT NULL DEFAULT 'all',
    title VARCHAR(120) NOT NULL,
    description TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_questionnaires_product_id ON questionnaires(product_id);
CREATE INDEX IF NOT EXISTS idx_questionnaires_category ON questionnaires(category);

CREATE TABLE IF NOT EXISTS questions (
    id VARCHAR(64) PRIMARY KEY,
    questionnaire_id VARCHAR(64) NOT NULL REFERENCES questionnaires(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    pillar_type VARCHAR(32) NOT NULL,
    code VARCHAR(64) NOT NULL,
    label TEXT NOT NULL,
    help_text TEXT,
    input_type VARCHAR(32) NOT NULL,
    placeholder VARCHAR(120),
    order_index INTEGER NOT NULL DEFAULT 0,
    validation_rules JSONB DEFAULT '{}'::jsonb,
    options JSONB DEFAULT '[]'::jsonb,
    parent_question_id VARCHAR(64) REFERENCES questions(id) ON DELETE SET NULL,
    show_if_parent_value JSONB,
    affects_pricing_field VARCHAR(64),
    underwriting_rules JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_questions_lookup ON questions(questionnaire_id, step_number, order_index);
CREATE INDEX IF NOT EXISTS idx_questions_parent ON questions(parent_question_id);

CREATE TABLE IF NOT EXISTS application_answers (
    id VARCHAR(64) PRIMARY KEY,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    question_id VARCHAR(64) NOT NULL REFERENCES questions(id) ON DELETE RESTRICT,
    code VARCHAR(64) NOT NULL,
    answer_value JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_application_answers_app_id ON application_answers(application_id);
CREATE INDEX IF NOT EXISTS idx_application_answers_question_id ON application_answers(question_id);
