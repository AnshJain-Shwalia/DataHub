import { app, BrowserWindow, ipcMain } from 'electron';
import * as path from 'path';
import * as isDev from 'electron-is-dev';
import { OAuthServer } from './oauth-server';

class MainApp {
  private mainWindow: BrowserWindow | null = null;
  private oauthServer: OAuthServer | null = null;

  constructor() {
    this.init();
  }

  private init(): void {
    // Handle app ready
    app.whenReady().then(() => {
      this.createWindow();
      this.setupOAuthServer();
      this.setupIpcHandlers();

      app.on('activate', () => {
        if (BrowserWindow.getAllWindows().length === 0) {
          this.createWindow();
        }
      });
    });

    // Quit when all windows are closed
    app.on('window-all-closed', () => {
      if (process.platform !== 'darwin') {
        this.cleanup();
        app.quit();
      }
    });

    app.on('before-quit', () => {
      this.cleanup();
    });
  }

  private createWindow(): void {
    this.mainWindow = new BrowserWindow({
      width: 800,
      height: 600,
      minWidth: 600,
      minHeight: 400,
      webPreferences: {
        nodeIntegration: false,
        contextIsolation: true,
        preload: path.join(__dirname, 'preload.js'),
        sandbox: false, // Disable sandbox to avoid permission issues
      },
      icon: isDev ? undefined : path.join(__dirname, '../../assets/icon.png'),
      titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default',
    });

    // Load the app - Force production mode
    // if (isDev) {
    //   this.mainWindow.loadURL('http://localhost:3000');
    //   this.mainWindow.webContents.openDevTools();
    // } else {
      this.mainWindow.loadFile(path.join(__dirname, '../renderer/index.html'));
    // }

    this.mainWindow.on('closed', () => {
      this.mainWindow = null;
    });
  }

  private setupOAuthServer(): void {
    this.oauthServer = new OAuthServer();
    
    this.oauthServer.on('auth-success', (provider: string, code: string, state: string) => {
      this.sendToRenderer('oauth-callback', { provider, code, state, success: true });
    });

    this.oauthServer.on('auth-error', (provider: string, error: string) => {
      this.sendToRenderer('oauth-callback', { provider, error, success: false });
    });
  }

  private setupIpcHandlers(): void {
    // Start OAuth server
    ipcMain.handle('start-oauth-server', async () => {
      try {
        if (this.oauthServer) {
          await this.oauthServer.start();
          return { success: true };
        }
        return { success: false, error: 'OAuth server not initialized' };
      } catch (error) {
        return { success: false, error: (error as Error).message };
      }
    });

    // Stop OAuth server
    ipcMain.handle('stop-oauth-server', async () => {
      try {
        if (this.oauthServer) {
          await this.oauthServer.stop();
          return { success: true };
        }
        return { success: false, error: 'OAuth server not initialized' };
      } catch (error) {
        return { success: false, error: (error as Error).message };
      }
    });

    // Open external URL
    ipcMain.handle('open-external', async (_, url: string) => {
      const { shell } = require('electron');
      await shell.openExternal(url);
      return { success: true };
    });
  }

  private sendToRenderer(channel: string, data: any): void {
    if (this.mainWindow && !this.mainWindow.isDestroyed()) {
      this.mainWindow.webContents.send(channel, data);
    }
  }

  private cleanup(): void {
    if (this.oauthServer) {
      this.oauthServer.stop().catch(console.error);
    }
  }
}

// Initialize the app
new MainApp();