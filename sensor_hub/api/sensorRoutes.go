package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSensorRoutes(router *gin.Engine) {
	sensorsGroup := router.Group("/sensors")
	{
		sensorsGroup.POST("/", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), addSensorHandler)
		sensorsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), updateSensorHandler)
		sensorsGroup.DELETE("/:name", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), deleteSensorHandler)
		sensorsGroup.GET("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getSensorByNameHandler)
		sensorsGroup.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getAllSensorsHandler)
		sensorsGroup.GET("/type/:type", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getSensorsByTypeHandler)
		sensorsGroup.HEAD("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), sensorExistsHandler)
		sensorsGroup.POST("/collect", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), collectAndStoreAllSensorReadingsHandler)
		sensorsGroup.POST("/collect/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), collectFromSensorByNameHandler)
		sensorsGroup.POST("/disable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), disableSensorHandler)
		sensorsGroup.POST("/enable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), enableSensorHandler)
		sensorsGroup.GET("/ws/:type", middleware.AuthRequired(), sensorWebSocketHandler)
		sensorsGroup.GET("/health/:name", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getSensorHealthHistoryByNameHandler)
		sensorsGroup.GET("/stats/total-readings", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), totalReadingsPerSensorHandler)
	}
}
