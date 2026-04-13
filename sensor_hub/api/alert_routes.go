package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAlertRoutes(router gin.IRouter) {
	alertsGroup := router.Group("/alerts")
	{
		// List all alert rules
		alertsGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAllAlertRulesHandler)

		// Alert rules by sensor
		alertsGroup.GET("/sensor/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertRulesBySensorIDHandler)
		alertsGroup.GET("/sensor/:sensorId/history", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertHistoryHandler)

		// Individual alert rule CRUD
		alertsGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), createAlertRuleHandler)
		alertsGroup.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertRuleByIDHandler)
		alertsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), updateAlertRuleHandler)
		alertsGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), deleteAlertRuleHandler)
	}
}
