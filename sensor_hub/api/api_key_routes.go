package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterApiKeyRoutes(router gin.IRouter) {
	keysGroup := router.Group("/api-keys")
	{
		keysGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.createApiKeyHandler)
		keysGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.listApiKeysHandler)
		keysGroup.PATCH("/:id/expiry", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.updateApiKeyExpiryHandler)
		keysGroup.POST("/:id/revoke", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.revokeApiKeyHandler)
		keysGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.deleteApiKeyHandler)
	}
}
