package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterMQTTRoutes(router gin.IRouter) {
	brokers := router.Group("/mqtt/brokers")
	{
		brokers.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), listBrokersHandler)
		brokers.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), createBrokerHandler)
		brokers.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), getBrokerHandler)
		brokers.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), updateBrokerHandler)
		brokers.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), deleteBrokerHandler)
	}

	subscriptions := router.Group("/mqtt/subscriptions")
	{
		subscriptions.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), listSubscriptionsHandler)
		subscriptions.POST("", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), createSubscriptionHandler)
		subscriptions.GET("/:id", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), getSubscriptionHandler)
		subscriptions.PUT("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), updateSubscriptionHandler)
		subscriptions.DELETE("/:id", middleware.AuthRequired(), middleware.RequirePermission("manage_mqtt"), deleteSubscriptionHandler)
	}

	router.GET("/mqtt/stats", middleware.AuthRequired(), middleware.RequirePermission("view_mqtt"), mqttStatsHandler)
}
