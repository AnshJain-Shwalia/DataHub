package main

import (
	"strconv"

	"github.com/AnshJain-Shwalia/DataHub/backend/auth"
	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/middleware"
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
		// Google OAuth routes
		googleGroup := authGroup.Group("/google")
		{
			googleGroup.POST("/", auth.GoogleAuthCodeHandler)
			googleGroup.GET("/oauth-url", auth.GenerateGoogleOAuthURLHandler)
		}

		// GitHub OAuth routes (require authentication for storage account management)
		githubGroup := authGroup.Group("/github")
		{
			// Protected routes that require JWT authentication
			githubGroup.Use(middleware.RequireJWT())
			githubGroup.POST("/accounts", auth.AddGitHubAccountHandler)
			githubGroup.GET("/accounts", auth.GetGitHubAccountsHandler)
			githubGroup.GET("/oauth-url", auth.GenerateGitHubOAuthURLHandler)
		}
	}
	router.Run(":" + strconv.Itoa(cfg.Port))
}
