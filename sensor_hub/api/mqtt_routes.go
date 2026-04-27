package api

import (
	"net/http"
	"strconv"

	gen "example/sensorHub/gen"
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterMQTTRoutes(router gin.IRouter) {
	brokers := router.Group("/mqtt/brokers")
	{
		brokers.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.ListMqttBrokers)
		brokers.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.CreateMqttBroker)
		brokers.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
				return
			}
			s.GetMqttBroker(c, id)
		})
		brokers.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
				return
			}
			s.UpdateMqttBroker(c, id)
		})
		brokers.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
				return
			}
			s.DeleteMqttBroker(c, id)
		})
	}

	subscriptions := router.Group("/mqtt/subscriptions")
	{
		subscriptions.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), func(c *gin.Context) {
			var params gen.ListMqttSubscriptionsParams
			if brokerParam := c.Query("broker_id"); brokerParam != "" {
				id, err := strconv.Atoi(brokerParam)
				if err != nil {
					c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker_id parameter"})
					return
				}
				params.BrokerId = &id
			}
			s.ListMqttSubscriptions(c, params)
		})
		subscriptions.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.CreateMqttSubscription)
		subscriptions.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
				return
			}
			s.GetMqttSubscription(c, id)
		})
		subscriptions.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
				return
			}
			s.UpdateMqttSubscription(c, id)
		})
		subscriptions.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
				return
			}
			s.DeleteMqttSubscription(c, id)
		})
	}

	router.GET("/mqtt/stats", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.GetMqttStats)
}
