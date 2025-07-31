package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/repositories"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// createGoogleOAuthConfig initializes and returns a new OAuth2 configuration for Google authentication.
// The redirectURL parameter specifies where Google should send the user after authentication.
// This function loads sensitive configuration (client ID and secret) from environment variables.
//
// Required scopes:
//   - userinfo.email: View the user's email address
//   - userinfo.profile: View basic profile information
//
// Returns a configured oauth2.Config ready for use in the OAuth2 flow.
func createGoogleOAuthConfig(redirectURL string) *oauth2.Config {
	envCfg := config.LoadConfig()
	return &oauth2.Config{
		ClientID:     envCfg.GoogleClientID,
		ClientSecret: envCfg.GoogleClientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GoogleUserInfo contains the user profile information returned by Google's OAuth2 userinfo endpoint.
// This struct is used to parse the JSON response from Google's API.
//
// Fields:
//   - Email: The user's email address (validated as a proper email format)
//   - Name: The user's full name
//
// Both fields are marked as required in the binding tag to ensure they are always present.
type GoogleUserInfo struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

// GetGoogleUserInfo fetches the authenticated user's profile information from Google's OAuth2 userinfo endpoint.
//
// Parameters:
//   - token: A valid OAuth2 token obtained from Google's token endpoint
//
// Returns:
//   - *GoogleUserInfo: The user's profile information if successful
//   - error: Any error that occurred during the request or parsing
//
// The function includes a 10-second timeout to prevent hanging on slow network requests.
// It automatically handles the OAuth2 client creation and token management.
func GetGoogleUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a client with the token
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// Make request to Google's userinfo endpoint
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status code is OK
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, body)
	}

	// Parse the JSON response
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	// Validate using Gin's validator
	if err := binding.Validator.ValidateStruct(userInfo); err != nil {
		return nil, fmt.Errorf("invalid user info: %v", err)
	}

	return &userInfo, nil
}

// GetGoogleUserInfoFromAccessToken fetches user profile information using a raw access token string.
// This is an alternative to GetGoogleUserInfo that doesn't require the full oauth2.Token struct.
//
// Parameters:
//   - accessToken: A valid OAuth2 access token string
//
// Returns:
//   - *GoogleUserInfo: The user's profile information if successful
//   - error: Any error that occurred during the request or parsing
//
// Note: This function makes a direct HTTP request to Google's userinfo endpoint
// and includes a 10-second timeout to prevent hanging on slow network requests.
func GetGoogleUserInfoFromAccessToken(accessToken string) (*GoogleUserInfo, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add the authorization header
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()
	resp.Cookies()
	// Check if the response status code is OK
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, body)
	}

	// Parse the JSON response
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	// Validate using Gin's validator
	if err := binding.Validator.ValidateStruct(userInfo); err != nil {
		return nil, fmt.Errorf("invalid user info: %v", err)
	}

	return &userInfo, nil
}

// ExchangeGoogleCodeForTokens exchanges an OAuth2 authorization code for an access token and refresh token.
// This is the second step in the OAuth2 authorization code flow.
//
// Parameters:
//   - code: The authorization code received from Google's OAuth2 redirect
//
// Returns:
//   - *oauth2.Token: The access token and refresh token if successful
//   - error: Any error that occurred during the token exchange
//
// The function includes a 10-second timeout and validates that the returned token is valid.
// The token can be used to make authenticated requests to Google APIs on behalf of the user.
func ExchangeGoogleCodeForTokens(code string) (*oauth2.Token, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new config with default redirect URL
	oauthConfig := createGoogleOAuthConfig(config.LoadConfig().GoogleCallbackURL)

	// Exchange will handle all the HTTP details and parameter encoding
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %v", err)
	}

	// Validate the token
	if !token.Valid() {
		return nil, fmt.Errorf("received invalid token from Google")
	}

	return token, nil
}

