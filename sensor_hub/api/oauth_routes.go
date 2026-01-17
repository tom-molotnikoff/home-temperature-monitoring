package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterOAuthRoutes(router *gin.Engine) {
	oauthGroup := router.Group("/oauth")
	{
		// Status endpoint requires manage_oauth permission
		oauthGroup.GET("/status", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), oauthStatusHandler)

		// Authorize endpoint requires manage_oauth permission
		oauthGroup.GET("/authorize", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), oauthAuthorizeHandler)

		// Submit code endpoint (for out-of-band flow) requires manage_oauth permission
		oauthGroup.POST("/submit-code", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), oauthSubmitCodeHandler)

		// Reload endpoint to re-read credentials from disk
		oauthGroup.POST("/reload", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), oauthReloadHandler)
	}
}
