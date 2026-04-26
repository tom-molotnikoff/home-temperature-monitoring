package api

import (
	"example/sensorHub/notifications"
	gen "example/sensorHub/gen"
	"example/sensorHub/service"
	"example/sensorHub/ws"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var notificationService service.NotificationServiceInterface

func InitNotificationsAPI(ns service.NotificationServiceInterface) {
	notificationService = ns
}

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

func listNotificationsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	includeDismissed := c.DefaultQuery("include_dismissed", "false") == "true"

	notifs, err := notificationService.GetNotificationsForUser(ctx, userID, limit, offset, includeDismissed)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get notifications", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, notifs)
}

func getUnreadCountHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	count, err := notificationService.GetUnreadCount(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get unread count", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"count": count})
}

func markAsReadHandler(c *gin.Context) {
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

	err = notificationService.MarkAsRead(ctx, userID, notifID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to mark as read", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func dismissNotificationHandler(c *gin.Context) {
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

	err = notificationService.Dismiss(ctx, userID, notifID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to dismiss notification", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "dismissed"})
}

func bulkMarkAsReadHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	err := notificationService.BulkMarkAsRead(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to mark all as read", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "all marked as read"})
}

func bulkDismissHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	err := notificationService.BulkDismiss(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to dismiss all", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "all dismissed"})
}

func getChannelPreferencesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getCurrentUserID(c)
	if userID == 0 {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	prefs, err := notificationService.GetChannelPreferences(ctx, userID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get preferences", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, prefs)
}

func setChannelPreferenceHandler(c *gin.Context) {
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

	err := notificationService.SetChannelPreference(ctx, userID, pref)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to set preference", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "preference saved"})
}

func notificationsWebSocketHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	topic := ws.UserNotificationTopic(userID)
	createPushWebSocket(ctx, topic)
}
