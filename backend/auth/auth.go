package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/http_util"
	"github.com/gin-gonic/gin"
)

type AuthCodeRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// Removed GenerateOAuthURLRequest struct for security reasons
// Now using fixed redirect URLs from config

// GoogleAuthCodeHandler handles the OAuth callback from Google
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
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Problem in exchanging code for token", err.Error()))
		return
	}
	fmt.Print(token)
	c.JSON(http.StatusOK, gin.H{"message": "success", "success": true})
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
		"expires_in":   int(token.Expiry.Sub(time.Now()).Seconds()),
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
