package main

import (
	"strconv"

	"github.com/AnshJain-Shwalia/DataHub/backend/auth"
	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
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
