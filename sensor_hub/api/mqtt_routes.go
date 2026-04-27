package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterMQTTRoutes(router gin.IRouter) {
	brokers := router.Group("/mqtt/brokers")
	{
		brokers.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.listBrokersHandler)
		brokers.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.createBrokerHandler)
		brokers.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.getBrokerHandler)
		brokers.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.updateBrokerHandler)
		brokers.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.deleteBrokerHandler)
	}

	subscriptions := router.Group("/mqtt/subscriptions")
	{
		subscriptions.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.listSubscriptionsHandler)
		subscriptions.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.createSubscriptionHandler)
		subscriptions.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.getSubscriptionHandler)
		subscriptions.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.updateSubscriptionHandler)
		subscriptions.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), s.deleteSubscriptionHandler)
	}

	router.GET("/mqtt/stats", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), s.mqttStatsHandler)
}
