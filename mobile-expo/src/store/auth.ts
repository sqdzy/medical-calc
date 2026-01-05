import { create } from 'zustand';
import * as SecureStore from 'expo-secure-store';
import { authApi } from '../api/client';
import type { User } from '../types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (email: string, password: string) => Promise<void>;
  register: (data: {
    email: string;
    password: string;
    first_name?: string;
    last_name?: string;
  }) => Promise<void>;
  logout: () => Promise<void>;
  loadStoredAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true,

  login: async (email: string, password: string) => {
    const { data } = await authApi.login(email, password);
    const { user, tokens } = data.data;

    await SecureStore.setItemAsync('access_token', tokens.access_token);
    await SecureStore.setItemAsync('refresh_token', tokens.refresh_token);

    set({ user, isAuthenticated: true, isLoading: false });
  },

  register: async (registerData) => {
    const { data } = await authApi.register(registerData);
    const { user, tokens } = data.data;

    await SecureStore.setItemAsync('access_token', tokens.access_token);
    await SecureStore.setItemAsync('refresh_token', tokens.refresh_token);

    set({ user, isAuthenticated: true, isLoading: false });
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // Ignore logout API errors
    }

    await SecureStore.deleteItemAsync('access_token');
    await SecureStore.deleteItemAsync('refresh_token');

    set({ user: null, isAuthenticated: false, isLoading: false });
  },

  loadStoredAuth: async () => {
    try {
      const accessToken = await SecureStore.getItemAsync('access_token');
      const refreshToken = await SecureStore.getItemAsync('refresh_token');

      // If access token is missing but refresh token exists, refresh session silently
      if (!accessToken && refreshToken) {
        const refreshed = await authApi.refresh(refreshToken);
        const { tokens } = refreshed.data.data;
        await SecureStore.setItemAsync('access_token', tokens.access_token);
        await SecureStore.setItemAsync('refresh_token', tokens.refresh_token);
      }

      const { data } = await authApi.me();
      set({ user: data.data as User, isAuthenticated: true, isLoading: false });
    } catch {
      try {
        // If access token invalid/expired but refresh token exists, retry once via refresh
        const refreshToken = await SecureStore.getItemAsync('refresh_token');
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const refreshed = await authApi.refresh(refreshToken);
        const { tokens } = refreshed.data.data;
        await SecureStore.setItemAsync('access_token', tokens.access_token);
        await SecureStore.setItemAsync('refresh_token', tokens.refresh_token);

        const { data } = await authApi.me();
        set({ user: data.data as User, isAuthenticated: true, isLoading: false });
      } catch {
        await SecureStore.deleteItemAsync('access_token');
        await SecureStore.deleteItemAsync('refresh_token');
        set({ user: null, isAuthenticated: false, isLoading: false });
      }
    }
  },
}));
