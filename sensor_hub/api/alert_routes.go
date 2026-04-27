package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterAlertRoutes(router gin.IRouter) {
	alertsGroup := router.Group("/alerts")
	{
		// List all alert rules
		alertsGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), s.getAllAlertRulesHandler)

		// Alert rules by sensor
		alertsGroup.GET("/sensor/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), s.getAlertRulesBySensorIDHandler)
		alertsGroup.GET("/sensor/:sensorId/history", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), s.getAlertHistoryHandler)

		// Individual alert rule CRUD
		alertsGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), s.createAlertRuleHandler)
		alertsGroup.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), s.getAlertRuleByIDHandler)
		alertsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), s.updateAlertRuleHandler)
		alertsGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), s.deleteAlertRuleHandler)
	}
}
