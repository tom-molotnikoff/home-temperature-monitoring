package service

import (
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

func (m *mockNotificationRepo) CreateNotification(n notifications.Notification) (int, error) {
	m.createCalled = true
	m.lastNotifID = 42
	return 42, nil
}

func (m *mockNotificationRepo) AssignNotificationToUsersWithPermission(id int, perm string) error {
	m.assignCalled = true
	m.lastPermission = perm
	return nil
}

func (m *mockNotificationRepo) AssignNotificationToUser(userID, notifID int) error {
	return nil
}

func (m *mockNotificationRepo) GetNotificationsForUser(userID, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error) {
	return []notifications.UserNotification{}, nil
}

func (m *mockNotificationRepo) GetUnreadCountForUser(userID int) (int, error) {
	return m.unreadCount, nil
}

func (m *mockNotificationRepo) MarkAsRead(userID, notifID int) error {
	m.markAsReadCalled = true
	return nil
}

func (m *mockNotificationRepo) DismissNotification(userID, notifID int) error {
	m.dismissCalled = true
	return nil
}

func (m *mockNotificationRepo) BulkMarkAsRead(userID int) error {
	return nil
}

func (m *mockNotificationRepo) BulkDismiss(userID int) error {
	return nil
}

func (m *mockNotificationRepo) DeleteOldNotifications(olderThan time.Time) (int64, error) {
	return 0, nil
}

func (m *mockNotificationRepo) GetChannelPreference(userID int, cat notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	if m.channelPref != nil {
		return m.channelPref, nil
	}
	return &notifications.ChannelPreference{EmailEnabled: true, InAppEnabled: true}, nil
}

func (m *mockNotificationRepo) GetAllChannelPreferences(userID int) ([]notifications.ChannelPreference, error) {
	return []notifications.ChannelPreference{}, nil
}

func (m *mockNotificationRepo) SetChannelPreference(pref notifications.ChannelPreference) error {
	m.setPreferenceCalled = true
	return nil
}

func (m *mockNotificationRepo) GetDefaultChannelPreference(cat notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	return &notifications.ChannelPreference{EmailEnabled: true, InAppEnabled: true}, nil
}

func (m *mockNotificationRepo) GetUserIDsWithPermission(permission string) ([]int, error) {
	return []int{1, 2, 3}, nil
}

func (m *mockNotificationRepo) GetUsersWithPermissionAndEmail(permission string) ([]database.UserEmailInfo, error) {
	return []database.UserEmailInfo{}, nil
}

func TestNotificationService_CreateNotification(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	notif := notifications.Notification{
		Category: notifications.CategoryUserManagement,
		Severity: notifications.SeverityInfo,
		Title:    "User Added",
		Message:  "User john was added",
	}

	id, err := svc.CreateNotification(notif, "view_notifications_user_mgmt")
	require.NoError(t, err)
	require.Equal(t, 42, id)
	require.True(t, repo.createCalled)
	require.True(t, repo.assignCalled)
	require.Equal(t, "view_notifications_user_mgmt", repo.lastPermission)
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	repo := &mockNotificationRepo{unreadCount: 7}
	svc := NewNotificationService(repo, nil)

	count, err := svc.GetUnreadCount(1)
	require.NoError(t, err)
	require.Equal(t, 7, count)
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	err := svc.MarkAsRead(1, 5)
	require.NoError(t, err)
	require.True(t, repo.markAsReadCalled)
}

func TestNotificationService_Dismiss(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	err := svc.Dismiss(1, 5)
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
	svc := NewNotificationService(repo, nil)

	shouldNotify, err := svc.ShouldNotifyChannel(1, notifications.CategoryUserManagement, "email")
	require.NoError(t, err)
	require.True(t, shouldNotify)

	shouldNotify, err = svc.ShouldNotifyChannel(1, notifications.CategoryUserManagement, "inapp")
	require.NoError(t, err)
	require.False(t, shouldNotify)
}

func TestNotificationService_SetChannelPreference(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	pref := notifications.ChannelPreference{
		Category:     notifications.CategoryThresholdAlert,
		EmailEnabled: false,
		InAppEnabled: true,
	}

	err := svc.SetChannelPreference(1, pref)
	require.NoError(t, err)
	require.True(t, repo.setPreferenceCalled)
}
