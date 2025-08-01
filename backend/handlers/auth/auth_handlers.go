package auth

import (
	"log"
	"net/http"

	"github.com/AnshJain-Shwalia/DataHub/backend/http_util"
	"github.com/AnshJain-Shwalia/DataHub/backend/middleware"
	authservice "github.com/AnshJain-Shwalia/DataHub/backend/services/auth"
	"github.com/gin-gonic/gin"
)


// GoogleAuthCodeHandler processes the OAuth2 authorization code received from Google's OAuth flow.
// The actual business logic is handled by the GoogleAuthService.
func GoogleAuthCodeHandler(c *gin.Context) {
	var body authservice.ProcessAuthCodeRequest
	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Incorrect body structure", err.Error()))
		return
	}

	googleAuthService := authservice.NewGoogleAuthService()
	response, err := googleAuthService.ProcessAuthCode(&body)
	if err != nil {
		if authErr, ok := err.(*authservice.AuthError); ok {
			c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, authErr.Message, authErr.Details))
			return
		}
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Authentication failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateGoogleOAuthURLHandler generates the OAuth URL for Google login
func GenerateGoogleOAuthURLHandler(c *gin.Context) {
	googleAuthService := authservice.NewGoogleAuthService()
	authURL, err := googleAuthService.GenerateOAuthURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authURL": authURL,
		"success": true,
	})
}

// AddGitHubAccountHandler processes the OAuth2 authorization code received from GitHub's OAuth flow
// to link a GitHub storage account to an already authenticated user.
// The actual business logic is handled by the GitHubAuthService.
func AddGitHubAccountHandler(c *gin.Context) {
	// Extract user ID from JWT token (user must be already authenticated)
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "User ID not found in token", nil))
		return
	}

	var body authservice.AddAccountRequest
	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Incorrect body structure", err.Error()))
		return
	}

	githubAuthService := authservice.NewGitHubAuthService()
	response, err := githubAuthService.AddAccount(userID, &body)
	if err != nil {
		if authErr, ok := err.(*authservice.AuthError); ok {
			c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, authErr.Message, authErr.Details))
			return
		}
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "Failed to add GitHub account", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetGitHubAccountsHandler lists all connected GitHub accounts for the authenticated user.
// The actual business logic is handled by the GitHubAuthService.
func GetGitHubAccountsHandler(c *gin.Context) {
	// Extract user ID from JWT token (user must be already authenticated)
	userID := middleware.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, http_util.NewErrorResponse(http.StatusUnauthorized, "User ID not found in token", nil))
		return
	}

	githubAuthService := authservice.NewGitHubAuthService()
	response, err := githubAuthService.GetAccounts(userID)
	if err != nil {
		if authErr, ok := err.(*authservice.AuthError); ok {
			c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, authErr.Message, authErr.Details))
			return
		}
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve GitHub accounts", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateGitHubOAuthURLHandler generates the OAuth URL for GitHub login
func GenerateGitHubOAuthURLHandler(c *gin.Context) {
	githubAuthService := authservice.NewGitHubAuthService()
	authURL, err := githubAuthService.GenerateOAuthURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, http_util.NewErrorResponse(http.StatusInternalServerError, "", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authURL": authURL,
		"success": true,
	})
}

// TTBD: Temporary endpoint for development - generates JWT token by user ID without authentication
// This endpoint bypasses normal OAuth flow and should be removed before production
func GetSigninTokenByIDHandler(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, http_util.NewErrorResponse(http.StatusBadRequest, "User ID is required", nil))
		return
	}

	authService := authservice.NewAuthService()
	tokenString, user, err := authService.GetTokenByUserID(userID)
	if err != nil {
		log.Printf("Failed to generate token for user %s: %v", userID, err)
		c.JSON(http.StatusNotFound, http_util.NewErrorResponse(http.StatusNotFound, "User not found", nil))
		return
	}

	// Return the JWT token
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   tokenString,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

