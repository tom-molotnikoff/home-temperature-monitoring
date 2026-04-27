package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoleRoutes(router gin.IRouter) {
	roles := router.Group("/roles")
	{
		roles.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), s.listRolesHandler)
		roles.GET("/permissions", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), s.listPermissionsHandler)
		roles.GET("/:id/permissions", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), s.getRolePermissionsHandler)
		roles.POST("/:id/permissions", middleware.AuthRequired(), middleware.RequirePermission("manage_roles"), s.assignPermissionHandler)
		roles.DELETE("/:id/permissions/:pid", middleware.AuthRequired(), middleware.RequirePermission("manage_roles"), s.removePermissionHandler)
	}
}
