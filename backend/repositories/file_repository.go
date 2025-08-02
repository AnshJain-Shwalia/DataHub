// Package repositories contains database interaction logic for all models
package repositories

import (
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
	// Create file struct with provided data and current timestamp
	file := &models.File{
		ID:        uuid.New().String(),
		Name:      name,
		Size:      size,
		UserID:    userID,
		FolderID:  folderID,
		CreatedAt: time.Now(),
	}

	// Create record in database and return any errors
	return file, db.DB.Create(file).Error
}