# DataHub OAuth Test App

A simple Electron application built with React, TypeScript, and Chakra UI to test the DataHub OAuth authentication flows.

## Features

- 🔐 Test Google OAuth authentication
- 🔐 Test GitHub OAuth authentication (supports multiple accounts)
- 💾 Secure JWT token storage
- 🔄 Real-time backend connection status
- 🎨 Clean UI with Chakra UI components
- ⚡ Hot reload during development

## Prerequisites

- Node.js (v16 or higher)
- Your DataHub backend server running

## Installation

```bash
npm install
```

## Development

```bash
# Start development mode (runs main process and renderer in parallel)
npm run dev

# In another terminal, start the Electron app
npm start
```

## Building

```bash
# Build for production
npm run build

# Package the app
npm run package
```

## Usage

1. Start your DataHub backend server
2. Update the backend URL in the app if needed (default: http://localhost:8080)
3. Click "Sign in with Google" or "Sign in with GitHub"
4. Complete the OAuth flow in your browser
5. View your authentication status and connected accounts

## OAuth Flow

The app implements the distributed OAuth pattern:
1. Starts a local HTTP server on port 9753
2. Requests OAuth URL from backend
3. Opens browser for user authentication
4. Captures OAuth callback on port 9753
5. Exchanges authorization code for JWT token via backend
6. Stores JWT securely using electron-store

## Troubleshooting

- **Port 9753 in use**: Close any applications using this port
- **Backend offline**: Ensure your Go backend server is running
- **Authentication fails**: Check backend logs for detailed error messages