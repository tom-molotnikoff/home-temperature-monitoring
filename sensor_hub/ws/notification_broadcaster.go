package ws

import "log/slog"

// NotificationBroadcaster implements the WebSocketNotifier interface for notification service
type NotificationBroadcaster struct {
	logger *slog.Logger
}

// NewNotificationBroadcaster creates a new notification broadcaster
func NewNotificationBroadcaster(logger *slog.Logger) *NotificationBroadcaster {
	return &NotificationBroadcaster{logger: logger.With("component", "notification_broadcaster")}
}

// BroadcastToUser sends a notification to a specific user via WebSocket
func (b *NotificationBroadcaster) BroadcastToUser(userID int, message interface{}) {
	BroadcastToUser(userID, message)
}
