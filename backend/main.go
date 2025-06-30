package main

import (
	"strconv"

	"github.com/AnshJain-Shwalia/DataHub/backend/auth"
	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	_, err := db.ConnectDB()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	err = db.AutoMigrate()
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
	router := gin.Default()
	authGroup := router.Group("/auth")
	{
		googleGroup := authGroup.Group("/google")
		{
			googleGroup.GET("/", auth.AuthCodeHandler)
			googleGroup.GET("/oauth-url", auth.GenerateOAuthURLHandler)
		}
	}
	router.Run(":" + strconv.Itoa(cfg.Port))
}
