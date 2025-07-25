package repositories

import (
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/models"
	"github.com/google/uuid"
)

func CreateUser(user *models.User) (*models.User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return user, db.DB.Create(user).Error
}

func FindUserByID(id string) (*models.User, error) {
	var user models.User
	result := db.DB.First(&user, id)
	return &user, result.Error
}

func FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := db.DB.Where("email = ?", email).First(&user)
	return &user, result.Error
}

func CreateUserIfNotPresent(email, name string) (*models.User, error) {
	user, err := FindUserByEmail(email)
	if err != nil {
		return CreateUser(&models.User{Email: email, Name: name})
	}
	return user, nil
}