// GenerateGoogleOAuthURL creates the URL that users should be redirected to for Google OAuth2 authentication.
// This is the first step in the OAuth2 authorization code flow.
//
// Parameters:
//   - state: A cryptographically random string used to protect against CSRF attacks
//
// Returns:
//   - string: The complete Google OAuth2 authorization URL
//
// The generated URL includes:
// - The client ID from configuration
// - The default callback URL from configuration
// - The requested scopes (email and profile)
// - The provided state parameter for CSRF protection
// - AccessTypeOffline to request a refresh token
// - ApprovalForce to ensure the consent screen is always shown
func GenerateGoogleOAuthURL(state string) string {
	oauthConfig := createGoogleOAuthConfig(config.LoadConfig().GoogleCallbackURL)
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// GenerateGoogleOAuthURLWithRedirectURL creates a Google OAuth2 authorization URL with a custom redirect URL.
// This is a more flexible version of GenerateGoogleOAuthURL that allows specifying a custom redirect URL.
//
// Parameters:
//   - state: A cryptographically random string used to protect against CSRF attacks
//   - redirectURL: The URL where Google should redirect after authentication.
//     If empty, falls back to the default callback URL from configuration.
//
// Returns:
//   - string: The complete Google OAuth2 authorization URL
//
// The generated URL includes all standard OAuth2 parameters plus:
// - The provided redirect URL (or default if not specified)
// - The state parameter for CSRF protection
// - AccessTypeOffline to request a refresh token
// - ApprovalForce to ensure the consent screen is always shown
//
// This function is useful when you need to override the default callback URL,
// such as when handling authentication from different domains or environments.
func GenerateGoogleOAuthURLWithRedirectURL(state string, redirectURL string) string {
	if redirectURL == "" {
		redirectURL = config.LoadConfig().GoogleCallbackURL
	}
	oauthConfig := createGoogleOAuthConfig(redirectURL)
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// IsGoogleRefreshTokenValid checks if a Google refresh token is valid by attempting to use it
//
// Parameters:
//   - tokenID: The ID of the token to validate
//
// Returns:
//   - bool: true if the refresh token is valid, false otherwise
func IsGoogleRefreshTokenValid(tokenID string) bool {
	token, err := repositories.GetTokenByID(tokenID)
	if err != nil || token.RefreshToken == nil {
		return false
	}

	oauthConfig := createGoogleOAuthConfig(config.LoadConfig().GoogleCallbackURL)
	
	// Create a token with just the refresh token
	oauthToken := &oauth2.Token{
		RefreshToken: *token.RefreshToken,
		Expiry: time.Now().Add(-time.Hour), // Force refresh by setting expired time
	}
	
	tokenSource := oauthConfig.TokenSource(context.Background(), oauthToken)
	_, err = tokenSource.Token()
	return err == nil
}

// IsGoogleAccessTokenValid checks if a Google access token is valid by making a request to Google's userinfo endpoint
//
// Parameters:
//   - tokenID: The ID of the token to validate
//
// Returns:
//   - bool: true if the access token is valid, false otherwise
func IsGoogleAccessTokenValid(tokenID string) bool {
	token, err := repositories.GetTokenByID(tokenID)
	if err != nil {
		return false
	}

	_, err = GetGoogleUserInfoFromAccessToken(token.AccessToken)
	return err == nil
}

// GoogleAccessTokenValidityLeft returns the remaining validity duration of a Google access token
//
// Parameters:
//   - tokenID: The ID of the token to check
//
// Returns:
//   - time.Duration: The remaining validity duration, 0 if expired or invalid
func GoogleAccessTokenValidityLeft(tokenID string) time.Duration {
	token, err := repositories.GetTokenByID(tokenID)
	if err != nil || token.AccessTokenExpiry == nil {
		return 0
	}

	now := time.Now()
	if token.AccessTokenExpiry.Before(now) {
		return 0
	}

	return token.AccessTokenExpiry.Sub(now)
}

// RefreshGoogleAccessToken refreshes a Google access token using the stored refresh token
//
// Parameters:
//   - tokenID: The ID of the token to refresh
//
// Returns:
//   - error: nil if successful, error if refresh failed
func RefreshGoogleAccessToken(tokenID string) error {
	token, err := repositories.GetTokenByID(tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	if token.RefreshToken == nil {
		return fmt.Errorf("no refresh token available")
	}

	oauthConfig := createGoogleOAuthConfig(config.LoadConfig().GoogleCallbackURL)
	
	// Create oauth2.Token with current values
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: *token.RefreshToken,
		Expiry:       time.Now().Add(-time.Hour), // Force refresh
	}
	
	tokenSource := oauthConfig.TokenSource(context.Background(), oauthToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update the token in database
	_, err = repositories.UpdateToken(token, newToken.AccessToken, &newToken.Expiry, &newToken.RefreshToken, nil)
	if err != nil {
		return fmt.Errorf("failed to update token in database: %v", err)
	}

	return nil
}
