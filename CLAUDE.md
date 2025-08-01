# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DataHub is a GitHub-as-cloud-storage desktop application built with **Electron frontend** and **Go backend**. It treats GitHub repositories as distributed cloud storage by chunking files and pushing them to user-owned GitHub repos, with an S3-compatible object store as temporary buffer.

**⚠️ Development Status**: This project is in active development/MVP phase with no deployed users. Database schema changes can be made freely.

**Architecture**: Electron app ↔ Go backend (Gin + GORM + PostgreSQL) ↔ S3-compatible storage ↔ GitHub repositories

## Tech Stack

- **Frontend**: Electron desktop app
- **Backend**: Go with Gin (HTTP), GORM (PostgreSQL ORM), gin-swagger (API docs)
- **Database**: PostgreSQL (with optional Redis for caching)
- **Temporary Storage**: S3-compatible (AWS S3, Cloudflare R2)
- **Permanent Storage**: GitHub repositories via OAuth
- **Authentication**: Google OAuth (user signup/login only) + GitHub OAuth (storage account linking for authenticated users)
- **Serverless**: Platform-agnostic functions (Go/TypeScript) in `/lambda/`

## Development Commands

### Backend (`/backend/`)
```bash
cd backend
go run main.go    # Starts server on configured PORT
```

### Environment Setup
Required environment variables (see `backend/config/config.go`):
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
- `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` 
- `DATABASE_URL` (PostgreSQL connection string)
- `PORT` (server port)
- `JWT_SECRET` (for token signing)
- `GOOGLE_CALLBACK_URL` (defaults to `http://localhost:9753/auth/google/callback`)
- `GITHUB_CALLBACK_URL` (defaults to `http://localhost:9753/auth/github/callback`)

## OAuth Authentication Architecture

### Distributed OAuth Flow Overview
DataHub uses a **distributed OAuth architecture** with distinct flows for authentication vs storage:

**User Authentication (Google OAuth)**:
- **Electron app** runs a local HTTP server on port 9753 for OAuth callbacks
- **Go backend** (cloud service) handles token exchange, CSRF protection, and user account creation
- **Google OAuth** redirects to the Electron app's local server
- **Purpose**: User signup and login only

**Storage Account Linking (GitHub OAuth)**:
- **Same technical flow** but requires existing authenticated user (JWT token)
- **GitHub OAuth** is used to add storage accounts to already authenticated users
- **Purpose**: Link GitHub repositories for distributed storage
- **Requirement**: User must be logged in with Google first

This pattern ensures OAuth client secrets remain secure on the backend while allowing the desktop app to handle authentication flows locally.

### Detailed Authentication Flow

#### Step-by-Step Process:

1. **User Initiates Authentication** (Electron)
   - **Initial signup/login**: User clicks "Sign in with Google" button (only option for new users)
   - **Add storage account**: Authenticated user clicks "Connect GitHub Account" button
   - Electron app starts HTTP server on port 9753
   - If port unavailable, show error: *"Unable to start authentication service. Please close applications using port 9753 and try again."*

2. **Get OAuth URL** (Electron → Backend)
   ```http
   GET /auth/google/oauth-url              # For user signup/login
   GET /auth/github/oauth-url              # For adding storage accounts
   Authorization: Bearer <jwt>             # Required for GitHub OAuth URL
   ```
   **Response:**
   ```json
   {
     "authURL": "https://accounts.google.com/oauth/authorize?client_id=...&state=BACKEND_GENERATED_STATE",
     "success": true
   }
   ```
   **Note**: Backend generates and manages the CSRF state parameter

3. **Open Browser** (Electron)
   - Open system browser to OAuth URL from backend
   - Display loading indicator: *"Waiting for authentication..."*
   - **No state handling needed** - backend manages CSRF protection

4. **User Authenticates** (Browser → OAuth Provider)
   - User completes OAuth flow with Google/GitHub
   - Provider validates credentials and permissions

