import axios from 'axios';
import type { AuthResponse } from '../types/global';

// Configure axios defaults
const api = axios.create({
  baseURL: 'http://localhost:8080', // Adjust this to match your backend server
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

export class AuthService {
  /**
   * Get Google OAuth URL from backend
   */
  static async getGoogleOAuthURL(): Promise<{ authURL: string; success: boolean }> {
    try {
      const response = await api.get('/auth/google/oauth-url');
      return response.data;
    } catch (error) {
      console.error('Failed to get Google OAuth URL:', error);
      throw new Error('Failed to get Google OAuth URL');
    }
  }

  /**
   * Get GitHub OAuth URL from backend (requires authentication)
   */
  static async getGitHubOAuthURL(token: string): Promise<{ authURL: string; success: boolean }> {
    try {
      const response = await api.get('/auth/github/oauth-url', {
        headers: { Authorization: `Bearer ${token}` },
      });
      return response.data;
    } catch (error) {
      console.error('Failed to get GitHub OAuth URL:', error);
      throw new Error('Failed to get GitHub OAuth URL');
    }
  }

  /**
   * Exchange Google authorization code for JWT token
   */
  static async exchangeGoogleCode(code: string, state: string): Promise<AuthResponse> {
    try {
      const response = await api.post('/auth/google/', { code, state });
      return response.data;
    } catch (error) {
      console.error('Failed to exchange Google code:', error);
      if (axios.isAxiosError(error) && error.response) {
        return error.response.data;
      }
      throw new Error('Failed to exchange Google authorization code');
    }
  }

  /**
   * Link GitHub storage account to authenticated user
   */
  static async linkGitHubAccount(code: string, state: string, token: string): Promise<AuthResponse> {
    try {
      const response = await api.post('/auth/github/accounts', { code, state }, {
        headers: { Authorization: `Bearer ${token}` },
      });
      return response.data;
    } catch (error) {
      console.error('Failed to link GitHub account:', error);
      if (axios.isAxiosError(error) && error.response) {
        return error.response.data;
      }
      throw new Error('Failed to link GitHub storage account');
    }
  }

  /**
   * Get list of connected GitHub storage accounts
   */
  static async getConnectedGitHubAccounts(token: string): Promise<{ success: boolean; accounts: string[] }> {
    try {
      const response = await api.get('/auth/github/accounts', {
        headers: { Authorization: `Bearer ${token}` },
      });
      return response.data;
    } catch (error) {
      console.error('Failed to get GitHub accounts:', error);
      throw new Error('Failed to get connected GitHub accounts');
    }
  }

  /**
   * Test if backend is reachable
   */
  static async testConnection(): Promise<boolean> {
    try {
      await api.get('/health');
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Validate JWT token (if you have such endpoint)
   */
  static async validateToken(token: string): Promise<boolean> {
    try {
      await api.get('/auth/validate', {
        headers: { Authorization: `Bearer ${token}` },
      });
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Set backend base URL (for configuration)
   */
  static setBackendURL(url: string): void {
    api.defaults.baseURL = url;
  }
}