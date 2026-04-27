package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterDashboardRoutes(router gin.IRouter) {
	group := router.Group("/dashboards")
	{
		group.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_dashboards"), s.listDashboardsHandler)
		group.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), s.createDashboardHandler)
		group.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_dashboards"), s.getDashboardHandler)
		group.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), s.updateDashboardHandler)
		group.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), s.deleteDashboardHandler)
		group.POST("/:id/share", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), s.shareDashboardHandler)
		group.PUT("/:id/default", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), s.setDefaultDashboardHandler)
	}
}
