package api

import (
	"fmt"
	"log"
	"os"
	"time"

	"example/sensorHub/api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitialiseAndListen() error {
	log.Println("API server is starting...")
	router := gin.Default()

	allowedOrigin := os.Getenv("SENSOR_HUB_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// basic health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// CSRF middleware requires the caller to send X-CSRF-Token header on state-changing requests
	router.Use(middleware.CSRFMiddleware())

	RegisterAuthRoutes(router)
	RegisterUserRoutes(router)
	RegisterRoleRoutes(router)
	RegisterTemperatureRoutes(router)
	RegisterSensorRoutes(router)
	RegisterPropertiesRoutes(router)

	log.Println("API server is running on port 8080")
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	return nil
}
