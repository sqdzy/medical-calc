-- ============================================
-- AI ADVICE (patient-facing recommendations)
-- ============================================

CREATE TABLE ai_advice (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(id),
    survey_code VARCHAR(50) NOT NULL,
    user_text TEXT,
    score DECIMAL(10,2),
    category VARCHAR(50),
    details JSONB,
    advice_text TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ai_advice_patient_created ON ai_advice(patient_id, created_at DESC);
