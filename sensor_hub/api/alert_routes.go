package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAlertRoutes(router *gin.Engine) {
	alertsGroup := router.Group("/alerts")
	{
		// View alert rules and history
		alertsGroup.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAllAlertRulesHandler)
		alertsGroup.GET("/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertRuleBySensorIDHandler)
		alertsGroup.GET("/:sensorId/history", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertHistoryHandler)

		// Manage alert rules (create, update, delete)
		alertsGroup.POST("/", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), createAlertRuleHandler)
		alertsGroup.PUT("/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), updateAlertRuleHandler)
		alertsGroup.DELETE("/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), deleteAlertRuleHandler)
	}
}