5. **OAuth Callback** (OAuth Provider → Electron)
   - Provider redirects to: `http://localhost:9753/auth/{provider}/callback?code=ABC123&state=BACKEND_STATE`
   - Electron HTTP server captures the callback
   - Extract `code` and `state` parameters (both as-received)
   - Close browser window automatically
   - Shut down local HTTP server

6. **Forward to Backend** (Electron → Backend)
   ```http
   POST /auth/google/                      # User signup/login
   Content-Type: application/json
   
   {
     "code": "ABC123",
     "state": "BACKEND_STATE"
   }
   ```
   ```http
   POST /auth/github/accounts              # Add storage account (requires auth)
   Authorization: Bearer <jwt>
   Content-Type: application/json
   
   {
     "code": "ABC123",
     "state": "BACKEND_STATE"
   }
   ```
   **Note**: Electron just forwards both parameters without validation

7. **Backend Processing** (Backend Internal)
   - Verify and consume state (CSRF protection)
   - Exchange authorization code for OAuth tokens
   - Fetch user profile from OAuth provider
   - **Google OAuth**: Create user account if doesn't exist, generate JWT token (1-week expiration)
   - **GitHub OAuth**: Extract user ID from JWT, link GitHub account to existing user
   - Store OAuth tokens in database:
     - **Google**: Single token per user (upsert) - creates user accounts
     - **GitHub**: Multiple tokens per user, one per GitHub account - requires existing user

8. **Return Response** (Backend → Electron)
   **Google OAuth** (returns JWT for new session):
   ```json
   {
     "message": "Authentication successful",
     "success": true,
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   }
   ```
   **GitHub OAuth** (confirms account linking):
   ```json
   {
     "message": "GitHub account linked successfully",
     "success": true,
     "githubUsername": "user123"
   }
   ```

9. **Store Session** (Electron)
   - Store JWT in secure Electron storage
   - Include JWT in all subsequent API requests: `Authorization: Bearer <jwt>`
   - Handle token expiration and refresh

### Electron Implementation Requirements

The Electron app must implement the following OAuth authentication components:

#### 1. **Local HTTP Server (Port 9753)**
```pseudocode
function startAuthServer():
  try:
    start HTTP server on port 9753
    return server instance
  catch EADDRINUSE:
    throw "Port 9753 is in use. Please close applications using this port and try again."
```

#### 2. **OAuth Callback Handler**
```pseudocode
route GET /auth/:provider/callback:
  extract code, state, error from query parameters
  if error exists:
    close browser window
    notify main process of error
    return

  close browser window automatically
  forward (provider, code, state) to backend exchange function
  shutdown local HTTP server
```

#### 3. **OAuth Initiation Flow**
```pseudocode
function initiateOAuth(provider):
  start local HTTP server on port 9753
  request OAuth URL from backend: GET /auth/{provider}/oauth-url
  open system browser to received authURL
  show "Waiting for authentication..." UI
  set 5-minute timeout to cleanup server
```

#### 4. **Backend Token Exchange**
```pseudocode
function exchangeCodeForJWT(provider, code, state):
  POST to backend /auth/{provider}/ with JSON: {code, state}
  if response.success:
    store JWT securely in encrypted storage
    notify main process: authentication successful
  else:
    notify main process: authentication failed with error message
```

#### 5. **Secure Token Management**
```pseudocode
function storeJWT(token):
  encrypt and store token in secure Electron storage
  store timestamp for expiration checking

function getValidJWT():
  retrieve stored token
  check if token expired (1 week)
  return token or null if expired
```

#### 6. **Error Handling Strategy**
```pseudocode
function handleAuthError(error):
  map technical errors to user-friendly messages:
    - EADDRINUSE → "Port 9753 is in use..."
    - NETWORK_ERROR → "Unable to connect to authentication service..."
    - INVALID_STATE → "Authentication failed due to security error..."
  
  display error dialog with retry option
```

