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
  private readonly TOKEN_EXPIRY = 7 * 24 * 60 * 60 * 1000; // 1 week in milliseconds

  constructor() {
    // Initialize localStorage with defaults if not present
    if (!localStorage.getItem('datahub-storage')) {
      localStorage.setItem('datahub-storage', JSON.stringify(defaultData));
    }
  }

  private getData(): StoredData {
    const data = localStorage.getItem('datahub-storage');
    return data ? JSON.parse(data) : defaultData;
  }

  private setData(data: StoredData): void {
    localStorage.setItem('datahub-storage', JSON.stringify(data));
  }

  private get(path: string): any {
    const data = this.getData();
    const keys = path.split('.');
    let result: any = data;
    for (const key of keys) {
      result = result?.[key];
    }
    return result;
  }

  private set(path: string, value: any): void {
    const data = this.getData();
    const keys = path.split('.');
    let current: any = data;
    for (let i = 0; i < keys.length - 1; i++) {
      if (!current[keys[i]]) current[keys[i]] = {};
      current = current[keys[i]];
    }
    current[keys[keys.length - 1]] = value;
    this.setData(data);
  }

  /**
   * Store JWT token securely
   */
  storeToken(token: string): void {
    this.set('auth.token', token);
    this.set('auth.timestamp', Date.now());
  }

  /**
   * Get stored JWT token if still valid
   */
  getToken(): string | null {
    const token = this.get('auth.token') as string | null;
    const timestamp = this.get('auth.timestamp') as number | null;

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
    this.set('auth.user', user);
  }

  /**
   * Get stored user information
   */
  getUser(): { email?: string; name?: string } | null {
    return this.get('auth.user') as { email?: string; name?: string } | null;
  }

  /**
   * Add connected GitHub account
   */
  addGitHubAccount(username: string): void {
    const accounts = (this.get('auth.connectedGitHubAccounts') as string[]) || [];
    if (!accounts.includes(username)) {
      accounts.push(username);
      this.set('auth.connectedGitHubAccounts', accounts);
    }
  }

  /**
   * Get connected GitHub accounts
   */
  getGitHubAccounts(): string[] {
    return (this.get('auth.connectedGitHubAccounts') as string[]) || [];
  }

  /**
   * Remove connected GitHub account
   */
  removeGitHubAccount(username: string): void {
    const accounts = (this.get('auth.connectedGitHubAccounts') as string[]) || [];
    const filtered = accounts.filter((account: string) => account !== username);
    this.set('auth.connectedGitHubAccounts', filtered);
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
    this.set('auth.token', null);
    this.set('auth.timestamp', null);
    this.set('auth.user', null);
    this.set('auth.connectedGitHubAccounts', []);
  }

  /**
   * Settings management
   */
  getBackendURL(): string {
    return this.get('settings.backendURL');
  }

  setBackendURL(url: string): void {
    this.set('settings.backendURL', url);
  }

  getRememberAuth(): boolean {
    return this.get('settings.rememberAuth');
  }

  setRememberAuth(remember: boolean): void {
    this.set('settings.rememberAuth', remember);
  }

  /**
   * Clear all stored data
   */
  clear(): void {
    localStorage.removeItem('datahub-storage');
  }
}

// Export singleton instance
export const storageService = new StorageService();