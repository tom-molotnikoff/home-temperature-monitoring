package api

import (
	"fmt"
	"log"
	"os"
	"time"

	"example/sensorHub/api/middleware"
	"example/sensorHub/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitialiseAndListen() error {
	log.Println("API server is starting...")
	router := gin.Default()

	// CORS is only needed when the UI is served from a different origin (e.g. Vite dev server)
	allowedOrigin := os.Getenv("SENSOR_HUB_ALLOWED_ORIGIN")
	if allowedOrigin != "" {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{allowedOrigin},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
			ExposeHeaders:    []string{"Content-Length", "Retry-After"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	// All API routes live under /api
	apiGroup := router.Group("/api")

	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	apiGroup.Use(middleware.CSRFMiddleware())

	RegisterAuthRoutes(apiGroup)
	RegisterUserRoutes(apiGroup)
	RegisterRoleRoutes(apiGroup)
	RegisterTemperatureRoutes(apiGroup)
	RegisterSensorRoutes(apiGroup)
	RegisterPropertiesRoutes(apiGroup)
	RegisterAlertRoutes(apiGroup)
	RegisterOAuthRoutes(apiGroup)
	RegisterNotificationRoutes(apiGroup)

	// Serve embedded UI for all non-API routes
	web.RegisterSPAHandler(router)

	log.Println("API server is running on port 8080")

	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Printf("Starting API server with TLS (cert=%s key=%s)", certFile, keyFile)
		if err := router.RunTLS("0.0.0.0:8080", certFile, keyFile); err != nil {
			return fmt.Errorf("failed to start TLS API server: %w", err)
		}
		return nil
	}

	if err := router.Run("0.0.0.0:8080"); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	return nil
}
