---
trigger: manual
---

# Secure GitHub OAuth Implementation for DataHub

## Architecture Overview
This implementation uses a distributed architecture where the frontend (Electron app) and backend (cloud service) run on separate machines. The Electron app runs a persistent local HTTP server on port 9753 to handle OAuth callbacks from GitHub, while the backend runs as a separate cloud service with its own public URL. The authentication flow works as follows: User initiates login → Electron opens browser to GitHub → GitHub redirects to http://localhost:9753/auth/github/callback → Electron captures the authorization code → Electron sends code to backend's /auth/github endpoint → Backend exchanges code for tokens with GitHub → Backend returns JWT to Electron for session management. This separation ensures the OAuth client secret remains secure on the backend while allowing the Electron app to handle the OAuth callback locally

User → Electron (port 9753) → GitHub OAuth → 
Electron callback (localhost:9753) → 
Backend API  → 
JWT back to Electron

## Detailed Implementation Steps

### A. Initial Setup

1. **GitHub OAuth App Setup**:
   - Go to GitHub Settings → Developer settings → OAuth Apps
   - Click "New OAuth App"
   - Fill in application details (app name, homepage URL, description)
   - Set Authorization callback URL to:
     - `http://localhost:9753/auth/github/callback`
   - Note your Client ID and Client Secret

2. **Environment Configuration**:
   - Backend: Store Client ID and Client Secret in environment variables
   - Frontend: Only store Client ID (never the Client Secret) in your Electron app

### B. Frontend Implementation (Electron)

1. **OAuth Initialization with Fixed Port**:
   - Use the fixed port 9753 for OAuth callback handling
   - Prepare OAuth configuration with Client ID and the fixed redirect URI:
     - `http://localhost:9753/auth/github/callback`

2. **Authentication Flow**:
   - When user clicks "Sign in with GitHub" button:
     - Generate a random state value for CSRF protection
     - Construct GitHub authorization URL with required parameters and the fixed port:
       `redirect_uri=http://localhost:9753/auth/github/callback`
     - Open external browser window to this URL
   - If the port is unavailable, show error message:
     "Unable to start authentication service. Please close applications that might be using port 9753 and try again."

3. **Authorization Callback**:
   - Create a local server endpoint to handle the callback on port 9753
   - When user completes authentication in browser, GitHub redirects to your callback
   - Extract authorization code and state from redirect URL
   - Verify state parameter matches originally sent value
   - Close browser window automatically
   - Shut down the local server after authentication is complete

4. **Token Exchange**:
   - Send the authorization code to your backend server
   - Do NOT perform token exchange in the frontend

5. **Session Management**:
   - Store JWT token from backend in secure Electron storage
   - Include JWT in all subsequent API requests as Authorization header
   - Implement token expiration checking
   - Request new JWT from backend when needed

6. **UX Considerations**:
   - Show loading indicators during authentication
   - Handle errors gracefully with user-friendly messages
   - Provide clear instructions if port 9753 is occupied
   - Offer a "Retry" button to attempt authentication again after user resolves port conflict

### C. Backend Implementation (Go/Gin)

1. **API Endpoint Setup**:
   - Create a `/auth/github` endpoint that accepts authorization codes
   - Implement proper request validation and error handling

2. **Exchange Code for Tokens**:
   - Perform the code-for-token exchange with GitHub's token endpoint
   - Use your Client ID and Client Secret for this exchange
   - Request both access_token and refresh_token
   - Handle HTTP errors and JSON parsing robustly

3. **User Verification**:
   - Call GitHub's user API endpoint with the access token
   - Validate user information and ensure account is active
   - Extract user details (username, email, name, avatar URL, GitHub ID)

4. **User Account Management**:
   - Check if user exists in your database
   - If not, create new user record with GitHub information
   - Update existing user information if necessary
   - Associate multiple OAuth providers with same account if needed

5. **Session Token Creation**:
   - Generate a JWT containing user ID and permissions
   - Set appropriate expiration time (e.g., 1 hour)
   - Sign JWT with a secure secret key
   - Return JWT to frontend as authentication token

6. **Token Storage**:
   - Securely store GitHub access token in your database
   - Encrypt sensitive tokens using a strong encryption key
   - Associate tokens with user accounts

7. **Token Refresh Mechanism**:
   - Create endpoint for JWT renewal (e.g., `/auth/refresh`)
   - When backend JWT expires, client requests new JWT
   - Backend uses stored access token to verify user status with GitHub if needed
   - Generate and return new JWT to client

8. **Security Measures**:
   - Implement rate limiting on authentication endpoints
   - Add CSRF protection for auth routes
   - Log authentication attempts and failures
   - Consider IP-based suspicious activity detection

9. **Token Revocation**:
   - Implement logout functionality to invalidate sessions
   - Create mechanism to revoke access tokens if needed
   - Handle cases where GitHub tokens become invalid

## Summary

### Frontend (Electron) Implementation
- Set up local HTTP server for OAuth callback on fixed port 9753
- Display error message if port 9753 is occupied
- Open external browser with GitHub authorization URL including the fixed port
- Capture authorization code from callback
- Send code to backend for token exchange
- Store and manage JWT for authenticated sessions

### Backend (Go/Gin) Implementation
- Handle authorization code from frontend
- Exchange code for GitHub tokens
- Verify user identity
- Create and manage user accounts
- Issue and refresh JWT tokens for session management
- Implement security best practices