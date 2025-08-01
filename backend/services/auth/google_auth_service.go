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

// GoogleAuthService handles Google OAuth authentication operations
type GoogleAuthService struct {
	authService *AuthService
}

// NewGoogleAuthService creates a new instance of GoogleAuthService
func NewGoogleAuthService() *GoogleAuthService {
	return &GoogleAuthService{
		authService: NewAuthService(),
	}
}

// ProcessAuthCodeRequest represents the request structure for processing auth codes
type ProcessAuthCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// ProcessAuthCodeResponse represents the response structure after processing auth codes
type ProcessAuthCodeResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

// ProcessAuthCode processes the OAuth2 authorization code received from Google's OAuth flow.
// This method extracts the complete business logic from GoogleAuthCodeHandler.
//
// This method performs the following steps in sequence:
// 1. Verifies the state parameter to prevent CSRF attacks (one-time use token)
// 2. Exchanges the authorization code for Google OAuth2 tokens (access and refresh)
// 3. Retrieves the user's profile information from Google using the obtained tokens
// 4. Creates a new user account if the user doesn't already exist in the system
// 5. Stores or updates the user's Google OAuth tokens in the database
// 6. Generates a JWT token for authenticated access to the application
//
// Parameters:
//   - request: The authorization code and state from the OAuth callback
//
// Returns:
//   - *ProcessAuthCodeResponse: Contains success status, message, and JWT token
//   - error: Any error that occurred during processing
func (s *GoogleAuthService) ProcessAuthCode(request *ProcessAuthCodeRequest) (*ProcessAuthCodeResponse, error) {
	// Check state BEFORE processing the code
	if !verifyAndConsumeState(request.State) {
		return nil, &AuthError{
			Message: "Invalid state parameter",
			Code:    "INVALID_STATE",
		}
	}

	token, err := exchangeGoogleCodeForTokens(request.Code)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to exchange authorization code for tokens",
			Code:    "TOKEN_EXCHANGE_FAILED",
			Details: err.Error(),
		}
	}

	// Exchange the token for user info
	userInfo, err := getGoogleUserInfo(token)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to retrieve user information from Google",
			Code:    "USER_INFO_FAILED",
			Details: err.Error(),
		}
	}

	// If the user is not present in our database, create a new user account
	user, err := repositories.CreateUserIfNotPresent(userInfo.Email, userInfo.Name)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to create or retrieve user account",
			Code:    "USER_CREATION_FAILED",
			Details: err.Error(),
		}
	}

	// Store or update the user's Google OAuth tokens in the database
	_, err = repositories.UpsertToken(user.ID, "GOOGLE", token.AccessToken, &token.Expiry, &token.RefreshToken, nil)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to store OAuth tokens in database",
			Code:    "TOKEN_STORAGE_FAILED",
			Details: err.Error(),
		}
	}

	// Generate a JWT token for authenticated access to the application
	tokenString, err := s.authService.GenerateJWT(user)
	if err != nil {
		return nil, &AuthError{
			Message: "Failed to generate authentication token",
			Code:    "JWT_GENERATION_FAILED",
			Details: err.Error(),
		}
	}

	// Return success response with JWT token
	return &ProcessAuthCodeResponse{
		Message: "Authentication successful",
		Success: true,
		Token:   tokenString,
	}, nil
}

// GenerateOAuthURL generates and returns the Google OAuth URL with state
func (s *GoogleAuthService) GenerateOAuthURL() (string, error) {
	state, err := generateAndAddState()
	if err != nil {
		return "", &AuthError{
			Message: "Failed to generate OAuth state",
			Code:    "STATE_GENERATION_FAILED",
			Details: err.Error(),
		}
	}

	// Use fixed redirect URL from config for security
	authURL := generateGoogleOAuthURL(state)
	return authURL, nil
}

// AuthError represents a structured error for authentication operations
type AuthError struct {
	Message string
	Code    string
	Details string
}

func (e *AuthError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Google OAuth Utility Functions (merged from oauth package)

// getGoogleUserInfo fetches the authenticated user's profile information from Google's OAuth2 userinfo endpoint.
func getGoogleUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
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

// exchangeGoogleCodeForTokens exchanges an OAuth2 authorization code for an access token and refresh token.
func exchangeGoogleCodeForTokens(code string) (*oauth2.Token, error) {
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

// generateGoogleOAuthURL creates the URL that users should be redirected to for Google OAuth2 authentication.
func generateGoogleOAuthURL(state string) string {
	oauthConfig := createGoogleOAuthConfig(config.LoadConfig().GoogleCallbackURL)
	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}