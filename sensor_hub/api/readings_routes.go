package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterReadingsRoutes(router gin.IRouter) {
	readingsGroup := router.Group("/readings")
	{
		readingsGroup.GET("/between", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), getReadingsBetweenDatesHandler)
		readingsGroup.GET("/ws/current", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), currentReadingsWebSocket)
	}
}
