package api

import (
	"net/http"
	"strconv"

	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterApiKeyRoutes(router gin.IRouter) {
	keysGroup := router.Group("/api-keys")
	{
		keysGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.CreateApiKey)
		keysGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), s.ListApiKeys)
		keysGroup.PATCH("/:id/expiry", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
				return
			}
			s.UpdateApiKeyExpiry(c, id)
		})
		keysGroup.POST("/:id/revoke", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
				return
			}
			s.RevokeApiKey(c, id)
		})
		keysGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_api_keys"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid key id"})
				return
			}
			s.DeleteApiKey(c, id)
		})
	}
}
