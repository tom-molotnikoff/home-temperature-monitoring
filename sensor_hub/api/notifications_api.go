package api

import (
	"example/sensorHub/notifications"
	"example/sensorHub/service"
	"example/sensorHub/types"
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
	user, ok := userObj.(*types.User)
	if !ok || user == nil {
		return 0
	}
	return user.Id
}

func listNotificationsHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	includeDismissed := ctx.DefaultQuery("include_dismissed", "false") == "true"

	notifs, err := notificationService.GetNotificationsForUser(userID, limit, offset, includeDismissed)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get notifications", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, notifs)
}

func getUnreadCountHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	count, err := notificationService.GetUnreadCount(userID)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get unread count", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"count": count})
}

func markAsReadHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	notifID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
		return
	}

	err = notificationService.MarkAsRead(userID, notifID)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to mark as read", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func dismissNotificationHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	notifID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
		return
	}

	err = notificationService.Dismiss(userID, notifID)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to dismiss notification", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "dismissed"})
}

func bulkMarkAsReadHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	err := notificationService.BulkMarkAsRead(userID)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to mark all as read", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "all marked as read"})
}

func bulkDismissHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	err := notificationService.BulkDismiss(userID)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to dismiss all", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "all dismissed"})
}

func getChannelPreferencesHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	prefs, err := notificationService.GetChannelPreferences(userID)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to get preferences", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, prefs)
}

func setChannelPreferenceHandler(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)
	if userID == 0 {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	var pref notifications.ChannelPreference
	if err := ctx.ShouldBindJSON(&pref); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body", "error": err.Error()})
		return
	}

	if pref.Category == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "category is required"})
		return
	}

	err := notificationService.SetChannelPreference(userID, pref)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to set preference", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "preference saved"})
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
