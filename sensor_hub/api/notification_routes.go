package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterNotificationRoutes(router gin.IRouter) {
	notifs := router.Group("/notifications")
	{
		notifs.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.listNotificationsHandler)
		notifs.GET("/unread-count", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.getUnreadCountHandler)
		notifs.POST("/:id/read", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.markAsReadHandler)
		notifs.POST("/:id/dismiss", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), s.dismissNotificationHandler)
		notifs.POST("/bulk/read", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.bulkMarkAsReadHandler)
		notifs.POST("/bulk/dismiss", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), s.bulkDismissHandler)
		notifs.GET("/preferences", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.getChannelPreferencesHandler)
		notifs.POST("/preferences", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), s.setChannelPreferenceHandler)
		notifs.GET("/ws", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.notificationsWebSocketHandler)
	}
}
