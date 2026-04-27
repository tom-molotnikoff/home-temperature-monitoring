package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterAuthRoutes(router gin.IRouter) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", s.loginHandler)
		authGroup.POST("/logout", middleware.AuthRequired(), s.logoutHandler)
		authGroup.GET("/me", middleware.AuthRequired(), s.meHandler)
		authGroup.GET("/sessions", middleware.AuthRequired(), s.listSessionsHandler)
		authGroup.DELETE("/sessions/:id", middleware.AuthRequired(), s.revokeSessionHandler)
	}
}
