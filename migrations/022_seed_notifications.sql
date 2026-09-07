INSERT INTO notifications (id, type, category, severity, title, message, link, is_read, created_at, read_at)
VALUES
(
    'notif_seed_001',
    'APPLICATION_SUBMITTED',
    'underwriting',
    'INFO',
    'Aplikasi Baru: Budi Santoso (#APP-2026-8819)',
    'Pengajuan polis Secure Life Plus baru masuk dan menunggu review underwriting.',
    '/underwriting',
    false,
    NOW() - INTERVAL '10 minutes',
    NULL
),
(
    'notif_seed_002',
    'SLA_WARNING',
    'underwriting',
    'WARNING',
    'SLA Warning: Aplikasi #APP-2026-8812',
    'Aplikasi telah berada dalam status UNDERWRITING_REVIEW selama lebih dari 20 jam.',
    '/underwriting',
    false,
    NOW() - INTERVAL '1 hour',
    NULL
),
(
    'notif_seed_003',
    'KNOWLEDGE_SYNCED',
    'knowledge',
    'SUCCESS',
    'Sinkronisasi Knowledge Base Berhasil',
    'Dokumen SOP Underwriting Terpadu 2026 berhasil divektorisasi ke knowledge base.',
    '/knowledge',
    false,
    NOW() - INTERVAL '3 hours',
    NULL
),
(
    'notif_seed_004',
    'SYSTEM_ALERT',
    'system',
    'INFO',
    'Audit Log Tamper Check Passed',
    'Verifikasi berkala SHA-256 tamper-proof audit trail selesai dengan integritas 100%.',
    '/health',
    true,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day' + INTERVAL '5 minutes'
),
(
    'notif_seed_005',
    'APPLICATION_APPROVED',
    'underwriting',
    'SUCCESS',
    'Polis Diterbitkan: POL-2026-SLP-08819',
    'Aplikasi #APP-2026-8819 telah disetujui dan polis resmi diterbitkan.',
    '/underwriting',
    true,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days' + INTERVAL '10 minutes'
)
ON CONFLICT (id) DO NOTHING;
