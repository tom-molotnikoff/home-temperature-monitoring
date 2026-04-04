package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(router gin.IRouter) {
	group := router.Group("/dashboards")
	{
		group.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_dashboards"), listDashboardsHandler)
		group.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), createDashboardHandler)
		group.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_dashboards"), getDashboardHandler)
		group.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), updateDashboardHandler)
		group.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), deleteDashboardHandler)
		group.POST("/:id/share", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), shareDashboardHandler)
		group.PUT("/:id/default", middleware.AuthRequired(), middleware.RequirePermission("manage_dashboards"), setDefaultDashboardHandler)
	}
}
