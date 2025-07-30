import type { ElectronAPI } from '../main/preload';

declare global {
  interface Window {
    electronAPI: ElectronAPI;
  }
}

export interface OAuthCallbackData {
  provider: string;
  code?: string;
  state?: string;
  error?: string;
  success: boolean;
}

export interface AuthResponse {
  message: string;
  success: boolean;
  token?: string;
  githubUsername?: string; // For GitHub account linking responses
  error?: string;
  statusCode?: number;
}

export interface AuthState {
  isAuthenticated: boolean;
  token: string | null;
  user: {
    email?: string;
    name?: string;
  } | null;
  connectedGitHubAccounts: string[];
}

export interface ServerResponse {
  success: boolean;
  error?: string;
}