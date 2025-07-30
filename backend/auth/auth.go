package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/http_util"
	"github.com/AnshJain-Shwalia/DataHub/backend/middleware"
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

// AddGitHubAccountHandler processes the OAuth2 authorization code received from GitHub's OAuth flow
// to link a GitHub storage account to an already authenticated user.
//
// This handler performs the following steps in sequence:
// 1. Extracts user ID from JWT token (user must be already authenticated)
// 2. Validates the request body structure containing the authorization code and state parameter
// 3. Verifies the state parameter to prevent CSRF attacks (one-time use token)
// 4. Exchanges the authorization code for GitHub OAuth2 tokens (access token)
// 5. Retrieves the user's profile information from GitHub using the obtained tokens
// 6. Links the GitHub account to the existing authenticated user (no new user creation)
// 7. Stores the GitHub OAuth token with account identifier to support multiple GitHub accounts
//
// Parameters:
//   - c *gin.Context: The Gin context containing the HTTP request and response (JWT required in Authorization header)
//
// Response:
//   - Success (200): Returns confirmation of GitHub account linking
//   - Error (400/401): Returns detailed error information if any step fails
//
// The function handles all error cases with appropriate HTTP status codes and messages.
func AddGitHubAccountHandler(c *gin.Context) {
	// Extract user ID from JWT token (user must be already authenticated)
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "User ID not found in token", nil))
		return
	}

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

	// Exchange the authorization code for an access token
	token, err := ExchangeGitHubCodeForTokens(body.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to exchange authorization code for tokens", err.Error()))
		return
	}

	// Exchange the token for user info
	userInfo, err := GetGitHubUserInfo(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to retrieve user information from GitHub", err.Error()))
		return
	}

	// Store the GitHub OAuth token with account identifier (GitHub username) for the authenticated user
	_, err = repositories.CreateOrUpdateGitHubToken(userID, userInfo.Login, token.AccessToken, &token.Expiry, nil, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to store GitHub OAuth token in database", err.Error()))
		return
	}

	// Return success response confirming GitHub account linking
	c.JSON(http.StatusOK, gin.H{
		"message": "GitHub account linked successfully", 
		"success": true, 
		"githubUsername": userInfo.Login,
	})
}

// GetGitHubAccountsHandler lists all connected GitHub accounts for the authenticated user.
//
// This handler performs the following steps:
// 1. Extracts user ID from JWT token (user must be already authenticated)
// 2. Retrieves all GitHub tokens associated with the user from the database
// 3. Returns a list of connected GitHub usernames
//
// Parameters:
//   - c *gin.Context: The Gin context containing the HTTP request and response (JWT required in Authorization header)
//
// Response:
//   - Success (200): Returns list of connected GitHub usernames
//   - Error (401/500): Returns detailed error information if any step fails
func GetGitHubAccountsHandler(c *gin.Context) {
	// Extract user ID from JWT token (user must be already authenticated)
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "User ID not found in token", nil))
		return
	}

	// Get all GitHub tokens for the user
	githubTokens, err := repositories.GetGitHubTokensForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve GitHub accounts", err.Error()))
		return
	}

	// Extract GitHub usernames from the tokens
	var githubUsernames []string
	for _, token := range githubTokens {
		if token.AccountIdentifier != nil {
			githubUsernames = append(githubUsernames, *token.AccountIdentifier)
		}
	}

	// Return the list of GitHub accounts
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"accounts": githubUsernames,
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
