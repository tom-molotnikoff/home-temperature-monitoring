package database

import (
	"database/sql"
	"fmt"
	"time"

	"example/sensorHub/notifications"
)

type UserEmailInfo struct {
	UserID int
	Email  string
}

type NotificationRepository interface {
	CreateNotification(notif notifications.Notification) (int, error)
	AssignNotificationToUser(userID, notificationID int) error
	AssignNotificationToUsersWithPermission(notificationID int, permission string) error
	GetUserIDsWithPermission(permission string) ([]int, error)
	GetUsersWithPermissionAndEmail(permission string) ([]UserEmailInfo, error)
	GetNotificationsForUser(userID int, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error)
	GetUnreadCountForUser(userID int) (int, error)
	MarkAsRead(userID, notificationID int) error
	DismissNotification(userID, notificationID int) error
	BulkMarkAsRead(userID int) error
	BulkDismiss(userID int) error
	DeleteOldNotifications(olderThan time.Time) (int64, error)
	GetChannelPreference(userID int, category notifications.NotificationCategory) (*notifications.ChannelPreference, error)
	GetAllChannelPreferences(userID int) ([]notifications.ChannelPreference, error)
	SetChannelPreference(pref notifications.ChannelPreference) error
	GetDefaultChannelPreference(category notifications.NotificationCategory) (*notifications.ChannelPreference, error)
}

type SqlNotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *SqlNotificationRepository {
	return &SqlNotificationRepository{db: db}
}

func (r *SqlNotificationRepository) CreateNotification(notif notifications.Notification) (int, error) {
	if err := notif.Validate(); err != nil {
		return 0, fmt.Errorf("invalid notification: %w", err)
	}

	metadataJSON, err := notif.MetadataJSON()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.Exec(
		"INSERT INTO notifications (category, severity, title, message, metadata) VALUES (?, ?, ?, ?, ?)",
		notif.Category, notif.Severity, notif.Title, notif.Message, metadataJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return int(id), nil
}

func (r *SqlNotificationRepository) AssignNotificationToUser(userID, notificationID int) error {
	_, err := r.db.Exec(
		"INSERT IGNORE INTO user_notifications (user_id, notification_id) VALUES (?, ?)",
		userID, notificationID,
	)
	return err
}

func (r *SqlNotificationRepository) AssignNotificationToUsersWithPermission(notificationID int, permission string) error {
	query := `
		INSERT IGNORE INTO user_notifications (user_id, notification_id)
		SELECT DISTINCT ur.user_id, ?
		FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE p.name = ?`
	_, err := r.db.Exec(query, notificationID, permission)
	return err
}

func (r *SqlNotificationRepository) GetUserIDsWithPermission(permission string) ([]int, error) {
	query := `
		SELECT DISTINCT ur.user_id
		FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE p.name = ?`
	rows, err := r.db.Query(query, permission)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}

func (r *SqlNotificationRepository) GetUsersWithPermissionAndEmail(permission string) ([]UserEmailInfo, error) {
	query := `
		SELECT DISTINCT ur.user_id, u.email
		FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		JOIN users u ON ur.user_id = u.id
		WHERE p.name = ? AND u.email IS NOT NULL AND u.email != '' AND u.disabled = FALSE`
	rows, err := r.db.Query(query, permission)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserEmailInfo
	for rows.Next() {
		var info UserEmailInfo
		if err := rows.Scan(&info.UserID, &info.Email); err != nil {
			return nil, err
		}
		users = append(users, info)
	}
	return users, rows.Err()
}

func (r *SqlNotificationRepository) GetNotificationsForUser(userID int, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error) {
	dismissedFilter := "AND un.is_dismissed = FALSE"
	if includeDismissed {
		dismissedFilter = ""
	}

	query := fmt.Sprintf(`
		SELECT un.id, un.user_id, un.notification_id, un.is_read, un.is_dismissed, un.read_at, un.dismissed_at,
		       n.id, n.category, n.severity, n.title, n.message, n.metadata, n.created_at
		FROM user_notifications un
		JOIN notifications n ON un.notification_id = n.id
		WHERE un.user_id = ? %s
		ORDER BY n.created_at DESC
		LIMIT ? OFFSET ?`, dismissedFilter)

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var results []notifications.UserNotification
	for rows.Next() {
		var un notifications.UserNotification
		var n notifications.Notification
		var metadataJSON []byte
		var readAt, dismissedAt sql.NullTime

		err := rows.Scan(
			&un.ID, &un.UserID, &un.NotificationID, &un.IsRead, &un.IsDismissed, &readAt, &dismissedAt,
			&n.ID, &n.Category, &n.Severity, &n.Title, &n.Message, &metadataJSON, &n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if readAt.Valid {
			un.ReadAt = &readAt.Time
		}
		if dismissedAt.Valid {
			un.DismissedAt = &dismissedAt.Time
		}
		n.Metadata, _ = notifications.ParseMetadataJSON(metadataJSON)
		un.Notification = &n
		results = append(results, un)
	}
	return results, nil
}

func (r *SqlNotificationRepository) GetUnreadCountForUser(userID int) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM user_notifications WHERE user_id = ? AND is_read = FALSE AND is_dismissed = FALSE",
		userID,
	).Scan(&count)
	return count, err
}

