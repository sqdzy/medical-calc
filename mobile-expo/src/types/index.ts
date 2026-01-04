// User types
export interface User {
  id: string;
  email: string;
  role: 'admin' | 'doctor' | 'patient';
  first_name?: string;
  last_name?: string;
  phone?: string;
  created_at: string;
}

// Auth types
export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

export interface LoginResponse {
  user: User;
  tokens: AuthTokens;
}

// Survey types
export interface QuestionOption {
  value: string;
  label: string;
}

export interface SurveyQuestion {
  id: string;
  text: string;
  type: 'boolean' | 'vas' | 'vas100' | 'select';
  score?: number;
  min_value?: number;
  max_value?: number;
  options?: QuestionOption[];
}

export interface SurveySection {
  section: string;
  questions: SurveyQuestion[];
}

export interface SurveyTemplate {
  id: string;
  code: string;
  name: string;
  description: string;
  category: string;
  questions: SurveyQuestion[];
}

export interface SurveyAnswer {
  question_id: string;
  value: number | boolean;
}

export interface SurveyResult {
  score: number;
  interpretation: string;
}

export interface SurveyResponse {
  id: string;
  template_id: string;
  patient_id: string;
  responses: Record<string, unknown>;
  score?: number;
  interpretation?: string;
  calculated_at?: string;
  created_at: string;
}

// Drug types
export interface Drug {
  id: string;
  name: string;
  inn: string;
  form: string;
  manufacturer?: string;
  atc_code?: string;
  description?: string;
}

// Therapy types
export interface TherapyLog {
  id: string;
  patient_id: string;
  drug_id: string;
  drug_name?: string;
  dosage: number;
  dosage_unit: string;
  route?: string;
  administered_at: string;
  next_scheduled?: string;
  notes?: string;
  created_at: string;
}

// API Response wrapper
export interface ApiResponse<T> {
  data: T;
  meta?: {
    page?: number;
    per_page?: number;
    total?: number;
  };
}

export interface ApiError {
  code: string;
  message: string;
  details?: unknown;
}
