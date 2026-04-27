package api

import (
	"example/sensorHub/api/middleware"
	gen "example/sensorHub/gen"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterReadingsRoutes(router gin.IRouter) {
	readingsGroup := router.Group("/readings")
	{
		readingsGroup.GET("/between", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), func(c *gin.Context) {
			var params gen.GetReadingsBetweenDatesParams
			params.Start = c.Query("start")
			params.End = c.Query("end")
			if sensor := c.Query("sensor"); sensor != "" {
				params.Sensor = &sensor
			}
			if t := c.Query("type"); t != "" {
				params.Type = &t
			}
			if agg := c.Query("aggregation"); agg != "" {
				v := gen.GetReadingsBetweenDatesParamsAggregation(agg)
				params.Aggregation = &v
			}
			if fn := c.Query("aggregation_function"); fn != "" {
				v := gen.GetReadingsBetweenDatesParamsAggregationFunction(fn)
				params.AggregationFunction = &v
			}
			s.GetReadingsBetweenDates(c, params)
		})
		readingsGroup.GET("/ws/current", middleware.AuthRequired(), middleware.RequirePermission("view_readings"), s.currentReadingsWebSocket)
	}
}
