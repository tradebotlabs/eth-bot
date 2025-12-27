import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: 'admin' | 'trader' | 'viewer';
  is_active: boolean;
  is_email_verified: boolean;
  created_at: string;
}

export interface TradingAccount {
  id: string;
  user_id: string;
  account_type: 'demo' | 'live';
  account_name: string;
  demo_initial_capital?: number;
  demo_current_balance?: number;
  binance_api_key_masked?: string;
  binance_testnet: boolean;
  trading_symbol: string;
  trading_mode: 'paper' | 'live';
  enabled_strategies: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  full_name: string;
  account_type: 'demo' | 'live';
  account_name: string;
  demo_initial_capital?: number;
  binance_api_key?: string;
  binance_secret_key?: string;
  binance_testnet?: boolean;
}

interface AuthState {
  // State
  user: User | null;
  accounts: TradingAccount[];
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<boolean>;
  setUser: (user: User | null) => void;
  setAccounts: (accounts: TradingAccount[]) => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
  clearAuth: () => void;
  setError: (error: string | null) => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      accounts: [],
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (credentials: LoginCredentials) => {
        try {
          set({ isLoading: true, error: null });

          const response = await fetch('http://localhost:8080/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credentials),
          });

          if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Login failed');
          }

          const data = await response.json();

          set({
            user: data.user,
            accounts: data.accounts || [],
            accessToken: data.access_token,
            refreshToken: data.refresh_token,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({
            isLoading: false,
            error: message,
            isAuthenticated: false
          });
          throw error;
        }
      },

      register: async (data: RegisterData) => {
        try {
          set({ isLoading: true, error: null });

          const response = await fetch('http://localhost:8080/api/v1/auth/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Registration failed');
          }

          const responseData = await response.json();

          // Auto-login after successful registration
          set({
            user: responseData.user,
            accounts: responseData.accounts || [],
            accessToken: responseData.access_token,
            refreshToken: responseData.refresh_token,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Registration failed';
          set({
            isLoading: false,
            error: message,
            isAuthenticated: false
          });
          throw error;
        }
      },

      logout: async () => {
        try {
          const { accessToken } = get();

          if (accessToken) {
            // Call logout endpoint to invalidate sessions
            await fetch('http://localhost:8080/api/v1/auth/logout', {
              method: 'POST',
              headers: {
                'Authorization': `Bearer ${accessToken}`,
              },
            });
          }
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          // Clear state regardless of API call success
          set({
            user: null,
            accounts: [],
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            error: null,
          });
        }
      },

      refreshAccessToken: async () => {
        try {
          const { refreshToken } = get();

          if (!refreshToken) {
            return false;
          }

          const response = await fetch('http://localhost:8080/api/v1/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
          });

          if (!response.ok) {
            // Refresh token expired or invalid
            get().clearAuth();
            return false;
          }

          const data = await response.json();

          set({
            accessToken: data.access_token,
            refreshToken: data.refresh_token,
          });

          return true;
        } catch (error) {
          console.error('Token refresh error:', error);
          get().clearAuth();
          return false;
        }
      },

      setUser: (user) => set({ user }),

      setAccounts: (accounts) => set({ accounts }),

      setTokens: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken, isAuthenticated: true }),

      clearAuth: () =>
        set({
          user: null,
          accounts: [],
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
          error: null,
        }),

      setError: (error) => set({ error }),

      setLoading: (loading) => set({ isLoading: loading }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        accounts: state.accounts,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
