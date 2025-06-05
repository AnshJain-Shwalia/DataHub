package main

import (
	"log"
	"strconv"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})
	router.Run(":" + strconv.Itoa(cfg.Port))
}
