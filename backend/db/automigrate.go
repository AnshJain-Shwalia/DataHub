package db

import "github.com/AnshJain-Shwalia/DataHub/backend/models"

func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Token{},
		&models.Repo{},
		&models.Branch{},
		&models.Folder{},
		&models.File{},
		&models.Chunk{},
	)
}
