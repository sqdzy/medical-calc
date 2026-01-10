-- 005_preop_risk_surveys.down.sql
-- Restore original survey templates

-- Remove new preoperative risk surveys
DELETE FROM survey_templates WHERE code IN ('ASA', 'RCRI', 'GOLDMAN', 'CAPRINI');

-- Note: This doesn't restore the original BVAS_V3, DAS28_CRP, BASDAI templates.
-- Run 002_seed_survey_templates.up.sql manually if needed.
