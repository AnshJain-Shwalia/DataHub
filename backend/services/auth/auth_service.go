package auth

import (
	"sync"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/AnshJain-Shwalia/DataHub/backend/repositories"
	"github.com/AnshJain-Shwalia/DataHub/backend/util"
	"github.com/golang-jwt/jwt/v5"
)

// stateRepo is the singleton instance for managing OAuth2 state tokens
var stateRepo = &StateRepo{
	tokens: make(map[string]struct{}),
}

// StateRepo provides thread-safe storage for OAuth2 state tokens
type StateRepo struct {
	mu     sync.RWMutex        // Protects tokens map
	tokens map[string]struct{} // Active state tokens as map keys
}

// AuthService handles core authentication operations
type AuthService struct{}

// NewAuthService creates a new instance of AuthService
func NewAuthService() *AuthService {
	return &AuthService{}
}

// GenerateJWT generates a JWT token for the given user
// This method extracts the JWT generation logic from the handler
func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
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

// GetTokenByUserID finds a user by ID and generates a JWT token
// This method extracts the logic from GetSigninTokenByIDHandler
func (s *AuthService) GetTokenByUserID(userID string) (string, *models.User, error) {
	// Find user by ID
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		return "", nil, err
	}

	// Generate JWT token for the user
	tokenString, err := s.GenerateJWT(user)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

// OAuth State Management Functions (merged from oauth package)

// isValidState checks if a state token exists without consuming it (read-only)
func isValidState(state string) bool {
	stateRepo.mu.RLock()
	defer stateRepo.mu.RUnlock()
	_, exists := stateRepo.tokens[state]
	return exists
}

// addState stores a new OAuth2 state token for CSRF protection
func addState(state string) {
	stateRepo.mu.Lock()
	defer stateRepo.mu.Unlock()
	stateRepo.tokens[state] = struct{}{} // Using empty struct{} as value to minimize memory usage
}

// verifyAndConsumeState validates and removes a state token (prevents replay attacks)
func verifyAndConsumeState(state string) bool {
	stateRepo.mu.Lock()
	defer stateRepo.mu.Unlock()
	_, exists := stateRepo.tokens[state]
	if exists {
		delete(stateRepo.tokens, state) // Remove the state token after verification
	}
	return exists
}

// removeState deletes a state token (for cleanup of expired/invalid states)
func removeState(state string) {
	stateRepo.mu.Lock()
	defer stateRepo.mu.Unlock()
	delete(stateRepo.tokens, state)
}

// generateAndAddState creates a secure random state token, stores it, and returns it.
func generateAndAddState() (string, error) {
	state, err := util.GenerateRandomState()
	if err != nil {
		return "", err
	}
	addState(state)
	return state, nil
}