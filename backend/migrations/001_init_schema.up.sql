-- 001_init_schema.up.sql
-- Initial database schema for GIBP Medical Application

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- ROLES & PERMISSIONS
-- ============================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================
-- USERS
-- ============================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    role_id UUID REFERENCES roles(id),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- ============================================
-- PATIENTS (with encrypted PII)
-- ============================================

CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    -- Encrypted fields (using application-level encryption)
    full_name_encrypted TEXT NOT NULL,
    birth_date_encrypted TEXT NOT NULL,
    snils_encrypted TEXT,  -- Russian social security number
    -- Non-sensitive data
    gender VARCHAR(10),
    diagnosis TEXT,
    diagnosis_date DATE,
    attending_doctor_id UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_patients_doctor ON patients(attending_doctor_id);

-- ============================================
-- DRUGS (GIBP medications)
-- ============================================

CREATE TABLE drugs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    international_name VARCHAR(255),
    trade_name VARCHAR(255),
    ncbi_pubchem_id VARCHAR(50),
    atc_code VARCHAR(20),  -- Anatomical Therapeutic Chemical code
    dosage_form VARCHAR(100),
    manufacturer VARCHAR(255),
    description TEXT,
    contraindications TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_drugs_name ON drugs(name);
CREATE INDEX idx_drugs_international ON drugs(international_name);

-- ============================================
-- SURVEY TEMPLATES
-- ============================================

CREATE TABLE survey_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),  -- 'vasculitis', 'arthritis', etc.
    questions JSONB NOT NULL,
    scoring_logic JSONB,  -- Scoring rules in JSON format
    interpretation_rules JSONB,  -- Interpretation thresholds
    version INT DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- SURVEY RESPONSES
-- ============================================

CREATE TABLE survey_responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES survey_templates(id),
    patient_id UUID NOT NULL REFERENCES patients(id),
    responses JSONB NOT NULL,
    calculated_score DECIMAL(10,2),
    score_breakdown JSONB,  -- Individual section scores
    interpretation TEXT,  -- AI-generated or rule-based
    ai_summary TEXT,  -- YandexGPT generated summary
    status VARCHAR(20) DEFAULT 'submitted',  -- 'draft', 'submitted', 'reviewed'
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_survey_responses_patient ON survey_responses(patient_id);
CREATE INDEX idx_survey_responses_template ON survey_responses(template_id);
CREATE INDEX idx_survey_responses_submitted ON survey_responses(submitted_at);

-- ============================================
-- MEDICAL INDICES HISTORY
-- ============================================

CREATE TABLE medical_indices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(id),
    index_type VARCHAR(50) NOT NULL,
    value DECIMAL(10,2) NOT NULL,
    category VARCHAR(50),  -- 'remission', 'low', 'moderate', 'high'
    survey_response_id UUID REFERENCES survey_responses(id),
    notes TEXT,
    recorded_by UUID REFERENCES users(id),
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_medical_indices_patient ON medical_indices(patient_id);
CREATE INDEX idx_medical_indices_type ON medical_indices(index_type);
CREATE INDEX idx_medical_indices_recorded ON medical_indices(recorded_at);

-- ============================================
-- THERAPY LOGS (GIBP injections)
-- ============================================

CREATE TABLE therapy_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(id),
    drug_id UUID NOT NULL REFERENCES drugs(id),
    dosage VARCHAR(100) NOT NULL,
    dosage_unit VARCHAR(20),  -- 'mg', 'ml', etc.
    route VARCHAR(50),  -- 'subcutaneous', 'intravenous', etc.
    administered_at TIMESTAMPTZ,
    next_scheduled TIMESTAMPTZ,
    cycle_number INT,
    batch_number VARCHAR(100),
    site VARCHAR(100),  -- injection site
    administered_by UUID REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'scheduled',  -- 'scheduled', 'completed', 'missed', 'cancelled'
    adverse_reactions TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_therapy_logs_patient ON therapy_logs(patient_id);
