package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterTemperatureRoutes(router *gin.Engine) {
	temperatureGroup := router.Group("/temperature")
	{
		temperatureGroup.GET("/readings/between", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getReadingsBetweenDatesHandler)
		temperatureGroup.GET("/readings/hourly/between", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getHourlyReadingsBetweenDatesHandler)
		temperatureGroup.GET("/ws/current-temperatures", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), currentTemperaturesWebSocket)
	}
}
