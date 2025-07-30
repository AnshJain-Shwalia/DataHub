package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/http_util"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/AnshJain-Shwalia/DataHub/backend/repositories"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// GoogleAuthCodeHandler processes the OAuth2 authorization code received from Google's OAuth flow.
//
// This handler performs the following steps in sequence:
// 1. Validates the request body structure containing the authorization code and state parameter
// 2. Verifies the state parameter to prevent CSRF attacks (one-time use token)
// 3. Exchanges the authorization code for Google OAuth2 tokens (access and refresh)
// 4. Retrieves the user's profile information from Google using the obtained tokens
// 5. Creates a new user account if the user doesn't already exist in the system
// 6. Stores or updates the user's Google OAuth tokens in the database
// 7. Generates a JWT token for authenticated access to the application
//
// Parameters:
//   - c *gin.Context: The Gin context containing the HTTP request and response
//
// Response:
//   - Success (200): Returns a JWT token for authenticated API access
//   - Error (400): Returns detailed error information if any step fails
//
// The function handles all error cases with appropriate HTTP status codes and messages.
func GoogleAuthCodeHandler(c *gin.Context) {
	var body AuthCodeRequest
	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Incorrect body structure", err.Error()))
		return
	}

	// Check state BEFORE processing the code
	if !VerifyAndConsumeState(body.State) {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Invalid state parameter", nil))
		return
	}

	token, err := ExchangeGoogleCodeForTokens(body.Code)

	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to exchange authorization code for tokens", err.Error()))
		return
	}
	fmt.Print(token)

	// Exchange the token for user info
	userInfo, err := GetGoogleUserInfo(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to retrieve user information from Google", err.Error()))
		return
	}
	// If the user is not present in our database, create a new user account
	user, err := repositories.CreateUserIfNotPresent(userInfo.Email, userInfo.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to create or retrieve user account", err.Error()))
		return
	}
	// Store or update the user's Google OAuth tokens in the database
	_, err = repositories.UpsertToken(user.ID, "GOOGLE", token.AccessToken, &token.Expiry, &token.RefreshToken, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to store OAuth tokens in database", err.Error()))
		return
	}
	// Generate a JWT token for authenticated access to the application
	tokenString, err := GenerateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to generate authentication token", err.Error()))
		return
	}
	// Return success response with JWT token
	c.JSON(http.StatusOK, gin.H{"message": "Authentication successful", "success": true, "token": tokenString})
}

// GenerateGoogleOAuthURLHandler generates the OAuth URL for Google login
func GenerateGoogleOAuthURLHandler(c *gin.Context) {
	state, err := GenerateAndAddState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "", nil))
		return
	}

	// Use fixed redirect URL from config for security
	authURL := GenerateGoogleOAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"authURL": authURL,
		"success": true,
	})
}

// GitHubAuthCodeHandler handles the OAuth callback from GitHub
// GitHub will call this endpoint with the following query parameters:
// - code: The authorization code that can be exchanged for an access token
// - state: The unguessable random string you provided in the initial request
// - error: (optional) If there was an error during authorization
// - error_description: (optional) Description of the error
// - error_uri: (optional) URI to a page with more information about the error
func GitHubAuthCodeHandler(c *gin.Context) {
	// First check for OAuth errors
	if errMsg := c.Query("error"); errMsg != "" {
		errDescription := c.Query("error_description")
		err := fmt.Errorf("GitHub OAuth error: %s", errMsg)
		if errDescription != "" {
			err = fmt.Errorf("%s: %s", err, errDescription)
		}
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "GitHub authorization failed", err.Error()))
		return
	}

	// Get the authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "No authorization code provided by GitHub", "missing_code"))
		return
	}

	// Verify the state parameter to prevent CSRF attacks
	state := c.Query("state")
	if state == "" || !VerifyAndConsumeState(state) {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Invalid or missing state parameter", "invalid_state"))
		return
	}

	// Exchange the authorization code for an access token
	token, err := ExchangeGitHubCodeForTokens(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "Failed to exchange authorization code for access token", err.Error()))
		return
	}

	// In a real application, you would:
	// 1. Store the token securely (encrypted in the database)
	// 2. Associate it with the current user's session
	// 3. Return a session token to the client

	// For now, we'll just return a success message with the token
	// In production, never expose the raw token to the client
	c.JSON(http.StatusOK, gin.H{
		"message":      "GitHub authentication successful",
		"access_token": token.AccessToken,
		"token_type":   token.TokenType,
		"expires_in":   int(time.Until(token.Expiry).Seconds()),
		"success":      true,
	})
}

// GenerateGitHubOAuthURLHandler generates the OAuth URL for GitHub login
func GenerateGitHubOAuthURLHandler(c *gin.Context) {
	state, err := GenerateAndAddState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "", nil))
		return
	}

	authURL := GenerateGitHubOAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"authURL": authURL,
		"success": true,
	})
}

func GenerateJWTToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour * 7)

	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.LoadConfig().JWTSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
