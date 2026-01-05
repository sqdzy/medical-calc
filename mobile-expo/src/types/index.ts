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
  type: 'boolean' | 'number' | 'scale' | 'select' | 'text' | 'vas' | 'vas100';
  score?: number;
  min?: number;
  max?: number;
  // Backend currently stores options as string[]; keep it flexible.
  options?: string[] | QuestionOption[];
  labels?: Record<string, string>;
  required?: boolean;
  extra?: Record<string, unknown>;
}

export interface SurveySection {
  section: string;
  title?: string;
  questions: SurveyQuestion[];
}

export interface SurveyTemplate {
  id: string;
  code: string;
  name: string;
  description: string;
  category: string;
  questions: SurveySection[];
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
  dosage: string;
  dosage_unit: string;
  route?: string;
  administered_at?: string;
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
