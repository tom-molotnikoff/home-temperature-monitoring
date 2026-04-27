package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterOAuthRoutes(router gin.IRouter) {
	oauthGroup := router.Group("/oauth")
	{
		// Status endpoint requires manage_oauth permission
		oauthGroup.GET("/status", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.oauthStatusHandler)

		// Authorize endpoint requires manage_oauth permission
		oauthGroup.GET("/authorize", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.oauthAuthorizeHandler)

		// Submit code endpoint (for out-of-band flow) requires manage_oauth permission
		oauthGroup.POST("/submit-code", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.oauthSubmitCodeHandler)

		// Reload endpoint to re-read credentials from disk
		oauthGroup.POST("/reload", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.oauthReloadHandler)
	}
}
