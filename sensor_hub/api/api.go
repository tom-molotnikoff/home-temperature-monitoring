package api

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitialiseAndListen() error {
	log.Println("API server is starting...")
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // For development, allow all
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	RegisterTemperatureRoutes(router)

	log.Println("API server is running on port 8080")
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	return nil
}
