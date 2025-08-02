// Package repositories contains database interaction logic for all models
package repositories

import (
	"fmt"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/google/uuid"
)

// CreateFile creates a new file in the database
// It takes file details as parameters and returns the created file
//
// Parameters:
//   - name: The name of the file
//   - size: The size of the file in bytes
//   - userID: The ID of the user who owns this file
//   - folderID: Optional pointer to the parent folder ID (can be nil for root files)
//
// Returns:
//   - A pointer to the created File model (with ID and timestamps populated)
//   - An error if the database operation fails
func CreateFile(name string, size int64, userID string, folderID *string) (*models.File, error) {
	// Input validation
	if name == "" || userID == "" {
		return nil, fmt.Errorf("name and userID are required")
	}
	if size < 0 {
		return nil, fmt.Errorf("size cannot be negative")
	}

	// Database check
	if db.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Create file struct with provided data and current timestamp
	file := &models.File{
		ID:        uuid.New().String(),
		Name:      name,
		Size:      size,
		UserID:    userID,
		FolderID:  folderID,
		CreatedAt: time.Now(),
	}

	// Create record in database with error context
	if err := db.DB.Create(file).Error; err != nil {
		return nil, fmt.Errorf("failed to create file (name: %s, userID: %s): %w", name, userID, err)
	}
	return file, nil
}