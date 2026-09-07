-- Seed realistic insurance applications and review checks for Underwriting CMS demonstration
INSERT INTO applications (
    id, product_id, full_name, email, phone, age, gender,
    sum_assured, payment_term, payment_frequency, smoker,
    occupation_class, health_risk, premium, status,
    created_at, updated_at, reviewed_at, reviewed_by, rejection_reason
) VALUES
(
    'app_demo_submitted_01',
    'prod_secure_life_plus',
    'Budi Santoso',
    'budi.santoso@example.com',
    '081234567890',
    34,
    'male',
    500000000,
    20,
    'monthly',
    'no',
    'standard',
    'low',
    450000,
    'submitted',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '2 hours',
    NULL,
    NULL,
    NULL
),
(
    'app_demo_review_02',
    'prod_health_guard_essential',
    'Siti Rahmawati',
    'siti.rahmawati@example.com',
    '082198765432',
    29,
    'female',
    250000000,
    10,
    'annual',
    'no',
    'low',
    'low',
    2640000,
    'under_review',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 hour',
    NULL,
    NULL,
    NULL
),
(
    'app_demo_ready_approve_03',
    'prod_secure_life_plus',
    'Ahmad Fauzi',
    'ahmad.fauzi@example.com',
    '081377889900',
    42,
    'male',
    750000000,
    15,
    'monthly',
    'no',
    'standard',
    'low',
    825000,
    'under_review',
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '30 minutes',
    NULL,
    NULL,
    NULL
),
(
    'app_demo_rejected_04',
    'prod_auto_shield_comprehensive',
    'Dewi Lestari',
    'dewi.lestari@example.com',
    '085611223344',
    38,
    'female',
    350000000,
    5,
    'annual',
    'no',
    'high',
    'low',
    3800000,
    'rejected',
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '4 days',
    NOW() - INTERVAL '4 days',
    'Lead Underwriter',
    'Riwayat klaim kendaraan tidak wajar dan dokumen polis tidak lengkap'
)
ON CONFLICT (id) DO NOTHING;

-- Seed review checks for the applications above
INSERT INTO application_review_checks (
    id, application_id, check_type, status, notes, reviewed_by, reviewed_at, created_at, updated_at
) VALUES
-- app_demo_submitted_01 (Baru masuk, semua masih pending)
('chk_sub01_identity', 'app_demo_submitted_01', 'identity_verified', 'pending', 'Menunggu verifikasi Dukcapil OCR', NULL, NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
('chk_sub01_income', 'app_demo_submitted_01', 'income_verified', 'pending', 'Menunggu validasi mutasi rekening / slip gaji', NULL, NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
('chk_sub01_docs', 'app_demo_submitted_01', 'documents_complete', 'pending', 'Menunggu kelengkapan dokumen pendukung', NULL, NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
('chk_sub01_medical', 'app_demo_submitted_01', 'medical_required', 'pending', 'Menunggu skrining riwayat penyakit', NULL, NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),

-- app_demo_review_02 (Sedang ditinjau: identitas & income passed, dokumen & medis pending)
('chk_rev02_identity', 'app_demo_review_02', 'identity_verified', 'passed', 'KTP terverifikasi cocok dengan data Dukcapil', 'System OCR Worker', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 hour'),
('chk_rev02_income', 'app_demo_review_02', 'income_verified', 'passed', 'DSR sehat (rasio premi 2.1% dari pendapatan bulanan)', 'Staff Underwriter', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 hour'),
('chk_rev02_docs', 'app_demo_review_02', 'documents_complete', 'pending', 'Foto selfie liveness sedang dalam antrean validasi', NULL, NULL, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
('chk_rev02_medical', 'app_demo_review_02', 'medical_required', 'pending', 'Kuesioner kesehatan sedang dianalisis', NULL, NULL, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),

-- app_demo_ready_approve_03 (Semua 4 pilar passed / not_needed -> siap di-approve)
('chk_app03_identity', 'app_demo_ready_approve_03', 'identity_verified', 'passed', 'Identitas dan biometrik terverifikasi 100%', 'Lead Underwriter', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 hour'),
('chk_app03_income', 'app_demo_ready_approve_03', 'income_verified', 'passed', 'Slip gaji valid & SPT tahunan terkonfirmasi', 'Lead Underwriter', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 hour'),
('chk_app03_docs', 'app_demo_ready_approve_03', 'documents_complete', 'passed', 'Semua dokumen lengkap dan sah secara legal', 'Lead Underwriter', NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '3 days', NOW() - INTERVAL '45 minutes'),
('chk_app03_medical', 'app_demo_ready_approve_03', 'medical_required', 'not_needed', 'Non-medical underwriting memenuhi syarat tier standar', 'Lead Underwriter', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '3 days', NOW() - INTERVAL '30 minutes'),

-- app_demo_rejected_04 (Dokumen failed -> ditolak)
('chk_rej04_identity', 'app_demo_rejected_04', 'identity_verified', 'passed', 'KTP valid', 'Staff Underwriter', NOW() - INTERVAL '4 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days'),
('chk_rej04_income', 'app_demo_rejected_04', 'income_verified', 'passed', 'Penghasilan memadai', 'Staff Underwriter', NOW() - INTERVAL '4 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days'),
('chk_rej04_docs', 'app_demo_rejected_04', 'documents_complete', 'failed', 'Dokumen kepemilikan kendaraan fiktif terdeteksi', 'Lead Underwriter', NOW() - INTERVAL '4 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days'),
('chk_rej04_medical', 'app_demo_rejected_04', 'medical_required', 'not_needed', 'Asuransi kendaraan tidak memerlukan skrining medis', 'Lead Underwriter', NOW() - INTERVAL '4 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days')
ON CONFLICT (id) DO NOTHING;
