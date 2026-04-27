package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterSensorRoutes(router gin.IRouter) {
	sensorsGroup := router.Group("/sensors")
	{
		sensorsGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.addSensorHandler)
		sensorsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.updateSensorHandler)
		sensorsGroup.DELETE("/:name", middleware.AuthRequired(), middleware.RequirePermission("delete_sensors"), s.deleteSensorHandler)
		sensorsGroup.GET("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.getSensorByNameHandler)
		sensorsGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.getAllSensorsHandler)
		sensorsGroup.GET("/driver/:driver", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.getSensorsByDriverHandler)
		sensorsGroup.HEAD("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.sensorExistsHandler)
		sensorsGroup.POST("/collect", middleware.AuthRequired(), middleware.RequirePermission("trigger_readings"), s.collectAndStoreAllSensorReadingsHandler)
		sensorsGroup.POST("/collect/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("trigger_readings"), s.collectFromSensorByNameHandler)
		sensorsGroup.POST("/disable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.disableSensorHandler)
		sensorsGroup.POST("/enable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.enableSensorHandler)
		sensorsGroup.GET("/ws", middleware.AuthRequired(), s.allSensorsWebSocketHandler)
		sensorsGroup.GET("/ws/:driver", middleware.AuthRequired(), s.sensorWebSocketHandler)
		sensorsGroup.GET("/health/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.getSensorHealthHistoryByNameHandler)
		sensorsGroup.GET("/stats/total-readings", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.totalReadingsPerSensorHandler)
		sensorsGroup.GET("/status/:status", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.getSensorsByStatusHandler)
		sensorsGroup.POST("/approve/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.approveSensorHandler)
		sensorsGroup.POST("/dismiss/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.dismissSensorHandler)
		sensorsGroup.GET("/by-id/:id/measurement-types", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.sensorMeasurementTypesHandler)
	}

	router.GET("/measurement-types", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.allMeasurementTypesHandler)
}
