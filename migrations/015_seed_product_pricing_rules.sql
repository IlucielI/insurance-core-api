-- 015_seed_product_pricing_rules.sql
-- Seed dynamic pricing rules for existing products and link underwriting questions.

INSERT INTO product_pricing_rules (
    id, product_id, rule_code, rule_name, rule_type, factors, is_active, order_index
) VALUES
-- Life Product (prod_secure_life_plus)
(
    'pr_life_base_rate',
    'prod_secure_life_plus',
    'base_rate',
    'Tarif Dasar Premi Jiwa',
    'base_rate',
    '{"rate": 0.0035}'::jsonb,
    TRUE,
    1
),
(
    'pr_life_age_bracket',
    'prod_secure_life_plus',
    'age_bracket',
    'Faktor Kelompok Usia',
    'bracket',
    '{"brackets": [{"min_age": 18, "max_age": 30, "factor": 1.0}, {"min_age": 31, "max_age": 40, "factor": 1.25}, {"min_age": 41, "max_age": 50, "factor": 1.75}, {"min_age": 51, "max_age": 60, "factor": 2.5}]}'::jsonb,
    TRUE,
    2
),
(
    'pr_life_gender',
    'prod_secure_life_plus',
    'gender',
    'Faktor Aktuaria Jenis Kelamin',
    'multiplier_map',
    '{"male": 1.05, "female": 1.0}'::jsonb,
    TRUE,
    3
),
(
    'pr_life_smoker',
    'prod_secure_life_plus',
    'smoker',
    'Faktor Status Merokok',
    'multiplier_map',
    '{"yes": 1.35, "no": 1.0}'::jsonb,
    TRUE,
    4
),
(
    'pr_life_occupation',
    'prod_secure_life_plus',
    'occupation_class',
    'Faktor Tingkat Risiko Pekerjaan',
    'multiplier_map',
    '{"low": 0.95, "standard": 1.0, "high": 1.4}'::jsonb,
    TRUE,
    5
),
(
    'pr_life_health_risk',
    'prod_secure_life_plus',
    'health_risk',
    'Faktor Profil Risiko Medis',
    'multiplier_map',
    '{"low": 1.0, "medium": 1.25, "high": 1.75}'::jsonb,
    TRUE,
    6
),
(
    'pr_life_frequency_loading',
    'prod_secure_life_plus',
    'frequency_loading',
    'Loading Frekuensi Pembayaran',
    'frequency_loading',
    '{"annual": 1.0, "semi_annual": 1.02, "quarterly": 1.035, "monthly": 1.06}'::jsonb,
    TRUE,
    7
),

-- Health Product (prod_health_guard_essential)
(
    'pr_health_base_rate',
    'prod_health_guard_essential',
    'base_rate',
    'Tarif Dasar Premi Kesehatan',
    'base_rate',
    '{"rate": 0.0042}'::jsonb,
    TRUE,
    1
),
(
    'pr_health_age_bracket',
    'prod_health_guard_essential',
    'age_bracket',
    'Faktor Kelompok Usia Kesehatan',
    'bracket',
    '{"brackets": [{"min_age": 18, "max_age": 30, "factor": 1.0}, {"min_age": 31, "max_age": 40, "factor": 1.2}, {"min_age": 41, "max_age": 50, "factor": 1.6}, {"min_age": 51, "max_age": 60, "factor": 2.2}]}'::jsonb,
    TRUE,
    2
),
(
    'pr_health_gender',
    'prod_health_guard_essential',
    'gender',
    'Faktor Aktuaria Jenis Kelamin',
    'multiplier_map',
    '{"male": 1.03, "female": 1.0}'::jsonb,
    TRUE,
    3
),
(
    'pr_health_smoker',
    'prod_health_guard_essential',
    'smoker',
    'Faktor Status Merokok',
    'multiplier_map',
    '{"yes": 1.25, "no": 1.0}'::jsonb,
    TRUE,
    4
),
(
    'pr_health_occupation',
    'prod_health_guard_essential',
    'occupation_class',
    'Faktor Tingkat Risiko Pekerjaan',
    'multiplier_map',
    '{"low": 0.95, "standard": 1.0, "high": 1.25}'::jsonb,
    TRUE,
    5
),
(
    'pr_health_health_risk',
    'prod_health_guard_essential',
    'health_risk',
    'Faktor Profil Risiko Medis',
    'multiplier_map',
    '{"low": 1.0, "medium": 1.3, "high": 1.9}'::jsonb,
    TRUE,
    6
),
(
    'pr_health_frequency_loading',
    'prod_health_guard_essential',
    'frequency_loading',
    'Loading Frekuensi Pembayaran',
    'frequency_loading',
    '{"annual": 1.0, "semi_annual": 1.02, "quarterly": 1.035, "monthly": 1.06}'::jsonb,
    TRUE,
    7
),

-- Vehicle Product (prod_auto_shield_comprehensive)
(
    'pr_vehicle_base_rate',
    'prod_auto_shield_comprehensive',
    'base_rate',
    'Tarif Dasar Premi Kendaraan',
    'base_rate',
    '{"rate": 0.012}'::jsonb,
    TRUE,
    1
),
(
    'pr_vehicle_age_bracket',
    'prod_auto_shield_comprehensive',
    'age_bracket',
    'Faktor Usia Pengemudi',
    'bracket',
    '{"brackets": [{"min_age": 18, "max_age": 25, "factor": 1.25}, {"min_age": 26, "max_age": 45, "factor": 1.0}, {"min_age": 46, "max_age": 60, "factor": 1.1}]}'::jsonb,
    TRUE,
    2
),
(
    'pr_vehicle_gender',
    'prod_auto_shield_comprehensive',
    'gender',
    'Faktor Jenis Kelamin',
    'multiplier_map',
    '{"male": 1.0, "female": 1.0}'::jsonb,
    TRUE,
    3
),
(
    'pr_vehicle_occupation',
    'prod_auto_shield_comprehensive',
    'occupation_class',
    'Faktor Penggunaan Kendaraan',
    'multiplier_map',
    '{"low": 0.95, "standard": 1.0, "high": 1.15}'::jsonb,
    TRUE,
    4
),
(
    'pr_vehicle_frequency_loading',
    'prod_auto_shield_comprehensive',
    'frequency_loading',
    'Loading Frekuensi Pembayaran',
    'frequency_loading',
    '{"annual": 1.0, "semi_annual": 1.015, "quarterly": 1.025, "monthly": 1.04}'::jsonb,
    TRUE,
    5
)
ON CONFLICT (id) DO UPDATE SET
    rule_code = EXCLUDED.rule_code,
    rule_name = EXCLUDED.rule_name,
    rule_type = EXCLUDED.rule_type,
    factors = EXCLUDED.factors,
    is_active = EXCLUDED.is_active,
    order_index = EXCLUDED.order_index;

-- Link standard template questions to default pricing rules
UPDATE questions SET pricing_rule_id = 'pr_life_gender' WHERE code = 'gender';
UPDATE questions SET pricing_rule_id = 'pr_life_occupation' WHERE code = 'occupation_class';
UPDATE questions SET pricing_rule_id = 'pr_life_smoker' WHERE code = 'is_smoker';
