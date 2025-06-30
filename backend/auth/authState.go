package auth

import (
	"sync"

	"github.com/AnshJain-Shwalia/DataHub/backend/util"
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

// IsValidState checks if a state token exists without consuming it (read-only)
func IsValidState(state string) bool {
	stateRepo.mu.RLock()
	defer stateRepo.mu.RUnlock()
	_, exists := stateRepo.tokens[state]
	return exists
}

// AddState stores a new OAuth2 state token for CSRF protection
func AddState(state string) {
	stateRepo.mu.Lock()
	defer stateRepo.mu.Unlock()
	stateRepo.tokens[state] = struct{}{} // Using empty struct{} as value to minimize memory usage
}

// VerifyAndConsumeState validates and removes a state token (prevents replay attacks)
func VerifyAndConsumeState(state string) bool {
	stateRepo.mu.Lock()
	defer stateRepo.mu.Unlock()
	_, exists := stateRepo.tokens[state]
	if exists {
		delete(stateRepo.tokens, state) // Remove the state token after verification
	}
	return exists
}

// RemoveState deletes a state token (for cleanup of expired/invalid states)
func RemoveState(state string) {
	stateRepo.mu.Lock()
	defer stateRepo.mu.Unlock()
	delete(stateRepo.tokens, state)
}

// GenerateAndAddState creates a secure random state token, stores it, and returns it.
func GenerateAndAddState() (string, error) {
	state, err := util.GenerateRandomState()
	if err != nil {
		return "", err
	}
	AddState(state)
	return state, nil
}
