package api

import (
	"example/sensorHub/api/middleware"
	gen "example/sensorHub/gen"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterAlertRoutes(router gin.IRouter) {
	alertsGroup := router.Group("/alerts")
	{
		alertsGroup.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), s.GetAllAlertRules)

		alertsGroup.GET("/sensor/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), func(c *gin.Context) {
			sensorId, err := strconv.Atoi(c.Param("sensorId"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
				return
			}
			s.GetAlertRulesBySensorId(c, sensorId)
		})

		alertsGroup.GET("/sensor/:sensorId/history", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), func(c *gin.Context) {
			sensorId, err := strconv.Atoi(c.Param("sensorId"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID"})
				return
			}
			var params gen.GetAlertHistoryParams
			if limitStr := c.Query("limit"); limitStr != "" {
				limit, err := strconv.Atoi(limitStr)
				if err != nil || limit < 1 || limit > 100 {
					limit = 50
				}
				params.Limit = &limit
			}
			s.GetAlertHistory(c, sensorId, params)
		})

		alertsGroup.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), s.CreateAlertRule)

		alertsGroup.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule ID"})
				return
			}
			s.GetAlertRuleById(c, id)
		})

		alertsGroup.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule ID"})
				return
			}
			s.UpdateAlertRule(c, id)
		})

		alertsGroup.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule ID"})
				return
			}
			s.DeleteAlertRule(c, id)
		})
	}
}

