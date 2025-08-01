package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	tokenservice "github.com/AnshJain-Shwalia/DataHub/backend/services/token"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// createGitHubOAuthConfig initializes and returns a new OAuth2 configuration for GitHub authentication.
func createGitHubOAuthConfig(redirectURL string) *oauth2.Config {
	envCfg := config.LoadConfig()
	return &oauth2.Config{
		ClientID:     envCfg.GitHubClientID,
		ClientSecret: envCfg.GitHubClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"repo", "delete_repo"}, // Hardcoded scopes as requested
		Endpoint:     github.Endpoint,
	}
}

// GitHubUserInfo contains the user profile information returned by GitHub's OAuth2 user endpoint.
type GitHubUserInfo struct {
	Login string  `json:"login" binding:"required"`           // GitHub username
	Name  *string `json:"name"`                               // Display name (optional)
	Email *string `json:"email"`                              // Primary email (optional, can be private)
	ID    int64   `json:"id" binding:"required"`              // Unique GitHub user ID
}

// GitHubAuthService handles GitHub OAuth authentication operations
type GitHubAuthService struct {
	tokenService *tokenservice.TokenService
}

// NewGitHubAuthService creates a new instance of GitHubAuthService
func NewGitHubAuthService() *GitHubAuthService {
	return &GitHubAuthService{
		tokenService: tokenservice.NewTokenService(),
	}
}

// AddAccountRequest represents the request structure for adding GitHub accounts
type AddAccountRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// AddAccountResponse represents the response structure after adding GitHub accounts
type AddAccountResponse struct {
	Message        string `json:"message"`
	Success        bool   `json:"success"`
	GitHubUsername string `json:"githubUsername"`
}

// GetAccountsResponse represents the response structure for listing GitHub accounts
type GetAccountsResponse struct {
	Success  bool     `json:"success"`
	Accounts []string `json:"accounts"`
}

// AddAccount processes the OAuth2 authorization code received from GitHub's OAuth flow
// to link a GitHub storage account to an already authenticated user.
// This method extracts the complete business logic from AddGitHubAccountHandler.
//
// This method performs the following steps in sequence:
// 1. Verifies the state parameter to prevent CSRF attacks (one-time use token)
// 2. Exchanges the authorization code for GitHub OAuth2 tokens (access token)
// 3. Retrieves the user's profile information from GitHub using the obtained tokens
// 4. Links the GitHub account to the existing authenticated user (no new user creation)
// 5. Stores the GitHub OAuth token with account identifier to support multiple GitHub accounts
//
// Parameters:
//   - userID: The ID of the authenticated user (from JWT token)
//   - request: The authorization code and state from the OAuth callback
//
// Returns:
//   - *AddAccountResponse: Contains success status, message, and GitHub username
//   - error: Any error that occurred during processing
func (s *GitHubAuthService) AddAccount(userID string, request *AddAccountRequest) (*AddAccountResponse, error) {
	// Check state BEFORE processing the code
	if !verifyAndConsumeState(request.State) {
		return nil, &AuthError{
			Message: "Invalid state parameter",
			Code:    "INVALID_STATE",
		}
	}

	// Exchange the authorization code for an access token
	token, err := exchangeGitHubCodeForTokens(request.Code)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to exchange authorization code for tokens",
			Code:    "TOKEN_EXCHANGE_FAILED",
			Details: err.Error(),
		}
	}

	// Exchange the token for user info
	userInfo, err := getGitHubUserInfo(token)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to retrieve user information from GitHub",
			Code:    "USER_INFO_FAILED",
			Details: err.Error(),
		}
	}

	// Store the GitHub OAuth token with account identifier (GitHub username) for the authenticated user
	_, err = s.tokenService.CreateOrUpdateGitHubToken(userID, userInfo.Login, token.AccessToken, &token.Expiry, nil, nil)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to store GitHub OAuth token in database",
			Code:    "TOKEN_STORAGE_FAILED",
			Details: err.Error(),
		}
	}

	// Return success response confirming GitHub account linking
	return &AddAccountResponse{
		Message:        "GitHub account linked successfully",
		Success:        true,
		GitHubUsername: userInfo.Login,
	}, nil
}

