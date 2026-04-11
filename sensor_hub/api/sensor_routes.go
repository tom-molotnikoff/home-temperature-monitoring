package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSensorRoutes(router gin.IRouter) {
	sensorsGroup := router.Group("/sensors")
	{
		sensorsGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), addSensorHandler)
		sensorsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), updateSensorHandler)
		sensorsGroup.DELETE("/:name", middleware.AuthRequired(), middleware.RequirePermission("delete_sensors"), deleteSensorHandler)
		sensorsGroup.GET("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), getSensorByNameHandler)
		sensorsGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), getAllSensorsHandler)
		sensorsGroup.GET("/driver/:driver", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), getSensorsByDriverHandler)
		sensorsGroup.HEAD("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), sensorExistsHandler)
		sensorsGroup.POST("/collect", middleware.AuthRequired(), middleware.RequirePermission("trigger_readings"), collectAndStoreAllSensorReadingsHandler)
		sensorsGroup.POST("/collect/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("trigger_readings"), collectFromSensorByNameHandler)
		sensorsGroup.POST("/disable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), disableSensorHandler)
		sensorsGroup.POST("/enable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), enableSensorHandler)
		sensorsGroup.GET("/ws/:driver", middleware.AuthRequired(), sensorWebSocketHandler)
		sensorsGroup.GET("/health/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), getSensorHealthHistoryByNameHandler)
		sensorsGroup.GET("/stats/total-readings", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), totalReadingsPerSensorHandler)
		sensorsGroup.GET("/status/:status", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), getSensorsByStatusHandler)
		sensorsGroup.POST("/approve/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), approveSensorHandler)
		sensorsGroup.POST("/dismiss/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), dismissSensorHandler)
	}
}
