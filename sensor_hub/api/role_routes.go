package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(router *gin.Engine) {
	roles := router.Group("/roles")
	{
		roles.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), listRolesHandler)
		roles.GET("/permissions", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), listPermissionsHandler)
		roles.GET("/:id/permissions", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), getRolePermissionsHandler)
		roles.POST("/:id/permissions", middleware.AuthRequired(), middleware.RequirePermission("manage_roles"), assignPermissionHandler)
		roles.DELETE("/:id/permissions/:pid", middleware.AuthRequired(), middleware.RequirePermission("manage_roles"), removePermissionHandler)
	}
}
