package api

import (
	"example/sensorHub/notifications"
	gen "example/sensorHub/gen"
	"example/sensorHub/ws"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)



func getCurrentUserID(ctx *gin.Context) int {
	userObj, exists := ctx.Get("currentUser")
	if !exists {
		return 0
	}
	user, ok := userObj.(*gen.User)
	if !ok || user == nil {
		return 0
	}
	return user.Id
}

func (s *Server) listNotificationsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	includeDismissed := c.DefaultQuery("include_dismissed", "false") == "true"

	notifs, err := s.notificationService.GetNotificationsForUser(ctx, userID, limit, offset, includeDismissed)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get notifications", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, notifs)
}

func (s *Server) getUnreadCountHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	count, err := s.notificationService.GetUnreadCount(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get unread count", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"count": count})
}

func (s *Server) markAsReadHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	notifID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
		return
	}

	err = s.notificationService.MarkAsRead(ctx, userID, notifID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to mark as read", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func (s *Server) dismissNotificationHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	notifID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
		return
	}

	err = s.notificationService.Dismiss(ctx, userID, notifID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to dismiss notification", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "dismissed"})
}

func (s *Server) bulkMarkAsReadHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	err := s.notificationService.BulkMarkAsRead(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to mark all as read", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "all marked as read"})
}

func (s *Server) bulkDismissHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	err := s.notificationService.BulkDismiss(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to dismiss all", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "all dismissed"})
}

func (s *Server) getChannelPreferencesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	prefs, err := s.notificationService.GetChannelPreferences(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get preferences", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, prefs)
}

func (s *Server) setChannelPreferenceHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	var pref notifications.ChannelPreference
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body", "error": err.Error()})
		return
	}

	if pref.Category == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "category is required"})
		return
	}

	err := s.notificationService.SetChannelPreference(ctx, userID, pref)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to set preference", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "preference saved"})
}

func (s *Server) notificationsWebSocketHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	topic := ws.UserNotificationTopic(userID)
	createPushWebSocket(ctx, topic)
}
