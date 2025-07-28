// Package repositories contains database interaction logic for all models
package repositories

import (
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/google/uuid"
)

// CreateUser creates a new user in the database
// It takes a pointer to a user model, assigns a UUID and timestamps, and persists it
//
// Parameters:
//   - user: Pointer to the user model to be created
//
// Returns:
//   - A pointer to the created user model (with ID and timestamps populated)
//   - An error if the database operation fails
func CreateUser(user *models.User) (*models.User, error) {
	// Generate a new UUID for the user
	user.ID = uuid.New().String()

	// Set creation and update timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Create record in database and return any errors
	return user, db.DB.Create(user).Error
}

// FindUserByID retrieves a user from the database by their ID
//
// Parameters:
//   - id: The unique identifier of the user to find
//
// Returns:
//   - A pointer to the found user model
//   - An error if the user is not found or if the database operation fails
func FindUserByID(id string) (*models.User, error) {
	var user models.User

	// Query the database for a user with the given ID
	result := db.DB.First(&user, id)

	return &user, result.Error
}

// FindUserByEmail retrieves a user from the database by their email address
//
// Parameters:
//   - email: The email address of the user to find
//
// Returns:
//   - A pointer to the found user model
//   - An error if the user is not found or if the database operation fails
func FindUserByEmail(email string) (*models.User, error) {
	var user models.User

	// Query the database for a user with the given email
	result := db.DB.Where("email = ?", email).First(&user)

	return &user, result.Error
}

// CreateUserIfNotPresent finds a user by email or creates a new one if not found
// This is useful for OAuth flows where users might be logging in for the first time
//
// Parameters:
//   - email: The email address to search for or use in creating a new user
//   - name: The name to use if creating a new user
//
// Returns:
//   - A pointer to the found or newly created user model
//   - An error if the database operation fails
func CreateUserIfNotPresent(email, name string) (*models.User, error) {
	// Try to find the user by email first
	user, err := FindUserByEmail(email)

	// If user not found, create a new one
	if err != nil {
		return CreateUser(&models.User{Email: email, Name: name})
	}

	// Return the existing user
	return user, nil
}