// GetAccounts lists all connected GitHub accounts for the authenticated user.
// This method extracts the complete business logic from GetGitHubAccountsHandler.
//
// This method performs the following steps:
// 1. Retrieves all GitHub tokens associated with the user from the database
// 2. Returns a list of connected GitHub usernames
//
// Parameters:
//   - userID: The ID of the authenticated user (from JWT token)
//
// Returns:
//   - *GetAccountsResponse: Contains success status and list of GitHub usernames
//   - error: Any error that occurred during processing
func (s *GitHubAuthService) GetAccounts(userID string) (*GetAccountsResponse, error) {
	// Get all GitHub tokens for the user
	githubTokens, err := s.tokenService.GetGitHubTokensForUser(userID)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to retrieve GitHub accounts",
			Code:    "ACCOUNTS_RETRIEVAL_FAILED",
			Details: err.Error(),
		}
	}

	// Extract GitHub usernames from the tokens
	var githubUsernames []string
	for _, token := range githubTokens {
		if token.AccountIdentifier != nil {
			githubUsernames = append(githubUsernames, *token.AccountIdentifier)
		}
	}

	// Return the list of GitHub accounts
	return &GetAccountsResponse{
		Success:  true,
		Accounts: githubUsernames,
	}, nil
}

// GenerateOAuthURL generates and returns the GitHub OAuth URL with state
func (s *GitHubAuthService) GenerateOAuthURL() (string, error) {
	state, err := generateAndAddState()
	if err != nil {
		return "", &AuthError{
			Message: "Failed to generate OAuth state",
			Code:    "STATE_GENERATION_FAILED",
			Details: err.Error(),
		}
	}

	authURL := generateGitHubOAuthURL(state)
	return authURL, nil
}

// GitHub OAuth Utility Functions (merged from oauth package)

// exchangeGitHubCodeForTokens exchanges an OAuth2 authorization code for an access token and refresh token.
func exchangeGitHubCodeForTokens(code string) (*oauth2.Token, error) {
	// Create a context with timeout (10 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new config with default redirect URL
	oauthConfig := createGitHubOAuthConfig(config.LoadConfig().GithubCallbackURL)

	// Exchange the code for a token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("GitHub code exchange failed: %v", err)
	}

	// GitHub's OAuth2 provider doesn't set the token type by default
	// So we set it to "Bearer" if it's empty
	if token.TokenType == "" {
		token.TokenType = "Bearer"
	}

	// Validate the token
	if !token.Valid() {
		return nil, fmt.Errorf("received invalid token from GitHub")
	}

	return token, nil
}

// generateGitHubOAuthURL creates the URL that users should be redirected to for GitHub OAuth2 authentication.
func generateGitHubOAuthURL(state string) string {
	oauthConfig := createGitHubOAuthConfig(config.LoadConfig().GithubCallbackURL)
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// getGitHubUserInfo fetches the authenticated user's profile information from GitHub's OAuth2 user endpoint.
func getGitHubUserInfo(token *oauth2.Token) (*GitHubUserInfo, error) {
	return getGitHubUserInfoFromAccessToken(token.AccessToken)
}

// getGitHubUserInfoFromAccessToken fetches user profile information using a raw access token string.
func getGitHubUserInfoFromAccessToken(accessToken string) (*GitHubUserInfo, error) {
	// Create resty client with timeout
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetAuthToken(accessToken)

	var userInfo GitHubUserInfo
	var errorResponse map[string]interface{}

	// Make the request
	resp, err := client.R().
		SetResult(&userInfo).
		SetError(&errorResponse).
		Get("https://api.github.com/user")

	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	// Check if the response was successful
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("GitHub API request failed with status %d: %v", resp.StatusCode(), errorResponse)
	}

	// Validate using Gin's validator
	if err := binding.Validator.ValidateStruct(userInfo); err != nil {
		return nil, fmt.Errorf("invalid user info received from GitHub: %v", err)
	}

	return &userInfo, nil
}