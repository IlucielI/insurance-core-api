INSERT INTO products (
    id, name, slug, category, short_description, description, target_customer,
    min_sum_assured, max_sum_assured, min_payment_term, max_payment_term,
    starting_premium, pricing_rules, benefits, exclusions, is_featured
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
    '{"base_rate":0.0035,"age_factors":[{"min_age":18,"max_age":30,"factor":1.0},{"min_age":31,"max_age":40,"factor":1.25},{"min_age":41,"max_age":50,"factor":1.75},{"min_age":51,"max_age":60,"factor":2.5}],"gender_factors":{"male":1.05,"female":1.0},"smoker_factors":{"yes":1.35,"no":1.0},"occupation_factors":{"low":0.95,"standard":1.0,"high":1.4},"health_factors":{"low":1.0,"medium":1.25,"high":1.75},"frequency_loading":{"annual":1.0,"semi_annual":1.02,"quarterly":1.035,"monthly":1.06}}'::jsonb,
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
    '{"base_rate":0.0042,"age_factors":[{"min_age":18,"max_age":30,"factor":1.0},{"min_age":31,"max_age":40,"factor":1.2},{"min_age":41,"max_age":50,"factor":1.6},{"min_age":51,"max_age":60,"factor":2.2}],"gender_factors":{"male":1.03,"female":1.0},"smoker_factors":{"yes":1.25,"no":1.0},"occupation_factors":{"low":0.95,"standard":1.0,"high":1.25},"health_factors":{"low":1.0,"medium":1.3,"high":1.9},"frequency_loading":{"annual":1.0,"semi_annual":1.02,"quarterly":1.035,"monthly":1.06}}'::jsonb,
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
    '{"base_rate":0.012,"age_factors":[{"min_age":18,"max_age":25,"factor":1.25},{"min_age":26,"max_age":45,"factor":1.0},{"min_age":46,"max_age":60,"factor":1.1}],"gender_factors":{"male":1.0,"female":1.0},"smoker_factors":{"yes":1.0,"no":1.0},"occupation_factors":{"low":0.95,"standard":1.0,"high":1.15},"health_factors":{"low":1.0,"medium":1.0,"high":1.0},"frequency_loading":{"annual":1.0,"semi_annual":1.015,"quarterly":1.025,"monthly":1.04}}'::jsonb,
    '["Accidental damage coverage", "Theft protection", "Third-party liability option"]'::jsonb,
    '["Driving without a valid license", "Commercial use without declaration"]'::jsonb,
    TRUE
)
ON CONFLICT (id) DO UPDATE SET
    pricing_rules = EXCLUDED.pricing_rules,
    benefits = EXCLUDED.benefits,
    exclusions = EXCLUDED.exclusions,
    updated_at = NOW();