#### 7. **GitHub Storage Account Management**
```pseudocode
function connectGitHubStorageAccount():
  // Requires user to be authenticated with Google first
  // Same OAuth flow but posts to /auth/github/accounts with JWT
  if not authenticated:
    show error: "Please sign in with Google first"
    return
  
  initiateOAuth('github')  // Posts to /auth/github/accounts with JWT

function listConnectedGitHubAccounts():
  GET /auth/github/accounts with JWT authorization header
  return list of connected GitHub usernames for storage
```

#### 8. **Session Management**
```pseudocode
function makeAuthenticatedRequest(url, options):
  jwt = getValidJWT()
  if jwt is null:
    redirect to login screen
    return

  add Authorization header: "Bearer " + jwt
  make HTTP request
  if response is 401 Unauthorized:
    clear stored JWT
    redirect to login screen
```

### File Storage Flow
1. **Upload**: Files >5MB chunked → S3 buffer → Serverless function pushes to GitHub → Metadata in PostgreSQL
2. **Download**: Client queries metadata → Git sparse checkout → Chunk reassembly

### System Limits
- **Chunk size**: Max 5MB per chunk
- **Per repo limit**: 500MB
- **User quota**: Up to 500GB across ~1,000 GitHub repos

## Database Models (GORM)

### Core Entity Overview
The database follows a hierarchical structure optimized for distributed file storage across GitHub repositories:

**Authentication Layer**: `User` ↔ `Token` (OAuth credentials)
**Storage Layer**: `Repo` → `Branch` → `Chunk` (GitHub storage hierarchy)  
**File System Layer**: `User` → `Folder` → `File` → `Chunk` (user file organization)

### Model Definitions (`/backend/models/`)

#### User (`user.go`)
**Purpose**: Core user account management for OAuth authentication
```go
type User struct {
    ID        string    // UUID primary key
    Name      string    // Display name from OAuth provider
    Email     string    // Unique email address (indexed)
    Tokens    []Token   // Associated OAuth tokens (not stored in DB)
    CreatedAt time.Time
    UpdatedAt time.Time
}
```
- **Authentication**: Single user per email address with Google OAuth signup
- **Relationships**: One-to-many with Token, Folder, and File entities
- **Indexing**: Email field indexed for fast login lookups

#### Token (`token.go`) 
**Purpose**: OAuth token storage with multi-platform support and duplicate prevention
```go
type Token struct {
    ID                   string     // UUID primary key
    UserID               string     // Foreign key to User
    Platform             string     // "GOOGLE" or "GITHUB"
    AccountIdentifier    *string    // GitHub username or Google email
    AccessToken          string     // OAuth access token (encrypted)
    AccessTokenExpiry    *time.Time // Token expiration
    RefreshToken         *string    // OAuth refresh token (encrypted)
    RefreshTokenExpiry   *time.Time
    AccessTokenIssuedAt  time.Time
    RefreshTokenIssuedAt *time.Time
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```
- **Multi-Platform**: Supports both Google (authentication) and GitHub (storage) tokens
- **Duplicate Prevention**: Unique constraint on (UserID, Platform, AccountIdentifier)
- **Token Management**: Full OAuth2 lifecycle with access/refresh token tracking
- **Security**: Tokens encrypted at rest with expiration monitoring

#### Repo (`repo.go`)
**Purpose**: GitHub repository metadata for distributed storage
```go
type Repo struct {
    ID        string    // UUID primary key
    GithubID  *string   // GitHub repository ID from API
    TokenID   string    // Foreign key to GitHub Token
    Name      string    // Repository name
    Branches  []Branch  // Associated branches (not stored in DB)
    CreatedAt time.Time
}
```
- **GitHub Integration**: Links to actual GitHub repositories via OAuth tokens
- **Storage Organization**: Each repo can contain multiple branches for data organization
- **Token Association**: Tied to specific GitHub OAuth tokens for access control

#### Branch (`branch.go`)
**Purpose**: Git branch organization within repositories for chunk storage
```go
type Branch struct {
    ID        string    // UUID primary key
    Name      string    // Git branch name
    RepoID    string    // Foreign key to Repo
    Chunks    []Chunk   // File chunks stored in this branch (not stored in DB)
    CreatedAt time.Time
}
```
- **Git Integration**: Represents actual Git branches in GitHub repositories
- **Chunk Organization**: Branches contain file chunks pushed to GitHub
- **Storage Strategy**: Separate branches can isolate different data types or users

