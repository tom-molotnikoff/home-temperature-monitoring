package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterOAuthRoutes(router gin.IRouter) {
	oauthGroup := router.Group("/oauth")
	{
		oauthGroup.GET("/status", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.GetOAuthStatus)
		oauthGroup.GET("/authorize", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.GetOAuthAuthorizeUrl)
		oauthGroup.POST("/submit-code", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.SubmitOAuthCode)
		oauthGroup.POST("/reload", middleware.AuthRequired(), middleware.RequirePermission("manage_oauth"), s.ReloadOAuth)
	}
}
