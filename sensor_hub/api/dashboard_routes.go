package api

import (
	"net/http"
	"strconv"

	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterDashboardRoutes(router gin.IRouter) {
	group := router.Group("/dashboards")
	{
		group.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_dashboards"), s.ListDashboards)
		group.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), s.CreateDashboard)
		group.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_dashboards"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
				return
			}
			s.GetDashboard(c, id)
		})
		group.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
				return
			}
			s.UpdateDashboard(c, id)
		})
		group.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
				return
			}
			s.DeleteDashboard(c, id)
		})
		group.POST("/:id/share", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
				return
			}
			s.ShareDashboard(c, id)
		})
		group.PUT("/:id/default", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
				return
			}
			s.SetDefaultDashboard(c, id)
		})
	}
}
