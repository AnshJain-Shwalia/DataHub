package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/repositories"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// createGitHubOAuthConfig initializes and returns a new OAuth2 configuration for GitHub authentication.
// The redirectURL parameter specifies where GitHub should send the user after authentication.
// This function loads sensitive configuration (client ID and secret) from environment variables.
//
// Required scopes:
//   - repo: Grants full control of private repositories (needed to push file chunks, manage repo contents)
//   - delete_repo: Grants permission to delete repositories (may be required for cleanup or repo management)
//
// Returns:
//   - *oauth2.Config: A configured OAuth2 config ready for use in the GitHub OAuth2 flow.
//
// Notes:
//   - Scopes are hardcoded to ensure minimum required permissions.
//   - All sensitive values are loaded from environment variables for security.
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

// ExchangeGitHubCodeForTokens exchanges an OAuth2 authorization code for an access token and refresh token.
// This is the second step in the OAuth2 authorization code flow for GitHub.
//
// Parameters:
//   - code: The authorization code received from GitHub's OAuth2 redirect.
//
// Returns:
//   - *oauth2.Token: The access token and refresh token if successful.
//   - error: Any error that occurred during the token exchange.
//
// Notes:
//   - The function includes a 10-second timeout to avoid hanging on slow network requests.
//   - The returned token is validated for correctness before use.
//   - The token can be used to make authenticated requests to GitHub APIs on behalf of the user.
//   - GitHub's OAuth2 provider may not set the token type; this function ensures it is set to "Bearer" if missing.
func ExchangeGitHubCodeForTokens(code string) (*oauth2.Token, error) {
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

// GenerateGitHubOAuthURL creates the URL that users should be redirected to for GitHub OAuth2 authentication.
// This is the first step in the OAuth2 authorization code flow.
//
// Parameters:
//   - state: A cryptographically random string used to protect against CSRF attacks.
//
// Returns:
//   - string: The complete GitHub OAuth2 authorization URL.
//
// The generated URL includes:
//   - The client ID from configuration.
//   - The default callback URL from configuration.
//   - The requested scopes (repo, delete_repo).
//   - The provided state parameter for CSRF protection.
//   - AccessTypeOffline to request a refresh token.
//   - ApprovalForce to ensure the consent screen is always shown.
//
// Notes:
//   - This URL should be used to redirect users to GitHub for authentication.
//   - The state parameter is required for CSRF protection.
func GenerateGitHubOAuthURL(state string) string {
	oauthConfig := createGitHubOAuthConfig(config.LoadConfig().GithubCallbackURL)
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// GenerateGitHubOAuthURLWithRedirectURL creates a GitHub OAuth2 authorization URL with a custom redirect URL.
// This is a more flexible version of GenerateGitHubOAuthURL that allows specifying a custom redirect URL.
//
// Parameters:
//   - state: A cryptographically random string used to protect against CSRF attacks.
//   - redirectURL: The URL where GitHub should redirect after authentication. If empty, falls back to the default callback URL from configuration.
//
// Returns:
//   - string: The complete GitHub OAuth2 authorization URL.
//
// The generated URL includes all standard OAuth2 parameters plus:
//   - The provided redirect URL (or default if not specified).
//   - The state parameter for CSRF protection.
//   - AccessTypeOffline to request a refresh token.
//   - A "prompt" parameter set to "consent" to ensure the authorization page is always shown.
//
// Notes:
//   - This function is useful when you need to override the default callback URL, such as when handling authentication from different domains or environments.
//   - The state parameter is required for CSRF protection.
//   - The "prompt=consent" parameter forces the consent screen to be shown even if the user previously authorized the app.
func GenerateGitHubOAuthURLWithRedirectURL(state string, redirectURL string) string {
	// If redirect URL is empty, use the default callback URL from configuration
	if redirectURL == "" {
		redirectURL = config.LoadConfig().GithubCallbackURL
	}
	oauthConfig := createGitHubOAuthConfig(redirectURL)

	// GitHub doesn't support ApprovalForce, but we can add a login parameter
	// to ensure the authorization page is always shown
	opts := []oauth2.AuthCodeOption{
		// Request offline access to get a refresh token
		oauth2.AccessTypeOffline,
		// Force showing the authorization page even if the user has previously authorized
		oauth2.SetAuthURLParam("prompt", "consent"),
	}

	return oauthConfig.AuthCodeURL(state, opts...)
}

// GitHubUserInfo contains the user profile information returned by GitHub's OAuth2 user endpoint.
// This struct is used to parse the JSON response from GitHub's API.
//
// Fields:
//   - Login: The user's GitHub username (used as unique identifier)
//   - Name: The user's display name (can be empty)
//   - Email: The user's primary email address (can be empty if private)
//   - ID: The user's unique GitHub ID (numeric)
//
// Login is marked as required since it's used as the account identifier.
type GitHubUserInfo struct {
	Login string  `json:"login" binding:"required"`           // GitHub username
	Name  *string `json:"name"`                               // Display name (optional)
	Email *string `json:"email"`                              // Primary email (optional, can be private)
	ID    int64   `json:"id" binding:"required"`              // Unique GitHub user ID
}

// GetGitHubUserInfo fetches the authenticated user's profile information from GitHub's OAuth2 user endpoint.
//
// Parameters:
//   - token: A valid OAuth2 token obtained from GitHub's token endpoint
//
// Returns:
//   - *GitHubUserInfo: The user's profile information if successful
//   - error: Any error that occurred during the request or parsing
//
// The function includes a 10-second timeout to prevent hanging on slow network requests.
func GetGitHubUserInfo(token *oauth2.Token) (*GitHubUserInfo, error) {
	return GetGitHubUserInfoFromAccessToken(token.AccessToken)
}

// GetGitHubUserInfoFromAccessToken fetches user profile information using a raw access token string.
// This function uses resty HTTP client with proper timeout and error handling.
//
// Parameters:
//   - accessToken: A valid OAuth2 access token string
//
// Returns:
//   - *GitHubUserInfo: The user's profile information if successful
//   - error: Any error that occurred during the request or parsing
func GetGitHubUserInfoFromAccessToken(accessToken string) (*GitHubUserInfo, error) {
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

// IsGitHubAccessTokenValid checks if a GitHub access token is valid by making a request to GitHub's user endpoint
//
// Parameters:
//   - tokenID: The ID of the token to validate
//
// Returns:
//   - bool: true if the access token is valid, false otherwise
func IsGitHubAccessTokenValid(tokenID string) bool {
	token, err := repositories.GetTokenByID(tokenID)
	if err != nil {
		return false
	}

	_, err = GetGitHubUserInfoFromAccessToken(token.AccessToken)
	return err == nil
}