CREATE INDEX idx_therapy_logs_drug ON therapy_logs(drug_id);
CREATE INDEX idx_therapy_logs_scheduled ON therapy_logs(next_scheduled);
CREATE INDEX idx_therapy_logs_status ON therapy_logs(status);

-- ============================================
-- AUDIT LOG (HIPAA compliance)
-- ============================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- ============================================
-- SEED DATA: Roles & Permissions
-- ============================================

-- Insert default roles
INSERT INTO roles (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'admin', 'System administrator with full access'),
    ('22222222-2222-2222-2222-222222222222', 'doctor', 'Medical doctor - can manage patients and therapy'),
    ('33333333-3333-3333-3333-333333333333', 'nurse', 'Nurse - can view patients and log therapy'),
    ('44444444-4444-4444-4444-444444444444', 'patient', 'Patient - can fill surveys and view own data');

-- Insert permissions
INSERT INTO permissions (id, name, description) VALUES
    -- User management
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'users:read', 'View user list'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaab', 'users:write', 'Create/update users'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaac', 'users:delete', 'Delete users'),
    -- Patient management
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbba', 'patients:read', 'View patient data'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2', 'patients:read_own', 'View own patient data'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'patients:write', 'Create/update patients'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbc', 'patients:delete', 'Delete patients'),
    -- Survey management
    ('cccccccc-cccc-cccc-cccc-ccccccccccca', 'surveys:read', 'View surveys'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'surveys:submit', 'Submit survey responses'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccd', 'surveys:review', 'Review survey responses'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccb', 'surveys:manage', 'Manage survey templates'),
    -- Therapy management
    ('dddddddd-dddd-dddd-dddd-ddddddddddda', 'therapy:read', 'View therapy data'),
    ('dddddddd-dddd-dddd-dddd-ddddddddddd2', 'therapy:read_own', 'View own therapy data'),
    ('dddddddd-dddd-dddd-dddd-dddddddddddb', 'therapy:write', 'Log therapy administration'),
    ('dddddddd-dddd-dddd-dddd-dddddddddddc', 'therapy:schedule', 'Schedule therapy sessions'),
    -- Drug management
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea', 'drugs:read', 'View drug catalog'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeeb', 'drugs:write', 'Manage drug catalog'),
    -- Admin
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'admin:full', 'Full admin access');

-- Assign permissions to roles
-- Admin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '11111111-1111-1111-1111-111111111111', id FROM permissions;

-- Doctor: patient, survey, therapy, drugs read
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbba'),
    ('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
    ('22222222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-ccccccccccca'),
    ('22222222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-cccccccccccd'),
    ('22222222-2222-2222-2222-222222222222', 'dddddddd-dddd-dddd-dddd-ddddddddddda'),
    ('22222222-2222-2222-2222-222222222222', 'dddddddd-dddd-dddd-dddd-dddddddddddb'),
    ('22222222-2222-2222-2222-222222222222', 'dddddddd-dddd-dddd-dddd-dddddddddddc'),
    ('22222222-2222-2222-2222-222222222222', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea');

-- Nurse: view patients, therapy read/write
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbba'),
    ('33333333-3333-3333-3333-333333333333', 'cccccccc-cccc-cccc-cccc-ccccccccccca'),
    ('33333333-3333-3333-3333-333333333333', 'dddddddd-dddd-dddd-dddd-ddddddddddda'),
    ('33333333-3333-3333-3333-333333333333', 'dddddddd-dddd-dddd-dddd-dddddddddddb'),
    ('33333333-3333-3333-3333-333333333333', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea');

-- Patient: own data, submit surveys
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('44444444-4444-4444-4444-444444444444', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2'),
    ('44444444-4444-4444-4444-444444444444', 'cccccccc-cccc-cccc-cccc-ccccccccccca'),
    ('44444444-4444-4444-4444-444444444444', 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
    ('44444444-4444-4444-4444-444444444444', 'dddddddd-dddd-dddd-dddd-ddddddddddd2'),
    ('44444444-4444-4444-4444-444444444444', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea');
