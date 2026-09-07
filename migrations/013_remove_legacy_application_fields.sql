-- 013_remove_legacy_application_fields.sql
-- Remove legacy tightly-coupled underwriting fields from applications table.
-- Detailed answers are now fully preserved dynamically in application_answers table.

ALTER TABLE applications
    DROP COLUMN IF EXISTS smoker,
    DROP COLUMN IF EXISTS occupation_class,
    DROP COLUMN IF EXISTS health_risk;
