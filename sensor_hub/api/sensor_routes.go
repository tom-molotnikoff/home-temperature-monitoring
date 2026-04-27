package api

import (
	"fmt"
	"net/http"
	"strconv"

	"example/sensorHub/api/middleware"
	gen "example/sensorHub/gen"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterSensorRoutes(router gin.IRouter) {
	sensorsGroup := router.Group("/sensors")
	{
		sensorsGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), s.AddSensor)
		sensorsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), func(c *gin.Context) {
			var id int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
				return
			}
			s.UpdateSensorById(c, id)
		})
		sensorsGroup.DELETE("/:name", middleware.AuthRequired(), middleware.RequirePermission("delete_sensors"), func(c *gin.Context) {
			s.DeleteSensorByName(c, c.Param("name"))
		})
		sensorsGroup.GET("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
			s.GetSensorByName(c, c.Param("name"))
		})
		sensorsGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.GetAllSensors)
		sensorsGroup.GET("/driver/:driver", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
			s.GetSensorsByDriver(c, c.Param("driver"))
		})
		sensorsGroup.HEAD("/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
			s.SensorExists(c, c.Param("name"))
		})
		sensorsGroup.POST("/collect", middleware.AuthRequired(), middleware.RequirePermission("trigger_readings"), s.CollectAllSensorReadings)
		sensorsGroup.POST("/collect/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("trigger_readings"), func(c *gin.Context) {
			s.CollectFromSensor(c, c.Param("sensorName"))
		})
		sensorsGroup.POST("/disable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), func(c *gin.Context) {
			s.DisableSensor(c, c.Param("sensorName"))
		})
		sensorsGroup.POST("/enable/:sensorName", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), func(c *gin.Context) {
			s.EnableSensor(c, c.Param("sensorName"))
		})
		sensorsGroup.GET("/ws", middleware.AuthRequired(), s.SubscribeAllSensors)
		sensorsGroup.GET("/ws/:driver", middleware.AuthRequired(), func(c *gin.Context) {
			s.SubscribeSensorsByDriver(c, c.Param("driver"))
		})
		sensorsGroup.GET("/health/:name", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
			var params gen.GetSensorHealthHistoryByNameParams
			if limitStr := c.Query("limit"); limitStr != "" {
				limit, err := strconv.Atoi(limitStr)
				if err != nil || limit <= 0 {
					c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
					return
				}
				params.Limit = &limit
			}
			s.GetSensorHealthHistoryByName(c, c.Param("name"), params)
		})
		sensorsGroup.GET("/stats/total-readings", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), s.GetTotalReadingsPerSensor)
		sensorsGroup.GET("/status/:status", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
			s.GetSensorsByStatus(c, gen.GetSensorsByStatusParamsStatus(c.Param("status")))
		})
		sensorsGroup.POST("/approve/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), func(c *gin.Context) {
			var id int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
				return
			}
			s.ApproveSensor(c, id)
		})
		sensorsGroup.POST("/dismiss/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_sensors"), func(c *gin.Context) {
			var id int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
				return
			}
			s.DismissSensor(c, id)
		})
		sensorsGroup.GET("/by-id/:id/measurement-types", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
			var id int
			if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
				return
			}
			s.GetSensorMeasurementTypes(c, id)
		})
	}

	router.GET("/measurement-types", middleware.AuthRequired(), middleware.RequirePermission("view_sensors"), func(c *gin.Context) {
		var params gen.GetAllMeasurementTypesParams
		if hr := c.Query("has_readings"); hr != "" {
			hasReadings := hr == "true"
			params.HasReadings = &hasReadings
		}
		s.GetAllMeasurementTypes(c, params)
	})
}

