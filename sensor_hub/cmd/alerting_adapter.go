package cmd

import (
	"context"

	"example/sensorHub/alerting"
	database "example/sensorHub/db"
	"example/sensorHub/notifications"
)

// notifRepoAdapter bridges database.NotificationRepository to alerting.NotificationRepository.
// The only difference is that GetUsersWithPermissionAndEmail returns []alerting.UserEmailInfo
// rather than []database.UserEmailInfo.
type notifRepoAdapter struct {
	repo database.NotificationRepository
}

func (a *notifRepoAdapter) CreateNotification(ctx context.Context, notif notifications.Notification) (int, error) {
	return a.repo.CreateNotification(ctx, notif)
}

func (a *notifRepoAdapter) AssignNotificationToUsersWithPermission(ctx context.Context, notifID int, permission string) error {
	return a.repo.AssignNotificationToUsersWithPermission(ctx, notifID, permission)
}

func (a *notifRepoAdapter) GetUserIDsWithPermission(ctx context.Context, permission string) ([]int, error) {
	return a.repo.GetUserIDsWithPermission(ctx, permission)
}

func (a *notifRepoAdapter) GetUsersWithPermissionAndEmail(ctx context.Context, permission string) ([]alerting.UserEmailInfo, error) {
	users, err := a.repo.GetUsersWithPermissionAndEmail(ctx, permission)
	if err != nil {
		return nil, err
	}
	result := make([]alerting.UserEmailInfo, len(users))
	for i, u := range users {
		result[i] = alerting.UserEmailInfo{UserID: u.UserID, Email: u.Email}
	}
	return result, nil
}

func (a *notifRepoAdapter) GetChannelPreference(ctx context.Context, userID int, category notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	return a.repo.GetChannelPreference(ctx, userID, category)
}
