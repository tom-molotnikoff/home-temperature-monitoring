package service

import (
	"context"
	"fmt"
	"log/slog"

	database "example/sensorHub/db"
	"example/sensorHub/notifications"
)

type WebSocketNotifier interface {
	BroadcastToUser(userID int, message interface{})
}

type EmailNotifier interface {
	SendNotification(recipient, title, message, category string) error
}

type NotificationService struct {
	repo    database.NotificationRepository
	ws      WebSocketNotifier
	emailer EmailNotifier
	logger  *slog.Logger
}

func NewNotificationService(repo database.NotificationRepository, ws WebSocketNotifier, logger *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, ws: ws, logger: logger.With("component", "notification_service")}
}

func (s *NotificationService) SetEmailNotifier(emailer EmailNotifier) {
	s.emailer = emailer
}

func (s *NotificationService) CreateNotification(ctx context.Context, notif notifications.Notification, targetPermission string) (int, error) {
	id, err := s.repo.CreateNotification(ctx, notif)
	if err != nil {
		return 0, fmt.Errorf("failed to create notification: %w", err)
	}

	if err := s.repo.AssignNotificationToUsersWithPermission(ctx, id, targetPermission); err != nil {
		return 0, fmt.Errorf("failed to assign notification: %w", err)
	}

	// Push via WebSocket to all assigned users
	if s.ws != nil {
		userIDs, err := s.repo.GetUserIDsWithPermission(ctx, targetPermission)
		if err == nil {
			notif.ID = id
			for _, userID := range userIDs {
				s.ws.BroadcastToUser(userID, notif)
			}
		}
	}

	// Send email notifications to users who have email enabled for this category
	if s.emailer != nil {
		go s.sendEmailNotifications(context.Background(), notif, targetPermission)
	}

	return id, nil
}

func (s *NotificationService) sendEmailNotifications(ctx context.Context, notif notifications.Notification, targetPermission string) {
	users, err := s.repo.GetUsersWithPermissionAndEmail(ctx, targetPermission)
	if err != nil {
		s.logger.Error("failed to get users for email notification", "error", err)
		return
	}

	s.logger.Debug("checking email preferences", "user_count", len(users), "category", notif.Category)

	for _, user := range users {
		pref, err := s.repo.GetChannelPreference(ctx, user.UserID, notif.Category)
		if err != nil {
			s.logger.Error("failed to get channel preference", "user_id", user.UserID, "error", err)
			continue
		}

		s.logger.Debug("user email preference", "user_id", user.UserID, "email", user.Email, "email_enabled", pref.EmailEnabled, "category", notif.Category)

		if !pref.EmailEnabled {
			continue
		}

		err = s.emailer.SendNotification(user.Email, notif.Title, notif.Message, string(notif.Category))
		if err != nil {
			s.logger.Error("failed to send email notification", "email", user.Email, "error", err)
		}
	}
}

func (s *NotificationService) GetNotificationsForUser(ctx context.Context, userID int, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error) {
	return s.repo.GetNotificationsForUser(ctx, userID, limit, offset, includeDismissed)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID int) (int, error) {
	return s.repo.GetUnreadCountForUser(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID int) error {
	return s.repo.MarkAsRead(ctx, userID, notificationID)
}

func (s *NotificationService) Dismiss(ctx context.Context, userID, notificationID int) error {
	return s.repo.DismissNotification(ctx, userID, notificationID)
}

func (s *NotificationService) BulkMarkAsRead(ctx context.Context, userID int) error {
	return s.repo.BulkMarkAsRead(ctx, userID)
}

func (s *NotificationService) BulkDismiss(ctx context.Context, userID int) error {
	return s.repo.BulkDismiss(ctx, userID)
}

func (s *NotificationService) GetChannelPreferences(ctx context.Context, userID int) ([]notifications.ChannelPreference, error) {
	return s.repo.GetAllChannelPreferences(ctx, userID)
}

func (s *NotificationService) SetChannelPreference(ctx context.Context, userID int, pref notifications.ChannelPreference) error {
	pref.UserID = userID
	return s.repo.SetChannelPreference(ctx, pref)
}

func (s *NotificationService) ShouldNotifyChannel(ctx context.Context, userID int, category notifications.NotificationCategory, channel string) (bool, error) {
	pref, err := s.repo.GetChannelPreference(ctx, userID, category)
	if err != nil {
		return false, err
	}
	switch channel {
	case "email":
		return pref.EmailEnabled, nil
	case "inapp":
		return pref.InAppEnabled, nil
	default:
		return false, fmt.Errorf("unknown channel: %s", channel)
	}
}
