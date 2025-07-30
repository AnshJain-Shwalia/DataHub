import { contextBridge, ipcRenderer } from 'electron';

// Define the API that will be exposed to the renderer process
const electronAPI = {
  // OAuth server management
  startOAuthServer: () => ipcRenderer.invoke('start-oauth-server'),
  stopOAuthServer: () => ipcRenderer.invoke('stop-oauth-server'),
  
  // External URL handling
  openExternal: (url: string) => ipcRenderer.invoke('open-external', url),
  
  // OAuth callback listener
  onOAuthCallback: (callback: (data: any) => void) => {
    const wrappedCallback = (_: any, data: any) => callback(data);
    ipcRenderer.on('oauth-callback', wrappedCallback);
    
    // Return cleanup function
    return () => {
      ipcRenderer.removeListener('oauth-callback', wrappedCallback);
    };
  },
};

// Expose the API to the renderer process
contextBridge.exposeInMainWorld('electronAPI', electronAPI);

// Type definitions for the exposed API
export type ElectronAPI = typeof electronAPI;