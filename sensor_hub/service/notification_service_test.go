package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	database "example/sensorHub/db"
	"example/sensorHub/notifications"

	"github.com/stretchr/testify/require"
)

type mockNotificationRepo struct {
	createCalled        bool
	assignCalled        bool
	markAsReadCalled    bool
	dismissCalled       bool
	setPreferenceCalled bool
	lastNotifID         int
	lastPermission      string
	unreadCount         int
	channelPref         *notifications.ChannelPreference
}

func (m *mockNotificationRepo) CreateNotification(ctx context.Context, n notifications.Notification) (int, error) {
	m.createCalled = true
	m.lastNotifID = 42
	return 42, nil
}

func (m *mockNotificationRepo) AssignNotificationToUsersWithPermission(ctx context.Context, id int, perm string) error {
	m.assignCalled = true
	m.lastPermission = perm
	return nil
}

func (m *mockNotificationRepo) AssignNotificationToUser(ctx context.Context, userID, notifID int) error {
	return nil
}

func (m *mockNotificationRepo) GetNotificationsForUser(ctx context.Context, userID, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error) {
	return []notifications.UserNotification{}, nil
}

func (m *mockNotificationRepo) GetUnreadCountForUser(ctx context.Context, userID int) (int, error) {
	return m.unreadCount, nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, userID, notifID int) error {
	m.markAsReadCalled = true
	return nil
}

func (m *mockNotificationRepo) DismissNotification(ctx context.Context, userID, notifID int) error {
	m.dismissCalled = true
	return nil
}

func (m *mockNotificationRepo) BulkMarkAsRead(ctx context.Context, userID int) error {
	return nil
}

func (m *mockNotificationRepo) BulkDismiss(ctx context.Context, userID int) error {
	return nil
}

func (m *mockNotificationRepo) DeleteOldNotifications(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}

func (m *mockNotificationRepo) GetChannelPreference(ctx context.Context, userID int, cat notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	if m.channelPref != nil {
		return m.channelPref, nil
	}
	return &notifications.ChannelPreference{EmailEnabled: true, InAppEnabled: true}, nil
}

func (m *mockNotificationRepo) GetAllChannelPreferences(ctx context.Context, userID int) ([]notifications.ChannelPreference, error) {
	return []notifications.ChannelPreference{}, nil
}

func (m *mockNotificationRepo) SetChannelPreference(ctx context.Context, pref notifications.ChannelPreference) error {
	m.setPreferenceCalled = true
	return nil
}

func (m *mockNotificationRepo) GetDefaultChannelPreference(ctx context.Context, cat notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	return &notifications.ChannelPreference{EmailEnabled: true, InAppEnabled: true}, nil
}

func (m *mockNotificationRepo) GetUserIDsWithPermission(ctx context.Context, permission string) ([]int, error) {
	return []int{1, 2, 3}, nil
}

func (m *mockNotificationRepo) GetUsersWithPermissionAndEmail(ctx context.Context, permission string) ([]database.UserEmailInfo, error) {
	return []database.UserEmailInfo{}, nil
}

func TestNotificationService_CreateNotification(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil, slog.Default())

	notif := notifications.Notification{
		Category: notifications.CategoryUserManagement,
		Severity: notifications.SeverityInfo,
		Title:    "User Added",
		Message:  "User john was added",
	}

	id, err := svc.CreateNotification(context.Background(), notif, "view_notifications_user_mgmt")
	require.NoError(t, err)
	require.Equal(t, 42, id)
	require.True(t, repo.createCalled)
	require.True(t, repo.assignCalled)
	require.Equal(t, "view_notifications_user_mgmt", repo.lastPermission)
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	repo := &mockNotificationRepo{unreadCount: 7}
	svc := NewNotificationService(repo, nil, slog.Default())

	count, err := svc.GetUnreadCount(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 7, count)
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil, slog.Default())

	err := svc.MarkAsRead(context.Background(), 1, 5)
	require.NoError(t, err)
	require.True(t, repo.markAsReadCalled)
}

func TestNotificationService_Dismiss(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil, slog.Default())

	err := svc.Dismiss(context.Background(), 1, 5)
	require.NoError(t, err)
	require.True(t, repo.dismissCalled)
}

func TestNotificationService_ShouldNotifyChannel_Email(t *testing.T) {
	repo := &mockNotificationRepo{
		channelPref: &notifications.ChannelPreference{
			EmailEnabled: true,
			InAppEnabled: false,
		},
	}
	svc := NewNotificationService(repo, nil, slog.Default())

	shouldNotify, err := svc.ShouldNotifyChannel(context.Background(), 1, notifications.CategoryUserManagement, "email")
	require.NoError(t, err)
	require.True(t, shouldNotify)

	shouldNotify, err = svc.ShouldNotifyChannel(context.Background(), 1, notifications.CategoryUserManagement, "inapp")
	require.NoError(t, err)
	require.False(t, shouldNotify)
}

func TestNotificationService_SetChannelPreference(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil, slog.Default())

	pref := notifications.ChannelPreference{
		Category:     notifications.CategoryThresholdAlert,
		EmailEnabled: false,
		InAppEnabled: true,
	}

	err := svc.SetChannelPreference(context.Background(), 1, pref)
	require.NoError(t, err)
	require.True(t, repo.setPreferenceCalled)
}
