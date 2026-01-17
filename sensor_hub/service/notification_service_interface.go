package service

import "example/sensorHub/notifications"

type NotificationServiceInterface interface {
	CreateNotification(notif notifications.Notification, targetPermission string) (int, error)
	GetNotificationsForUser(userID int, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error)
	GetUnreadCount(userID int) (int, error)
	MarkAsRead(userID, notificationID int) error
	Dismiss(userID, notificationID int) error
	BulkMarkAsRead(userID int) error
	BulkDismiss(userID int) error
	GetChannelPreferences(userID int) ([]notifications.ChannelPreference, error)
	SetChannelPreference(userID int, pref notifications.ChannelPreference) error
	ShouldNotifyChannel(userID int, category notifications.NotificationCategory, channel string) (bool, error)
}