#### Folder (`folder.go`)
**Purpose**: Hierarchical folder structure for user file organization
```go
type Folder struct {
    ID             string    // UUID primary key
    Name           string    // Folder name
    ParentFolderID *string   // Self-referencing for hierarchy
    ParentFolder   *Folder   // Parent folder relationship
    UserID         string    // Foreign key to User (indexed)
    Files          []File    // Files in this folder (not stored in DB)
    Subfolders     []Folder  // Child folders (not stored in DB)
    CreatedAt      time.Time
}
```
- **Hierarchical Structure**: Self-referencing tree structure for nested folders
- **User Isolation**: Each folder belongs to a specific user
- **File Organization**: Provides familiar file system structure over distributed storage

#### File (`file.go`)
**Purpose**: File metadata with chunking information for large file handling
```go
type File struct {
    ID        string    // UUID primary key
    Name      string    // Original filename
    FolderID  *string   // Optional parent folder
    Size      int64     // Total file size in bytes
    UserID    string    // Foreign key to User (indexed)
    Chunks    []Chunk   // File chunks (not stored in DB)
    CreatedAt time.Time
}
```
- **Chunking Support**: Large files split into 5MB chunks for GitHub storage
- **Size Tracking**: Maintains original file size for reconstruction
- **User Ownership**: Files belong to specific users with folder organization
- **Orphan Support**: Files can exist without folders (FolderID nullable)

#### Chunk (`chunk.go`)
**Purpose**: Individual file chunks stored in GitHub repositories with ordering
```go
type Chunk struct {
    ID        string    // UUID primary key
    FileID    string    // Foreign key to File (indexed)
    Rank      int       // Chunk order for file reconstruction
    Size      int64     // Chunk size in bytes (≤5MB)
    Path      string    // File path in GitHub repository
    BranchID  string    // Foreign key to Branch (indexed)
    CreatedAt time.Time
}
```
- **File Reconstruction**: Rank field ensures proper chunk ordering during download
- **GitHub Storage**: Path field maps to actual file location in repository
- **Size Limits**: Each chunk ≤5MB to comply with GitHub file size limits
- **Distribution**: Chunks distributed across branches and repositories for load balancing

### Database Relationships

#### Primary Relationships
- `User` → `Token` (1:N) - Users can have multiple OAuth tokens
- `User` → `Folder` (1:N) - Users own their folder hierarchies  
- `User` → `File` (1:N) - Users own their files
- `Token` → `Repo` (1:N) - GitHub tokens can access multiple repositories
- `Repo` → `Branch` (1:N) - Repositories contain multiple branches
- `Branch` → `Chunk` (1:N) - Branches store file chunks
- `File` → `Chunk` (1:N) - Files split into ordered chunks
- `Folder` → `Folder` (1:N) - Hierarchical folder structure
- `Folder` → `File` (1:N) - Folders contain files

#### Key Constraints
- **User.Email**: UNIQUE constraint prevents duplicate accounts
- **Token (UserID, Platform, AccountIdentifier)**: UNIQUE constraint prevents duplicate OAuth tokens
- **Chunk.Rank**: Ordering constraint ensures proper file reconstruction
- **Foreign Keys**: Enforced referential integrity across all relationships

### Storage Architecture Integration

#### File Upload Flow
1. **File** metadata created with total size
2. **File** split into **Chunks** (≤5MB each) with rank ordering
3. **Chunks** uploaded to S3 temporary buffer
4. Serverless function pushes **Chunks** to **GitHub** via **Branch**/**Repo**
5. **Chunk** records updated with GitHub **Path** locations

#### File Download Flow  
1. Query **File** metadata and associated **Chunks** by rank
2. Retrieve **Chunk** locations from **Branch**/**Repo** GitHub paths
3. Git sparse checkout to download specific **Chunks**
4. Reassemble **Chunks** in rank order to reconstruct original **File**

