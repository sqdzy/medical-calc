import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

function getDefaultApiUrl() {
  // Android emulator cannot reach the host machine via localhost.
  // Use 10.0.2.2 which maps to the host loopback.
  const host = Platform.OS === 'android' ? '10.0.2.2' : 'localhost';
  return `http://${host}:8080/api/v1`;
}

const API_URL = process.env.EXPO_PUBLIC_API_URL || getDefaultApiUrl();

export const api = axios.create({
  baseURL: API_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - attach access token
api.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    const token = await SecureStore.getItemAsync('access_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // If 401 and not already retrying, attempt token refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = await SecureStore.getItemAsync('refresh_token');
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const { data } = await axios.post(`${API_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        const { access_token, refresh_token } = data.data.tokens;

        await SecureStore.setItemAsync('access_token', access_token);
        await SecureStore.setItemAsync('refresh_token', refresh_token);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
        }

        return api(originalRequest);
      } catch (refreshError) {
        // Refresh failed - clear tokens
        await SecureStore.deleteItemAsync('access_token');
        await SecureStore.deleteItemAsync('refresh_token');
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

// Auth API
export const authApi = {
  login: (email: string, password: string) =>
    api.post('/auth/login', { email, password }),

  register: (data: {
    email: string;
    password: string;
    first_name?: string;
    last_name?: string;
    phone?: string;
  }) => api.post('/auth/register', data),

  refresh: (refreshToken: string) =>
    api.post('/auth/refresh', { refresh_token: refreshToken }),

  logout: () => api.post('/auth/logout'),

  me: () => api.get('/auth/me'),
};

import type { SurveyTemplate, SurveyAnswer, SurveyResult } from '../types';

// Surveys API
export const surveysApi = {
  getTemplates: async (): Promise<SurveyTemplate[]> => {
    const res = await api.get('/surveys/templates');
    return res.data.data || res.data;
  },

  getTemplate: async (code: string): Promise<SurveyTemplate> => {
    const res = await api.get(`/surveys/templates/${code}`);
    return res.data.data || res.data;
  },

  submitResponse: async (code: string, answers: SurveyAnswer[]): Promise<SurveyResult> => {
    const res = await api.post(`/surveys/${code}/calculate`, { answers });
    return res.data.data || res.data;
  },

  getResponse: (id: string) => api.get(`/surveys/responses/${id}`),

  getPatientHistory: (patientId: string) => api.get(`/patients/${patientId}/surveys`),
};

// Drugs API
export const drugsApi = {
  list: (search?: string) => api.get('/drugs', { params: { search } }),

  get: (id: string) => api.get(`/drugs/${id}`),

  searchPubChem: (query: string) => api.get('/drugs/pubchem/search', { params: { q: query } }),

  verifyPubChem: (name: string) => api.get('/drugs/pubchem/verify', { params: { name } }),

  searchPubMed: (drug: string) => api.get('/drugs/pubmed/search', { params: { drug } }),
};

// Therapy API
export const therapyApi = {
  createLog: (data: {
    patient_id: string;
    drug_id: string;
    dosage: number;
    dosage_unit: string;
    route?: string;
    administered_at?: string;
    next_scheduled?: string;
    notes?: string;
  }) => api.post('/therapy/logs', data),

  listByPatient: (patientId: string) => api.get(`/patients/${patientId}/therapy`),
};
