package ws

// NotificationBroadcaster implements the WebSocketNotifier interface for notification service
type NotificationBroadcaster struct{}

// NewNotificationBroadcaster creates a new notification broadcaster
func NewNotificationBroadcaster() *NotificationBroadcaster {
	return &NotificationBroadcaster{}
}

// BroadcastToUser sends a notification to a specific user via WebSocket
func (b *NotificationBroadcaster) BroadcastToUser(userID int, message interface{}) {
	BroadcastToUser(userID, message)
}
