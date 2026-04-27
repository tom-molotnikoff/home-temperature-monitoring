package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterUserRoutes(router gin.IRouter) {
	usersGroup := router.Group("/users")
	{
		usersGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), s.createUserHandler)
		usersGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_users"), s.listUsersHandler)
		usersGroup.PUT("/password", middleware.AuthRequired(), s.changePasswordHandler)
		usersGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), s.deleteUserHandler)
		usersGroup.PATCH("/:id/must_change", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), s.setMustChangeHandler)
		usersGroup.POST("/:id/roles", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), s.setRolesHandler)
	}
}
