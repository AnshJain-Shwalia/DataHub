// Package repositories contains database interaction logic for all models
package repositories

import (
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateToken creates a new authentication token in the database
// It takes user identification and token details as parameters and returns the created token
//
// Parameters:
//   - userID: The ID of the user this token belongs to
//   - platform: The platform/provider of the token (e.g., "GOOGLE", "GITHUB")
//   - accessToken: The actual access token string
//   - accessTokenExpiry: Pointer to the expiration time of the access token
//   - refreshToken: Optional pointer to refresh token string (can be nil)
//   - refreshTokenExpiry: Optional pointer to refresh token expiration time (can be nil)
//
// Returns:
//   - A pointer to the created Token model
//   - An error if the database operation fails
func CreateToken(
	userID string,
	platform string,
	accessToken string,
	accessTokenExpiry *time.Time,
	refreshToken *string,
	refreshTokenExpiry *time.Time) (*models.Token, error) {
	return CreateTokenWithAccountIdentifier(userID, platform, nil, accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry)
}

// CreateTokenWithAccountIdentifier creates a new authentication token with account identifier
// This is used for platforms that support multiple accounts per user (like GitHub)
//
// Parameters:
//   - userID: The ID of the user this token belongs to
//   - platform: The platform/provider of the token (e.g., "GOOGLE", "GITHUB")
//   - accountIdentifier: Optional account identifier (GitHub username, Google email, etc.)
//   - accessToken: The actual access token string
//   - accessTokenExpiry: Pointer to the expiration time of the access token
//   - refreshToken: Optional pointer to refresh token string (can be nil)
//   - refreshTokenExpiry: Optional pointer to refresh token expiration time (can be nil)
//
// Returns:
//   - A pointer to the created Token model
//   - An error if the database operation fails
func CreateTokenWithAccountIdentifier(
	userID string,
	platform string,
	accountIdentifier *string,
	accessToken string,
	accessTokenExpiry *time.Time,
	refreshToken *string,
	refreshTokenExpiry *time.Time) (*models.Token, error) {
	// Get current time for timestamps
	now := time.Now()
	
	// Only set refresh token issuance time if a refresh token exists
	var refreshIssuedAt *time.Time
	if refreshToken != nil {
		refreshIssuedAt = &now
	}
	
	// Create token struct with provided data and current timestamps
	token := &models.Token{
		ID:                   uuid.New().String(),
		UserID:               userID,
		Platform:             platform,
		AccountIdentifier:    accountIdentifier,
		AccessToken:          accessToken,
		AccessTokenExpiry:    accessTokenExpiry,
		RefreshToken:         refreshToken,
		RefreshTokenExpiry:   refreshTokenExpiry,
		AccessTokenIssuedAt:  now,
		RefreshTokenIssuedAt: refreshIssuedAt,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	
	// Create record in database and return any errors
	return token, db.DB.Create(token).Error
}

// UpdateToken updates an existing token in the database
// It takes the same parameters as CreateToken but updates an existing record
//
// Parameters:
//   - existingToken: The existing token to update
//   - accessToken: The actual access token string
//   - accessTokenExpiry: Pointer to the expiration time of the access token
//   - refreshToken: Optional pointer to refresh token string (can be nil)
//   - refreshTokenExpiry: Optional pointer to refresh token expiration time (can be nil)
//
// Returns:
//   - A pointer to the updated Token model
//   - An error if the database operation fails
func UpdateToken(
	existingToken *models.Token,
	accessToken string,
	accessTokenExpiry *time.Time,
	refreshToken *string,
	refreshTokenExpiry *time.Time) (*models.Token, error) {
	// Get current time for timestamps
	now := time.Now()
	
	// Only set refresh token issuance time if a refresh token exists
	var refreshIssuedAt *time.Time
	if refreshToken != nil {
		refreshIssuedAt = &now
	}
	
	// Update the existing token with new values
	existingToken.AccessToken = accessToken
	existingToken.AccessTokenExpiry = accessTokenExpiry
	existingToken.AccessTokenIssuedAt = now
	existingToken.RefreshToken = refreshToken
	existingToken.RefreshTokenExpiry = refreshTokenExpiry
	existingToken.RefreshTokenIssuedAt = refreshIssuedAt
	existingToken.UpdatedAt = now
	
	// Save the updated token
	err := db.DB.Save(existingToken).Error
	return existingToken, err
}

// UpsertToken creates a new token or updates an existing one in the database
// For Google, this behaves like the old behavior (single token per user)
// For GitHub, this should NOT be used - use CreateOrUpdateGitHubToken instead
//
// Parameters:
//   - userID: The ID of the user this token belongs to
//   - platform: The platform/provider of the token (e.g., "GOOGLE", "GITHUB")
//   - accessToken: The actual access token string
//   - accessTokenExpiry: Pointer to the expiration time of the access token
//   - refreshToken: Optional pointer to refresh token string (can be nil)
//   - refreshTokenExpiry: Optional pointer to refresh token expiration time (can be nil)
//
// Returns:
//   - A pointer to the created or updated Token model
//   - An error if the database operation fails
func UpsertToken(
	userID string,
	platform string,
	accessToken string,
	accessTokenExpiry *time.Time,
	refreshToken *string,
	refreshTokenExpiry *time.Time) (*models.Token, error) {
	// Check if a token already exists for this user and platform
	var existingToken models.Token
	result := db.DB.Where("user_id = ? AND platform = ?", userID, platform).First(&existingToken)
	
	// If token exists, update it
	if result.Error == nil {
		return UpdateToken(&existingToken, accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry)
	} else if result.Error == gorm.ErrRecordNotFound {
		// If token doesn't exist, create a new one
		return CreateToken(userID, platform, accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry)
	}
	
	// Return any other database errors
	return nil, result.Error
}

// CreateOrUpdateGitHubToken creates a new GitHub token or updates an existing one for the same GitHub account
// This function ensures that each user can have multiple GitHub tokens but only one per unique GitHub account
//
// Parameters:
//   - userID: The ID of the user this token belongs to
//   - githubUsername: The GitHub username/login (used as account identifier)
//   - accessToken: The actual access token string
//   - accessTokenExpiry: Pointer to the expiration time of the access token
//   - refreshToken: Optional pointer to refresh token string (can be nil)
//   - refreshTokenExpiry: Optional pointer to refresh token expiration time (can be nil)
//
// Returns:
//   - A pointer to the created or updated Token model
//   - An error if the database operation fails
func CreateOrUpdateGitHubToken(
	userID string,
	githubUsername string,
	accessToken string,
	accessTokenExpiry *time.Time,
	refreshToken *string,
	refreshTokenExpiry *time.Time) (*models.Token, error) {
	
	// Check if a GitHub token already exists for this user and GitHub account
	var existingToken models.Token
	result := db.DB.Where("user_id = ? AND platform = ? AND account_identifier = ?", userID, "GITHUB", githubUsername).First(&existingToken)
	
	// If token exists for this GitHub account, update it
	if result.Error == nil {
		return UpdateToken(&existingToken, accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry)
	} else if result.Error == gorm.ErrRecordNotFound {
		// If token doesn't exist for this GitHub account, create a new one
		return CreateTokenWithAccountIdentifier(userID, "GITHUB", &githubUsername, accessToken, accessTokenExpiry, refreshToken, refreshTokenExpiry)
	}
	
	// Return any other database errors
	return nil, result.Error
}

// GetGitHubTokensForUser retrieves all GitHub tokens for a specific user
// This is useful for showing all connected GitHub accounts
//
// Parameters:
//   - userID: The ID of the user
//
// Returns:
//   - A slice of Token models for all GitHub accounts
//   - An error if the database operation fails
func GetGitHubTokensForUser(userID string) ([]models.Token, error) {
	var tokens []models.Token
	err := db.DB.Where("user_id = ? AND platform = ?", userID, "GITHUB").Find(&tokens).Error
	return tokens, err
}
