package api

import (
	"example/sensorHub/api/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterUserRoutes(router gin.IRouter) {
	usersGroup := router.Group("/users")
	{
		usersGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), s.CreateUser)
		usersGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_users"), s.ListUsers)
		usersGroup.PUT("/password", middleware.AuthRequired(), s.ChangePassword)
		usersGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
				return
			}
			s.DeleteUser(c, id)
		})
		usersGroup.PATCH("/:id/must_change", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
				return
			}
			s.SetMustChangePassword(c, id)
		})
		usersGroup.POST("/:id/roles", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
				return
			}
			s.SetUserRoles(c, id)
		})
	}
}