func (r *SqlNotificationRepository) MarkAsRead(userID, notificationID int) error {
	_, err := r.db.Exec(
		"UPDATE user_notifications SET is_read = TRUE, read_at = NOW() WHERE user_id = ? AND notification_id = ?",
		userID, notificationID,
	)
	return err
}

func (r *SqlNotificationRepository) DismissNotification(userID, notificationID int) error {
	_, err := r.db.Exec(
		"UPDATE user_notifications SET is_dismissed = TRUE, dismissed_at = NOW() WHERE user_id = ? AND notification_id = ?",
		userID, notificationID,
	)
	return err
}

func (r *SqlNotificationRepository) BulkMarkAsRead(userID int) error {
	_, err := r.db.Exec(
		"UPDATE user_notifications SET is_read = TRUE, read_at = NOW() WHERE user_id = ? AND is_read = FALSE",
		userID,
	)
	return err
}

func (r *SqlNotificationRepository) BulkDismiss(userID int) error {
	_, err := r.db.Exec(
		"UPDATE user_notifications SET is_dismissed = TRUE, dismissed_at = NOW() WHERE user_id = ? AND is_dismissed = FALSE",
		userID,
	)
	return err
}

func (r *SqlNotificationRepository) DeleteOldNotifications(olderThan time.Time) (int64, error) {
	result, err := r.db.Exec("DELETE FROM notifications WHERE created_at < ?", olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *SqlNotificationRepository) GetChannelPreference(userID int, category notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	var pref notifications.ChannelPreference
	err := r.db.QueryRow(
		"SELECT user_id, category, email_enabled, inapp_enabled FROM notification_channel_preferences WHERE user_id = ? AND category = ?",
		userID, category,
	).Scan(&pref.UserID, &pref.Category, &pref.EmailEnabled, &pref.InAppEnabled)

	if err == sql.ErrNoRows {
		return r.GetDefaultChannelPreference(category)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel preference: %w", err)
	}
	return &pref, nil
}

func (r *SqlNotificationRepository) GetDefaultChannelPreference(category notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	var pref notifications.ChannelPreference
	pref.Category = category
	err := r.db.QueryRow(
		"SELECT email_enabled, inapp_enabled FROM notification_channel_defaults WHERE category = ?",
		category,
	).Scan(&pref.EmailEnabled, &pref.InAppEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to get default preference: %w", err)
	}
	return &pref, nil
}

func (r *SqlNotificationRepository) GetAllChannelPreferences(userID int) ([]notifications.ChannelPreference, error) {
	categories := []notifications.NotificationCategory{
		notifications.CategoryThresholdAlert,
		notifications.CategoryUserManagement,
		notifications.CategoryConfigChange,
	}
	var prefs []notifications.ChannelPreference
	for _, cat := range categories {
		pref, err := r.GetChannelPreference(userID, cat)
		if err != nil {
			return nil, err
		}
		pref.UserID = userID
		prefs = append(prefs, *pref)
	}
	return prefs, nil
}

func (r *SqlNotificationRepository) SetChannelPreference(pref notifications.ChannelPreference) error {
	_, err := r.db.Exec(
		`INSERT INTO notification_channel_preferences (user_id, category, email_enabled, inapp_enabled)
		 VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE email_enabled = VALUES(email_enabled), inapp_enabled = VALUES(inapp_enabled)`,
		pref.UserID, pref.Category, pref.EmailEnabled, pref.InAppEnabled,
	)
	return err
}
