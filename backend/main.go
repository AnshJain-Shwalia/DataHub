package main

import (
	"log"
	"strconv"

	"github.com/AnshJain-Shwalia/DataHub/backend/auth"
	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/AnshJain-Shwalia/DataHub/backend/db"
	"github.com/AnshJain-Shwalia/DataHub/backend/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting DataHub backend server...")
	
	cfg := config.LoadConfig()
	log.Printf("Configuration loaded - Port: %d", cfg.Port)
	
	log.Println("Connecting to database...")
	_, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established")
	
	log.Println("Running database migrations...")
	err = db.AutoMigrate()
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrations completed")
	
	log.Println("Setting up routes...")
	router := gin.Default()
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "DataHub backend is running",
		})
	})
	
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
	
	log.Printf("Server starting on port %d...", cfg.Port)
	log.Printf("Server running at http://localhost:%d", cfg.Port)
	
	if err := router.Run(":" + strconv.Itoa(cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
