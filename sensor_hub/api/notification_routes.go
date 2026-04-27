package api

import (
	"example/sensorHub/api/middleware"
	gen "example/sensorHub/gen"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterNotificationRoutes(router gin.IRouter) {
	notifs := router.Group("/notifications")
	{
		notifs.GET("", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), func(c *gin.Context) {
			var params gen.ListNotificationsParams
			if lStr := c.Query("limit"); lStr != "" {
				if v, err := strconv.Atoi(lStr); err == nil {
					params.Limit = &v
				}
			}
			if oStr := c.Query("offset"); oStr != "" {
				if v, err := strconv.Atoi(oStr); err == nil {
					params.Offset = &v
				}
			}
			if dStr := c.Query("include_dismissed"); dStr != "" {
				v := gen.ListNotificationsParamsIncludeDismissed(dStr)
				params.IncludeDismissed = &v
			}
			s.ListNotifications(c, params)
		})
		notifs.GET("/unread-count", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.GetUnreadCount)
		notifs.POST("/:id/read", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
				return
			}
			s.MarkAsRead(c, id)
		})
		notifs.POST("/:id/dismiss", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
				return
			}
			s.DismissNotification(c, id)
		})
		notifs.POST("/bulk/read", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.BulkMarkAsRead)
		notifs.POST("/bulk/dismiss", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), s.BulkDismiss)
		notifs.GET("/preferences", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.GetChannelPreferences)
		notifs.POST("/preferences", middleware.AuthRequired(), middleware.RequirePermission("manage_notifications"), s.SetChannelPreference)
		notifs.GET("/ws", middleware.AuthRequired(), middleware.RequirePermission("view_notifications"), s.notificationsWebSocketHandler)
	}
}
