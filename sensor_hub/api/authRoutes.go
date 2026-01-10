package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", loginHandler)
		authGroup.POST("/logout", middleware.AuthRequired(), logoutHandler)
		authGroup.GET("/me", middleware.AuthRequired(), meHandler)
		authGroup.GET("/sessions", middleware.AuthRequired(), listSessionsHandler)
		authGroup.DELETE("/sessions/:id", middleware.AuthRequired(), revokeSessionHandler)
	}
}
