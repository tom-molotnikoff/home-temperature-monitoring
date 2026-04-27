package api

import (
	"example/sensorHub/api/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterAuthRoutes(router gin.IRouter) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", s.Login)
		authGroup.POST("/logout", middleware.AuthRequired(), s.Logout)
		authGroup.GET("/me", middleware.AuthRequired(), s.GetCurrentUser)
		authGroup.GET("/sessions", middleware.AuthRequired(), s.ListSessions)
		authGroup.DELETE("/sessions/:id", middleware.AuthRequired(), func(c *gin.Context) {
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid session id"})
				return
			}
			s.RevokeSession(c, id)
		})
	}
}
