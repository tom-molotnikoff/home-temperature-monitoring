package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.Engine) {
	usersGroup := router.Group("/users")
	{
		usersGroup.POST("/", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), createUserHandler)
		usersGroup.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_users"), listUsersHandler)
		usersGroup.PUT("/password", middleware.AuthRequired(), changePasswordHandler)
		usersGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), deleteUserHandler)
		usersGroup.PATCH("/:id/must_change", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), setMustChangeHandler)
		usersGroup.POST("/:id/roles", middleware.AuthRequired(), middleware.RequirePermission("manage_users"), setRolesHandler)
	}
}
