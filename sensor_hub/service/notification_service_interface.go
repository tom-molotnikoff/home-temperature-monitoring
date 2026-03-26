package service

import (
	"context"
	"example/sensorHub/notifications"
)

type NotificationServiceInterface interface {
	CreateNotification(ctx context.Context, notif notifications.Notification, targetPermission string) (int, error)
	GetNotificationsForUser(ctx context.Context, userID int, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error)
	GetUnreadCount(ctx context.Context, userID int) (int, error)
	MarkAsRead(ctx context.Context, userID, notificationID int) error
	Dismiss(ctx context.Context, userID, notificationID int) error
	BulkMarkAsRead(ctx context.Context, userID int) error
	BulkDismiss(ctx context.Context, userID int) error
	GetChannelPreferences(ctx context.Context, userID int) ([]notifications.ChannelPreference, error)
	SetChannelPreference(ctx context.Context, userID int, pref notifications.ChannelPreference) error
	ShouldNotifyChannel(ctx context.Context, userID int, category notifications.NotificationCategory, channel string) (bool, error)
}
