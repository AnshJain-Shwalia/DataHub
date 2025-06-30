---
trigger: manual
---

# Secure Google OAuth Implementation for DataHub

## Detailed Implementation Steps

### A. Initial Setup

1. **Google API Console Setup**:
   - Create a project in Google Cloud Console
   - Enable Google Sign-In API
   - Configure OAuth consent screen (include app name, logo, etc.)
   - Create OAuth credentials (Web Application type)
   - Set authorized JavaScript origins to include `http://localhost` and production URLs
   - Set authorized redirect URIs to include multiple port options:
     - `http://localhost:3000/auth/google/callback`
     - `http://localhost:3001/auth/google/callback`
     - ...up to `http://localhost:3010/auth/google/callback`
   - Note your Client ID and Client Secret

2. **Environment Configuration**:
   - Backend: Store Client ID and Client Secret in environment variables
   - Frontend: Only store Client ID (never the Client Secret) in your Electron app

### B. Frontend Implementation (Electron)

1. **OAuth Initialization and Port Selection**:
   - Create a function to find an available port from range 3000-3010:
     - Try to initialize a local HTTP server on port 3000
     - If port 3000 is occupied, try 3001, then 3002, etc.
     - If all ports are occupied, display a user-friendly error message
     - Store the successfully bound port for use in the OAuth flow
   - Prepare OAuth configuration with Client ID and dynamic redirect URI

2. **Authentication Flow**:
   - When user clicks "Sign in with Google" button:
     - Generate a random state value for CSRF protection
     - Construct Google authorization URL with required parameters and the available port:
       `redirect_uri=http://localhost:${availablePort}/auth/google/callback`
     - Open external browser window to this URL
   - If no ports are available, show error message:
     "Unable to start authentication service. Please close applications that might be using ports 3000-3010 and try again."

3. **Authorization Callback**:
   - Create a local server endpoint to handle the callback on the selected available port
   - When user completes authentication in browser, Google redirects to your callback
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
   - Provide clear instructions if all ports are occupied
   - Offer a "Retry" button to attempt authentication again after user resolves port conflicts
   - Implement automatic retry logic for transient failures

### C. Backend Implementation (Go/Gin)

1. **API Endpoint Setup**:
   - Create a `/auth/google` endpoint that accepts authorization codes
   - Implement proper request validation and error handling

2. **Exchange Code for Tokens**:
   - Perform the code-for-token exchange with Google's token endpoint
   - Use your Client ID and Client Secret for this exchange
   - Request both access_token and refresh_token
   - Handle HTTP errors and JSON parsing robustly

3. **User Verification**:
   - Call Google's userinfo endpoint with the access token
   - Validate email, ensure email is verified
   - Extract user details (name, email, profile picture, Google ID)

4. **User Account Management**:
   - Check if user exists in your database
   - If not, create new user record with Google information
   - Update existing user information if necessary
   - Associate multiple OAuth providers with same account if needed

5. **Session Token Creation**:
   - Generate a JWT containing user ID and permissions
   - Set appropriate expiration time (e.g., 1 hour)
   - Sign JWT with a secure secret key
   - Return JWT to frontend as authentication token

6. **Token Storage**:
   - Securely store Google refresh token in your database
   - Encrypt sensitive tokens using a strong encryption key
   - Associate tokens with user accounts

7. **Token Refresh Mechanism**:
   - Create endpoint for JWT renewal (e.g., `/auth/refresh`)
   - When backend JWT expires, client requests new JWT
   - Backend uses stored refresh token to get new Google access token if needed
   - Generate and return new JWT to client

8. **Security Measures**:
   - Implement rate limiting on authentication endpoints
   - Add CSRF protection for auth routes
   - Log authentication attempts and failures
   - Consider IP-based suspicious activity detection

9. **Token Revocation**:
   - Implement logout functionality to invalidate sessions
   - Create mechanism to revoke refresh tokens if needed
   - Handle cases where Google tokens become invalid

### D. Testing and Validation

1. **Test Authentication Flow**:
   - Verify complete sign-in process works end-to-end
   - Test port availability handling by intentionally blocking ports
   - Test error scenarios (network issues, invalid codes, all ports occupied, etc.)
   - Validate token refresh mechanism

2. **Security Testing**:
   - Attempt CSRF attacks to confirm protection
   - Test token expiration and revocation
   - Validate secure storage of sensitive tokens

## Summary

### Frontend (Electron) Implementation
- Set up local HTTP server for OAuth callback with dynamic port selection (3000-3010)
- Display error message if all configured ports are occupied
- Open external browser with Google authorization URL including selected port
- Capture authorization code from callback
- Send code to backend for token exchange
- Store and manage