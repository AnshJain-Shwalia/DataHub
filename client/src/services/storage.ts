import Store from 'electron-store';
import type { AuthState } from '../types/global';

interface StoredData {
  auth: {
    token: string | null;
    timestamp: number | null;
    user: {
      email?: string;
      name?: string;
    } | null;
    connectedGitHubAccounts: string[];
  };
  settings: {
    backendURL: string;
    rememberAuth: boolean;
  };
}

const defaultData: StoredData = {
  auth: {
    token: null,
    timestamp: null,
    user: null,
    connectedGitHubAccounts: [],
  },
  settings: {
    backendURL: 'http://localhost:8080',
    rememberAuth: true,
  },
};

class StorageService {
  private store: Store<StoredData>;
  private readonly TOKEN_EXPIRY = 7 * 24 * 60 * 60 * 1000; // 1 week in milliseconds

  constructor() {
    this.store = new Store<StoredData>({
      defaults: defaultData,
      encryptionKey: 'datahub-oauth-test-key', // In production, use a more secure key
    });
  }

  /**
   * Store JWT token securely
   */
  storeToken(token: string): void {
    this.store.set('auth.token', token);
    this.store.set('auth.timestamp', Date.now());
  }

  /**
   * Get stored JWT token if still valid
   */
  getToken(): string | null {
    const token = this.store.get('auth.token') as string | null;
    const timestamp = this.store.get('auth.timestamp') as number | null;

    if (!token || !timestamp || typeof timestamp !== 'number') {
      return null;
    }

    // Check if token is expired (1 week)
    if (Date.now() - timestamp > this.TOKEN_EXPIRY) {
      this.clearAuth();
      return null;
    }

    return token;
  }

  /**
   * Store user information
   */
  storeUser(user: { email?: string; name?: string }): void {
    this.store.set('auth.user', user);
  }

  /**
   * Get stored user information
   */
  getUser(): { email?: string; name?: string } | null {
    return this.store.get('auth.user') as { email?: string; name?: string } | null;
  }

  /**
   * Add connected GitHub account
   */
  addGitHubAccount(username: string): void {
    const accounts = (this.store.get('auth.connectedGitHubAccounts') as string[]) || [];
    if (!accounts.includes(username)) {
      accounts.push(username);
      this.store.set('auth.connectedGitHubAccounts', accounts);
    }
  }

  /**
   * Get connected GitHub accounts
   */
  getGitHubAccounts(): string[] {
    return (this.store.get('auth.connectedGitHubAccounts') as string[]) || [];
  }

  /**
   * Remove connected GitHub account
   */
  removeGitHubAccount(username: string): void {
    const accounts = (this.store.get('auth.connectedGitHubAccounts') as string[]) || [];
    const filtered = accounts.filter((account: string) => account !== username);
    this.store.set('auth.connectedGitHubAccounts', filtered);
  }

  /**
   * Get current auth state
   */
  getAuthState(): AuthState {
    const token = this.getToken();
    const user = this.getUser();
    const connectedGitHubAccounts = this.getGitHubAccounts();

    return {
      isAuthenticated: !!token,
      token,
      user,
      connectedGitHubAccounts,
    };
  }

  /**
   * Clear all authentication data
   */
  clearAuth(): void {
    this.store.set('auth.token', null);
    this.store.set('auth.timestamp', null);
    this.store.set('auth.user', null);
    this.store.set('auth.connectedGitHubAccounts', []);
  }

  /**
   * Settings management
   */
  getBackendURL(): string {
    return this.store.get('settings.backendURL');
  }

  setBackendURL(url: string): void {
    this.store.set('settings.backendURL', url);
  }

  getRememberAuth(): boolean {
    return this.store.get('settings.rememberAuth');
  }

  setRememberAuth(remember: boolean): void {
    this.store.set('settings.rememberAuth', remember);
  }

  /**
   * Clear all stored data
   */
  clear(): void {
    this.store.clear();
  }
}

// Export singleton instance
export const storageService = new StorageService();