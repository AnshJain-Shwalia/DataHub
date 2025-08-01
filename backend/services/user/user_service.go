package user

import (
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/AnshJain-Shwalia/DataHub/backend/repositories"
)

// UserService handles user-related operations
type UserService struct{}

// NewUserService creates a new instance of UserService
func NewUserService() *UserService {
	return &UserService{}
}

// FindByID retrieves a user by their ID
func (s *UserService) FindByID(userID string) (*models.User, error) {
	return repositories.FindUserByID(userID)
}

// FindByEmail retrieves a user by their email address
func (s *UserService) FindByEmail(email string) (*models.User, error) {
	return repositories.FindUserByEmail(email)
}

// CreateIfNotPresent finds a user by email or creates a new one if not found
func (s *UserService) CreateIfNotPresent(email, name string) (*models.User, error) {
	return repositories.CreateUserIfNotPresent(email, name)
}

// Create creates a new user
func (s *UserService) Create(user *models.User) (*models.User, error) {
	return repositories.CreateUser(user)
}