UPDATE products
SET pricing_rules = jsonb_set(
    jsonb_set(
        pricing_rules,
        '{sum_assured_presets}',
        '[100000000, 250000000, 500000000, 1000000000]'::jsonb
    ),
    '{payment_term_presets}',
    '[5, 10, 15, 20]'::jsonb
)
WHERE id = 'prod_secure_life_plus';

UPDATE products
SET pricing_rules = jsonb_set(
    jsonb_set(
        pricing_rules,
        '{sum_assured_presets}',
        '[50000000, 100000000, 250000000, 500000000]'::jsonb
    ),
    '{payment_term_presets}',
    '[1, 3, 5, 10]'::jsonb
)
WHERE id = 'prod_health_guard_essential';

UPDATE products
SET pricing_rules = jsonb_set(
    jsonb_set(
        pricing_rules,
        '{sum_assured_presets}',
        '[75000000, 150000000, 300000000, 750000000]'::jsonb
    ),
    '{payment_term_presets}',
    '[1, 2, 3, 5]'::jsonb
)
WHERE id = 'prod_auto_shield_comprehensive';
