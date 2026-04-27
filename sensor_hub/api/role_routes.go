package api

import (
	"fmt"
	"net/http"

	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoleRoutes(router gin.IRouter) {
	roles := router.Group("/roles")
	{
		roles.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), s.ListRoles)
		roles.GET("/permissions", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), s.ListPermissions)
		roles.GET("/:id/permissions", middleware.AuthRequired(), middleware.RequirePermission("view_roles"), func(c *gin.Context) {
			var id int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
				return
			}
			s.GetRolePermissions(c, id)
		})
		roles.POST("/:id/permissions", middleware.AuthRequired(), middleware.RequirePermission("manage_roles"), func(c *gin.Context) {
			var id int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
				return
			}
			s.AssignPermission(c, id)
		})
		roles.DELETE("/:id/permissions/:pid", middleware.AuthRequired(), middleware.RequirePermission("manage_roles"), func(c *gin.Context) {
			var id, pid int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid role id"})
				return
			}
			if _, err := fmt.Sscan(c.Param("pid"), &pid); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid permission id"})
				return
			}
			s.RemovePermission(c, id, pid)
		})
	}
}

