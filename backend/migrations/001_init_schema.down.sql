-- 001_init_schema.down.sql
-- Rollback initial schema

-- Drop permissions mappings first
DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS therapy_logs;
DROP TABLE IF EXISTS medical_indices;
DROP TABLE IF EXISTS survey_responses;
DROP TABLE IF EXISTS survey_templates;
DROP TABLE IF EXISTS drugs;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
