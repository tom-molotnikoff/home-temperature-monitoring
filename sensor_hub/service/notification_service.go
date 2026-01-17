package service

import (
	"fmt"
	"log"

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
}

func NewNotificationService(repo database.NotificationRepository, ws WebSocketNotifier) *NotificationService {
	return &NotificationService{repo: repo, ws: ws}
}

func (s *NotificationService) SetEmailNotifier(emailer EmailNotifier) {
	s.emailer = emailer
}

func (s *NotificationService) CreateNotification(notif notifications.Notification, targetPermission string) (int, error) {
	id, err := s.repo.CreateNotification(notif)
	if err != nil {
		return 0, fmt.Errorf("failed to create notification: %w", err)
	}

	if err := s.repo.AssignNotificationToUsersWithPermission(id, targetPermission); err != nil {
		return 0, fmt.Errorf("failed to assign notification: %w", err)
	}

	// Push via WebSocket to all assigned users
	if s.ws != nil {
		userIDs, err := s.repo.GetUserIDsWithPermission(targetPermission)
		if err == nil {
			notif.ID = id
			for _, userID := range userIDs {
				s.ws.BroadcastToUser(userID, notif)
			}
		}
	}

	// Send email notifications to users who have email enabled for this category
	if s.emailer != nil {
		go s.sendEmailNotifications(notif, targetPermission)
	}

	return id, nil
}

func (s *NotificationService) sendEmailNotifications(notif notifications.Notification, targetPermission string) {
	users, err := s.repo.GetUsersWithPermissionAndEmail(targetPermission)
	if err != nil {
		log.Printf("Failed to get users for email notification: %v", err)
		return
	}

	log.Printf("Checking email preferences for %d users, category=%s", len(users), notif.Category)

	for _, user := range users {
		pref, err := s.repo.GetChannelPreference(user.UserID, notif.Category)
		if err != nil {
			log.Printf("Failed to get channel preference for user %d: %v", user.UserID, err)
			continue
		}

		log.Printf("User %d (%s): email_enabled=%v for category %s", user.UserID, user.Email, pref.EmailEnabled, notif.Category)

		if !pref.EmailEnabled {
			continue
		}

		err = s.emailer.SendNotification(user.Email, notif.Title, notif.Message, string(notif.Category))
		if err != nil {
			log.Printf("Failed to send email to %s: %v", user.Email, err)
		}
	}
}

func (s *NotificationService) GetNotificationsForUser(userID int, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error) {
	return s.repo.GetNotificationsForUser(userID, limit, offset, includeDismissed)
}

func (s *NotificationService) GetUnreadCount(userID int) (int, error) {
	return s.repo.GetUnreadCountForUser(userID)
}

func (s *NotificationService) MarkAsRead(userID, notificationID int) error {
	return s.repo.MarkAsRead(userID, notificationID)
}

func (s *NotificationService) Dismiss(userID, notificationID int) error {
	return s.repo.DismissNotification(userID, notificationID)
}

func (s *NotificationService) BulkMarkAsRead(userID int) error {
	return s.repo.BulkMarkAsRead(userID)
}

func (s *NotificationService) BulkDismiss(userID int) error {
	return s.repo.BulkDismiss(userID)
}

func (s *NotificationService) GetChannelPreferences(userID int) ([]notifications.ChannelPreference, error) {
	return s.repo.GetAllChannelPreferences(userID)
}

func (s *NotificationService) SetChannelPreference(userID int, pref notifications.ChannelPreference) error {
	pref.UserID = userID
	return s.repo.SetChannelPreference(pref)
}

func (s *NotificationService) ShouldNotifyChannel(userID int, category notifications.NotificationCategory, channel string) (bool, error) {
	pref, err := s.repo.GetChannelPreference(userID, category)
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