This model design supports DataHub's core functionality of treating GitHub repositories as distributed cloud storage while maintaining familiar file system semantics for users.

## API Endpoints

### Authentication Routes (`/auth/`)

#### OAuth URL Generation
- `GET /auth/google/oauth-url` - Returns Google OAuth authorization URL with backend-generated state (public)
- `GET /auth/github/oauth-url` - Returns GitHub OAuth authorization URL with backend-generated state (requires JWT authorization)

**Response Format:**
```json
{
  "authURL": "https://accounts.google.com/oauth/authorize?client_id=...&state=...",
  "success": true
}
```

#### Authentication & Storage Account Management
- `POST /auth/google/` - Exchanges Google authorization code for JWT token (creates user accounts)
- `POST /auth/github/accounts` - Links GitHub storage account to authenticated user (requires JWT authorization)
- `GET /auth/github/accounts` - Lists connected GitHub storage accounts (requires JWT authorization)

**Request Format:**
```json
{
  "code": "authorization_code_from_oauth_callback",
  "state": "state_parameter_from_oauth_callback"
}
```

**Google OAuth Success Response:**
```json
{
  "message": "Authentication successful",
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**GitHub Account Linking Success Response:**
```json
{
  "message": "GitHub account linked successfully",
  "success": true,
  "githubUsername": "user123"
}
```

**List GitHub Accounts Response:**
```json
{
  "success": true,
  "accounts": ["user123", "user456"]
}
```

**Error Response:**
```json
{
  "error": "Invalid state parameter",
  "success": false,
  "statusCode": 400
}
```

### Token Management Behavior
- **Google OAuth**: Single token per user (upserts existing token) - **Creates user accounts**
- **GitHub OAuth**: Multiple tokens per user, one per unique GitHub account - **Requires existing authenticated user**
- **JWT Expiration**: 1 week from generation
- **Token Storage**: OAuth tokens encrypted in database with account identifiers
- **Authentication Flow**: Users must authenticate with Google first, then can add GitHub storage accounts

### Troubleshooting Common Issues

#### Port 9753 Already in Use
- **Symptom**: "EADDRINUSE" error when starting OAuth flow
- **Solution**: Kill processes using port 9753 or restart system
- **Prevention**: Always cleanup server instances after auth completion

#### Authentication Timeout
- **Symptom**: User doesn't complete OAuth flow within 5 minutes
- **Solution**: Restart authentication process
- **Implementation**: Set server timeout and cleanup resources

#### State Parameter Mismatch
- **Symptom**: "Invalid state parameter" error from backend
- **Cause**: CSRF protection - state was already consumed or invalid
- **Solution**: Generate new OAuth URL and restart flow

#### GitHub Authentication Without Google Login
- **Symptom**: "Unauthorized" error when trying to connect GitHub account
- **Cause**: User attempting to use GitHub OAuth without being authenticated with Google first
- **Solution**: Ensure user completes Google authentication before attempting GitHub account linking

#### JWT Token Expiration
- **Symptom**: 401 Unauthorized errors after 1 week
- **Solution**: Implement token refresh or re-authentication flow
- **Implementation**: Check token expiry before API calls

### Database Configuration
- PostgreSQL with GORM auto-migration on startup
- Connection pooling optimized for Neon free tier (5 idle, 20 max connections)
- UTC timezone enforcement, 300ms slow query logging
- Custom logger with proper connection lifecycle management

## Development Guidelines

### OAuth Implementation Security
- **Frontend**: Only store OAuth Client IDs, never Client Secrets
- **Backend**: Secure token exchange, encrypt stored refresh tokens
- **Port Management**: Handle port 9753 conflicts gracefully with user-friendly error messages
- **Session Management**: JWT tokens with 1-week expiration, proper refresh mechanism

### Testing
Use Bruno API client with endpoint files in `backend/endpoints/`:
- Google/GitHub OAuth URL generation and authentication flows
- Authorization testing and state management