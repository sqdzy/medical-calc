-- 002_seed_survey_templates.down.sql
DELETE FROM survey_templates WHERE code IN ('BVAS_V3', 'DAS28_CRP', 'BASDAI');
