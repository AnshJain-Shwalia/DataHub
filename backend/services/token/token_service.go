package token

import (
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/AnshJain-Shwalia/DataHub/backend/repositories"
)

// TokenService handles token-related operations
type TokenService struct{}

// NewTokenService creates a new instance of TokenService
func NewTokenService() *TokenService {
	return &TokenService{}
}

// UpsertGoogleToken creates or updates a Google OAuth token for a user
func (s *TokenService) UpsertGoogleToken(userID, accessToken string, expiry *time.Time, refreshToken *string) (*models.Token, error) {
	return repositories.UpsertToken(userID, "GOOGLE", accessToken, expiry, refreshToken, nil)
}

// CreateOrUpdateGitHubToken creates or updates a GitHub OAuth token for a user
func (s *TokenService) CreateOrUpdateGitHubToken(userID, githubUsername, accessToken string, expiry *time.Time, refreshToken *string, refreshTokenExpiry *time.Time) (*models.Token, error) {
	return repositories.CreateOrUpdateGitHubToken(userID, githubUsername, accessToken, expiry, refreshToken, refreshTokenExpiry)
}

// GetGitHubTokensForUser retrieves all GitHub tokens for a user
func (s *TokenService) GetGitHubTokensForUser(userID string) ([]models.Token, error) {
	return repositories.GetGitHubTokensForUser(userID)
}

// GetTokenByID retrieves a token by its ID
func (s *TokenService) GetTokenByID(tokenID string) (*models.Token, error) {
	return repositories.GetTokenByID(tokenID)
}