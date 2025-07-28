// Package repositories contains database interaction logic for all models
package repositories

import (
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
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
	// Get current time for timestamps
	now := time.Now()
	
	// Only set refresh token issuance time if a refresh token exists
	var refreshIssuedAt *time.Time
	if refreshToken != nil {
		refreshIssuedAt = &now
	}
	
	// Create token struct with provided data and current timestamps
	token := &models.Token{
		UserID:               userID,
		Platform:             platform,
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
// It takes the same parameters as CreateToken and either creates a new record
// or updates an existing one based on userID and platform
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
