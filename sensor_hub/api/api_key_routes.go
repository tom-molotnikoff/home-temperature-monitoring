package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterApiKeyRoutes(router gin.IRouter) {
	keysGroup := router.Group("/api-keys")
	{
		keysGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), createApiKeyHandler)
		keysGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), listApiKeysHandler)
		keysGroup.PATCH("/:id/expiry", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), updateApiKeyExpiryHandler)
		keysGroup.POST("/:id/revoke", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), revokeApiKeyHandler)
		keysGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), deleteApiKeyHandler)
	}
}
