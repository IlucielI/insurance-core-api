INSERT INTO products (
    id, name, slug, category, short_description, description, target_customer,
    min_sum_assured, max_sum_assured, min_payment_term, max_payment_term,
    starting_premium, benefits, exclusions, is_featured
) VALUES
(
    'prod_secure_life_plus',
    'Secure Life Plus',
    'secure-life-plus',
    'life',
    'Life insurance for long-term family protection.',
    'Provides financial protection for loved ones if the insured passes away or is diagnosed with a terminal illness.',
    'Young families and main income earners',
    100000000,
    1000000000,
    5,
    20,
    185000,
    '["Death benefit", "Terminal illness benefit", "Optional accidental death rider"]'::jsonb,
    '["Fraudulent claims", "Pre-existing conditions within waiting period"]'::jsonb,
    TRUE
),
(
    'prod_health_guard_essential',
    'Health Guard Essential',
    'health-guard-essential',
    'health',
    'Health coverage for hospital and essential medical costs.',
    'Covers inpatient care, surgery, and selected outpatient benefits for individuals and families.',
    'Individuals and families seeking basic medical protection',
    50000000,
    500000000,
    1,
    10,
    220000,
    '["Hospital room benefit", "Surgery coverage", "Emergency treatment"]'::jsonb,
    '["Cosmetic procedures", "Non-medically necessary treatment"]'::jsonb,
    TRUE
),
(
    'prod_auto_shield_comprehensive',
    'Auto Shield Comprehensive',
    'auto-shield-comprehensive',
    'vehicle',
    'Comprehensive vehicle protection for car owners.',
    'Protects vehicles from accidental damage, theft, natural disaster, and selected third-party liabilities.',
    'Private car owners',
    75000000,
    750000000,
    1,
    5,
    95000,
    '["Accidental damage coverage", "Theft protection", "Third-party liability option"]'::jsonb,
    '["Driving without a valid license", "Commercial use without declaration"]'::jsonb,
    TRUE
)
ON CONFLICT (id) DO NOTHING;
