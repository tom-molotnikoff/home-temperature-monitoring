package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(router *gin.Engine) {
	notifs := router.Group("/notifications")
	{
		notifs.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), listNotificationsHandler)
		notifs.GET("/unread-count", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), getUnreadCountHandler)
		notifs.POST("/:id/read", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), markAsReadHandler)
		notifs.POST("/:id/dismiss", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), dismissNotificationHandler)
		notifs.POST("/bulk/read", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), bulkMarkAsReadHandler)
		notifs.POST("/bulk/dismiss", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), bulkDismissHandler)
		notifs.GET("/preferences", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), getChannelPreferencesHandler)
		notifs.POST("/preferences", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), setChannelPreferenceHandler)
		notifs.GET("/ws", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), notificationsWebSocketHandler)
	}
}
