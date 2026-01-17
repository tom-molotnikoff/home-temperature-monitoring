# In-App Notifications Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Implement a full-featured in-app notification system with bell icon, dismissable notifications, category-based filtering, RBAC integration, per-user channel preferences, and real-time WebSocket delivery.

**Architecture:** Notifications stored in MySQL with category (threshold_alert, user_management, config_change), severity (info, warning, error), and per-user dismissal tracking. NotificationService creates notifications and distributes to users based on permissions. Existing AlertService extended to create in-app notifications alongside emails, sharing rate limits. WebSocket pushes real-time updates to connected clients. Users configure channel preferences (email vs in-app) per category with admin-defined global defaults. Frontend displays bell icon with unread count badge, dropdown preview of last 5 notifications, and dedicated page for history/bulk management.

**Tech Stack:** Go/Gin (backend), MySQL (persistence), React/TypeScript/MUI (frontend), WebSocket (real-time)

---

## Task Outline

### Phase 1: Database Schema
- **Task 1:** Notifications tables (notifications, user_notifications)
- **Task 2:** Permissions and channel preference tables

### Phase 2: Backend - Core Types & Repository
- **Task 3:** Notification types and models
- **Task 4:** Notification repository (CRUD + user assignment)
- **Task 5:** Channel preference repository methods

### Phase 3: Backend - Service Layer
- **Task 6:** Notification service (create, distribute, dismiss, preferences)
- **Task 7:** Integrate with user management events
- **Task 8:** Integrate with sensor configuration events
- **Task 9:** Integrate with AlertService for threshold alerts

### Phase 4: Backend - API & WebSocket
- **Task 10:** Notification API endpoints
- **Task 11:** Channel preference API endpoints
- **Task 12:** WebSocket push integration

### Phase 5: Backend - Maintenance
- **Task 13:** Cleanup service integration (90-day auto-purge)

### Phase 6: Frontend - Core
- **Task 14:** Notification types and API client
- **Task 15:** Notification context provider with WebSocket subscription
- **Task 16:** Bell icon component with badge

### Phase 7: Frontend - UI Components
- **Task 17:** Notification dropdown (last 5 + quick actions)
- **Task 18:** Notifications page (history + bulk management)
- **Task 19:** Channel preferences settings UI
- **Task 20:** Navigation integration

### Phase 8: Testing
- **Task 21:** Repository unit tests
- **Task 22:** Service unit tests
- **Task 23:** API integration tests

### Phase 9: Documentation
- **Task 24:** Create notification-system.md documentation

---

## Detailed Task Steps

### Task 1: Notifications Tables

**Files:**
- Create: `sensor_hub/db/changesets/V17__add_notifications.sql`

**Step 1: Create migration file**

```sql
-- V17: In-app notifications system

CREATE TABLE notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_notifications_category (category),
    INDEX idx_notifications_created_at (created_at)
);

CREATE TABLE user_notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    notification_id INT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    is_dismissed BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP NULL,
    dismissed_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_notification (user_id, notification_id),
    INDEX idx_user_notifications_unread (user_id, is_read, is_dismissed)
);
```

**Step 2: Verify migration syntax**

Run: `cd sensor_hub && cat db/changesets/V17__add_notifications.sql`
Expected: File contents displayed without syntax errors

**Step 3: Commit**

```bash
git add sensor_hub/db/changesets/V17__add_notifications.sql
git commit -m "feat(db): add notifications and user_notifications tables"
```

---

### Task 2: Permissions and Channel Preference Tables

**Files:**
- Create: `sensor_hub/db/changesets/V18__add_notification_permissions.sql`

**Step 1: Create migration file**

```sql
-- V18: Notification permissions and channel preferences

INSERT IGNORE INTO permissions (name, description) VALUES 
('view_notifications', 'Access and view in-app notifications'),
('view_notifications_user_mgmt', 'Receive user management notifications'),
('view_notifications_config', 'Receive configuration change notifications'),
('manage_notifications', 'Configure notification channel preferences');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.name = 'admin' AND p.name IN (
    'view_notifications', 
    'view_notifications_user_mgmt', 
    'view_notifications_config', 
    'manage_notifications'
);

CREATE TABLE notification_channel_defaults (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category VARCHAR(50) NOT NULL UNIQUE,
    email_enabled BOOLEAN DEFAULT TRUE,
    inapp_enabled BOOLEAN DEFAULT TRUE
);

INSERT INTO notification_channel_defaults (category, email_enabled, inapp_enabled) VALUES
('threshold_alert', TRUE, TRUE),
('user_management', FALSE, TRUE),
('config_change', FALSE, TRUE);

CREATE TABLE notification_channel_preferences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    category VARCHAR(50) NOT NULL,
    email_enabled BOOLEAN NOT NULL,
    inapp_enabled BOOLEAN NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_category_pref (user_id, category)
);
```

**Step 2: Verify migration syntax**

Run: `cd sensor_hub && cat db/changesets/V18__add_notification_permissions.sql`
Expected: File contents displayed without syntax errors

**Step 3: Commit**

```bash
git add sensor_hub/db/changesets/V18__add_notification_permissions.sql
git commit -m "feat(db): add notification permissions and channel preferences"
```

---

### Task 3: Notification Types and Models

**Files:**
- Create: `sensor_hub/notifications/types.go`
- Create: `sensor_hub/notifications/types_test.go`

**Step 1: Write failing test**

Create `sensor_hub/notifications/types_test.go`:

```go
package notifications

import "testing"

func TestNotification_Validate_Valid(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "User added",
		Message:  "User john was added to the system",
	}
	if err := n.Validate(); err != nil {
		t.Errorf("expected valid notification, got error: %v", err)
	}
}

func TestNotification_Validate_InvalidCategory(t *testing.T) {
	n := Notification{
		Category: "invalid_category",
		Severity: SeverityInfo,
		Title:    "Test",
		Message:  "Test message",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for invalid category")
	}
}

func TestNotification_Validate_EmptyTitle(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: SeverityInfo,
		Title:    "",
		Message:  "Test message",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for empty title")
	}
}

func TestNotification_Validate_InvalidSeverity(t *testing.T) {
	n := Notification{
		Category: CategoryUserManagement,
		Severity: "critical",
		Title:    "Test",
		Message:  "Test message",
	}
	if err := n.Validate(); err == nil {
		t.Error("expected error for invalid severity")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd sensor_hub && go test ./notifications/... -v`
Expected: FAIL - package notifications does not exist

**Step 3: Implement types**

Create `sensor_hub/notifications/types.go`:

```go
package notifications

import (
	"encoding/json"
	"fmt"
	"time"
)

type NotificationCategory string

const (
	CategoryThresholdAlert NotificationCategory = "threshold_alert"
	CategoryUserManagement NotificationCategory = "user_management"
	CategoryConfigChange   NotificationCategory = "config_change"
)

type NotificationSeverity string

const (
	SeverityInfo    NotificationSeverity = "info"
	SeverityWarning NotificationSeverity = "warning"
	SeverityError   NotificationSeverity = "error"
)

type Notification struct {
	ID        int                    `json:"id"`
	Category  NotificationCategory   `json:"category"`
	Severity  NotificationSeverity   `json:"severity"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type UserNotification struct {
	ID             int           `json:"id"`
	UserID         int           `json:"user_id"`
	NotificationID int           `json:"notification_id"`
	IsRead         bool          `json:"is_read"`
	IsDismissed    bool          `json:"is_dismissed"`
	ReadAt         *time.Time    `json:"read_at,omitempty"`
	DismissedAt    *time.Time    `json:"dismissed_at,omitempty"`
	Notification   *Notification `json:"notification,omitempty"`
}

type ChannelPreference struct {
	UserID       int                  `json:"user_id,omitempty"`
	Category     NotificationCategory `json:"category"`
	EmailEnabled bool                 `json:"email_enabled"`
	InAppEnabled bool                 `json:"inapp_enabled"`
}

var validCategories = map[NotificationCategory]bool{
	CategoryThresholdAlert: true,
	CategoryUserManagement: true,
	CategoryConfigChange:   true,
}

var validSeverities = map[NotificationSeverity]bool{
	SeverityInfo:    true,
	SeverityWarning: true,
	SeverityError:   true,
}

func (n *Notification) Validate() error {
	if !validCategories[n.Category] {
		return fmt.Errorf("invalid category: %s", n.Category)
	}
	if !validSeverities[n.Severity] {
		return fmt.Errorf("invalid severity: %s", n.Severity)
	}
	if n.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if n.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	return nil
}

func (n *Notification) MetadataJSON() ([]byte, error) {
	if n.Metadata == nil {
		return []byte("null"), nil
	}
	return json.Marshal(n.Metadata)
}

func ParseMetadataJSON(data []byte) (map[string]interface{}, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd sensor_hub && go test ./notifications/... -v`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add sensor_hub/notifications/
git commit -m "feat(notifications): add notification types and validation"
```

<!-- Tasks 4-6 details below -->

---

### Task 4: Notification Repository

**Files:**
- Create: `sensor_hub/db/notification_repository.go`
- Create: `sensor_hub/db/notification_repository_test.go`

**Step 1: Write failing test**

Create `sensor_hub/db/notification_repository_test.go`:

```go
package database

import (
	"testing"
	"time"

	"example/sensorHub/notifications"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNotificationRepository_CreateNotification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectExec("INSERT INTO notifications").
		WithArgs("user_management", "info", "Test Title", "Test Message", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	notif := notifications.Notification{
		Category: notifications.CategoryUserManagement,
		Severity: notifications.SeverityInfo,
		Title:    "Test Title",
		Message:  "Test Message",
	}

	id, err := repo.CreateNotification(notif)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}

func TestNotificationRepository_GetUnreadCountForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.GetUnreadCountForUser(1)
	if err != nil {
		t.Fatalf("GetUnreadCountForUser failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd sensor_hub && go test ./db/... -run TestNotificationRepository -v`
Expected: FAIL - undefined: NewNotificationRepository

**Step 3: Implement repository**

Create `sensor_hub/db/notification_repository.go`:

```go
package database

import (
	"database/sql"
	"fmt"
	"time"

	"example/sensorHub/notifications"
)

type NotificationRepository interface {
	CreateNotification(notif notifications.Notification) (int, error)
	AssignNotificationToUser(userID, notificationID int) error
	AssignNotificationToUsersWithPermission(notificationID int, permission string) error
	GetNotificationsForUser(userID int, limit, offset int, includeDissmissed bool) ([]notifications.UserNotification, error)
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
```

**Step 4: Run tests to verify they pass**

Run: `cd sensor_hub && go test ./db/... -run TestNotificationRepository -v`
Expected: PASS

**Step 5: Commit**

```bash
git add sensor_hub/db/notification_repository.go sensor_hub/db/notification_repository_test.go
git commit -m "feat(db): add notification repository with CRUD operations"
```

---

### Task 5: Channel Preference Repository Methods

**Files:**
- Modify: `sensor_hub/db/notification_repository.go`
- Modify: `sensor_hub/db/notification_repository_test.go`

**Step 1: Add test for channel preferences**

Append to `sensor_hub/db/notification_repository_test.go`:

```go
func TestNotificationRepository_GetChannelPreference_UserOverride(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT .* FROM notification_channel_preferences").
		WithArgs(1, "user_management").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "category", "email_enabled", "inapp_enabled"}).
			AddRow(1, "user_management", false, true))

	pref, err := repo.GetChannelPreference(1, notifications.CategoryUserManagement)
	if err != nil {
		t.Fatalf("GetChannelPreference failed: %v", err)
	}
	if pref.EmailEnabled != false || pref.InAppEnabled != true {
		t.Errorf("unexpected preference values: %+v", pref)
	}
}

func TestNotificationRepository_GetChannelPreference_FallbackToDefault(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT .* FROM notification_channel_preferences").
		WithArgs(1, "threshold_alert").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("SELECT .* FROM notification_channel_defaults").
		WithArgs("threshold_alert").
		WillReturnRows(sqlmock.NewRows([]string{"email_enabled", "inapp_enabled"}).
			AddRow(true, true))

	pref, err := repo.GetChannelPreference(1, notifications.CategoryThresholdAlert)
	if err != nil {
		t.Fatalf("GetChannelPreference failed: %v", err)
	}
	if pref.EmailEnabled != true || pref.InAppEnabled != true {
		t.Errorf("unexpected preference values: %+v", pref)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd sensor_hub && go test ./db/... -run TestNotificationRepository_GetChannelPreference -v`
Expected: FAIL - undefined method

**Step 3: Implement preference methods**

Append to `sensor_hub/db/notification_repository.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

Run: `cd sensor_hub && go test ./db/... -run TestNotificationRepository -v`
Expected: PASS

**Step 5: Commit**

```bash
git add sensor_hub/db/notification_repository.go sensor_hub/db/notification_repository_test.go
git commit -m "feat(db): add channel preference repository methods"
```

---

### Task 6: Notification Service

**Files:**
- Create: `sensor_hub/service/notification_service.go`
- Create: `sensor_hub/service/notification_service_interface.go`
- Create: `sensor_hub/service/notification_service_test.go`

**Step 1: Create interface**

Create `sensor_hub/service/notification_service_interface.go`:

```go
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
```

**Step 2: Write failing test**

Create `sensor_hub/service/notification_service_test.go`:

```go
package service

import (
	"testing"

	"example/sensorHub/notifications"
)

type mockNotificationRepo struct {
	createCalled    bool
	assignCalled    bool
	lastNotifID     int
	lastPermission  string
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

// Implement remaining interface methods as stubs...
func (m *mockNotificationRepo) AssignNotificationToUser(int, int) error { return nil }
func (m *mockNotificationRepo) GetNotificationsForUser(int, int, int, bool) ([]notifications.UserNotification, error) { return nil, nil }
func (m *mockNotificationRepo) GetUnreadCountForUser(int) (int, error) { return 0, nil }
func (m *mockNotificationRepo) MarkAsRead(int, int) error { return nil }
func (m *mockNotificationRepo) DismissNotification(int, int) error { return nil }
func (m *mockNotificationRepo) BulkMarkAsRead(int) error { return nil }
func (m *mockNotificationRepo) BulkDismiss(int) error { return nil }
func (m *mockNotificationRepo) DeleteOldNotifications(t time.Time) (int64, error) { return 0, nil }
func (m *mockNotificationRepo) GetChannelPreference(int, notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	return &notifications.ChannelPreference{EmailEnabled: true, InAppEnabled: true}, nil
}
func (m *mockNotificationRepo) GetAllChannelPreferences(int) ([]notifications.ChannelPreference, error) { return nil, nil }
func (m *mockNotificationRepo) SetChannelPreference(notifications.ChannelPreference) error { return nil }
func (m *mockNotificationRepo) GetDefaultChannelPreference(notifications.NotificationCategory) (*notifications.ChannelPreference, error) { return nil, nil }

func TestNotificationService_CreateNotification(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo)

	notif := notifications.Notification{
		Category: notifications.CategoryUserManagement,
		Severity: notifications.SeverityInfo,
		Title:    "User Added",
		Message:  "User john was added",
	}

	id, err := svc.CreateNotification(notif, "view_notifications_user_mgmt")
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id 42, got %d", id)
	}
	if !repo.createCalled {
		t.Error("expected CreateNotification to be called")
	}
	if !repo.assignCalled {
		t.Error("expected AssignNotificationToUsersWithPermission to be called")
	}
	if repo.lastPermission != "view_notifications_user_mgmt" {
		t.Errorf("expected permission view_notifications_user_mgmt, got %s", repo.lastPermission)
	}
}
```

**Step 3: Run test to verify it fails**

Run: `cd sensor_hub && go test ./service/... -run TestNotificationService -v`
Expected: FAIL - undefined: NewNotificationService

**Step 4: Implement service**

Create `sensor_hub/service/notification_service.go`:

```go
package service

import (
	database "example/sensorHub/db"
	"example/sensorHub/notifications"
	"fmt"
)

type NotificationService struct {
	repo database.NotificationRepository
}

func NewNotificationService(repo database.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) CreateNotification(notif notifications.Notification, targetPermission string) (int, error) {
	id, err := s.repo.CreateNotification(notif)
	if err != nil {
		return 0, fmt.Errorf("failed to create notification: %w", err)
	}

	if err := s.repo.AssignNotificationToUsersWithPermission(id, targetPermission); err != nil {
		return 0, fmt.Errorf("failed to assign notification: %w", err)
	}

	return id, nil
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
```

**Step 5: Run tests to verify they pass**

Run: `cd sensor_hub && go test ./service/... -run TestNotificationService -v`
Expected: PASS

**Step 6: Commit**

```bash
git add sensor_hub/service/notification_service*.go
git commit -m "feat(service): add notification service with channel preference support"
```

<!-- Tasks 7-9 details below -->

---

### Task 7: Integrate with User Management Events

**Files:**
- Modify: `sensor_hub/service/user_service.go`
- Modify: `sensor_hub/api/users_api.go`

**Step 1: Add notification dependency to user service**

Modify `sensor_hub/service/user_service.go` to add notification triggers:

```go
// Add to UserService struct
type UserService struct {
	userRepo    database.UserRepository
	notifSvc    NotificationServiceInterface // Add this field
}

// Update constructor
func NewUserService(u database.UserRepository, n NotificationServiceInterface) *UserService {
	return &UserService{userRepo: u, notifSvc: n}
}

// Add helper method for creating user notifications
func (s *UserService) notifyUserEvent(action, username string, metadata map[string]interface{}) {
	if s.notifSvc == nil {
		return
	}
	notif := notifications.Notification{
		Category: notifications.CategoryUserManagement,
		Severity: notifications.SeverityInfo,
		Title:    fmt.Sprintf("User %s", action),
		Message:  fmt.Sprintf("User '%s' was %s", username, action),
		Metadata: metadata,
	}
	go s.notifSvc.CreateNotification(notif, "view_notifications_user_mgmt")
}
```

**Step 2: Add notification calls to user operations**

In `CreateUser` after successful creation:
```go
s.notifyUserEvent("added", user.Username, map[string]interface{}{"user_id": id})
```

In `DeleteUser` after successful deletion:
```go
// Get username before deletion for notification
user, _ := s.userRepo.GetUserById(userId)
// ... delete logic ...
if user != nil {
	s.notifyUserEvent("removed", user.Username, map[string]interface{}{"user_id": userId})
}
```

In `SetUserRoles` after successful role change:
```go
user, _ := s.userRepo.GetUserById(userId)
if user != nil {
	s.notifyUserEvent("role changed", user.Username, map[string]interface{}{
		"user_id": userId,
		"roles":   roles,
	})
}
```

**Step 3: Update API initialization**

Modify `sensor_hub/api/users_api.go` and `sensor_hub/main.go` to pass notification service to user service.

**Step 4: Run existing tests to ensure no regressions**

Run: `cd sensor_hub && go test ./service/... -run TestUserService -v`
Expected: Tests may need mock updates for new constructor signature

**Step 5: Commit**

```bash
git add sensor_hub/service/user_service.go sensor_hub/api/users_api.go sensor_hub/main.go
git commit -m "feat(users): integrate notification creation on user events"
```

---

### Task 8: Integrate with Sensor Configuration Events

**Files:**
- Modify: `sensor_hub/service/sensor_service.go`
- Modify: `sensor_hub/api/sensor_api.go`

**Step 1: Add notification dependency to sensor service**

Modify `sensor_hub/service/sensor_service.go`:

```go
// Add to SensorService struct
type SensorService struct {
	sensorRepo database.SensorRepository
	notifSvc   NotificationServiceInterface // Add this field
}

// Update constructor - add notifSvc parameter
func NewSensorService(s database.SensorRepository, n NotificationServiceInterface) *SensorService {
	return &SensorService{sensorRepo: s, notifSvc: n}
}

// Add helper method
func (s *SensorService) notifyConfigEvent(action, sensorName string, metadata map[string]interface{}) {
	if s.notifSvc == nil {
		return
	}
	notif := notifications.Notification{
		Category: notifications.CategoryConfigChange,
		Severity: notifications.SeverityInfo,
		Title:    fmt.Sprintf("Sensor %s", action),
		Message:  fmt.Sprintf("Sensor '%s' was %s", sensorName, action),
		Metadata: metadata,
	}
	go s.notifSvc.CreateNotification(notif, "view_notifications_config")
}
```

**Step 2: Add notification calls to sensor operations**

In `CreateSensor` (or equivalent add sensor method):
```go
s.notifyConfigEvent("added", sensor.Name, map[string]interface{}{"sensor_id": id, "type": sensor.Type})
```

In `DeleteSensor`:
```go
// Get sensor name before deletion
sensor, _ := s.sensorRepo.GetSensorById(sensorId)
// ... delete logic ...
if sensor != nil {
	s.notifyConfigEvent("removed", sensor.Name, map[string]interface{}{"sensor_id": sensorId})
}
```

**Step 3: Update API initialization**

Update `sensor_hub/main.go` to pass notification service to sensor service.

**Step 4: Run existing tests**

Run: `cd sensor_hub && go test ./service/... -run TestSensorService -v`
Expected: Tests may need mock updates

**Step 5: Commit**

```bash
git add sensor_hub/service/sensor_service.go sensor_hub/api/sensor_api.go sensor_hub/main.go
git commit -m "feat(sensors): integrate notification creation on config events"
```

---

### Task 9: Integrate with AlertService for Threshold Alerts

**Files:**
- Modify: `sensor_hub/alerting/service.go`
- Modify: `sensor_hub/alerting/service_test.go`

**Step 1: Add notification and preference dependencies**

Modify `sensor_hub/alerting/service.go`:

```go
type NotificationCreator interface {
	CreateNotification(notif notifications.Notification, targetPermission string) (int, error)
}

type ChannelPreferenceChecker interface {
	ShouldNotifyChannel(userID int, category notifications.NotificationCategory, channel string) (bool, error)
}

type AlertService struct {
	repo      AlertRepository
	notifier  Notifier
	notifSvc  NotificationCreator       // Add this
	prefSvc   ChannelPreferenceChecker  // Add this (can be same service)
}

func NewAlertService(repo AlertRepository, notifier Notifier, notifSvc NotificationCreator) *AlertService {
	return &AlertService{
		repo:     repo,
		notifier: notifier,
		notifSvc: notifSvc,
	}
}
```

**Step 2: Modify ProcessReadingAlert to create in-app notifications**

```go
func (s *AlertService) ProcessReadingAlert(sensorID int, sensorName, sensorType string, numericValue float64, statusValue string) error {
	// ... existing rule fetching and validation ...

	shouldAlert, reason := rule.ShouldAlert(numericValue, statusValue)
	if !shouldAlert {
		return nil
	}

	if rule.IsRateLimited() {
		return nil
	}

	// Record the alert (shared rate limit for both channels)
	err = s.repo.RecordAlertSent(rule.ID, sensorID, reason, numericValue, statusValue)
	if err != nil {
		log.Printf("Warning: failed to record alert: %v", err)
	}

	// Send email notification (existing behavior)
	err = s.notifier.SendAlert(sensorName, sensorType, reason, numericValue, statusValue)
	if err != nil {
		log.Printf("Warning: failed to send email alert: %v", err)
	}

	// Create in-app notification
	if s.notifSvc != nil {
		notif := notifications.Notification{
			Category: notifications.CategoryThresholdAlert,
			Severity: s.determineSeverity(reason),
			Title:    fmt.Sprintf("Alert: %s", sensorName),
			Message:  reason,
			Metadata: map[string]interface{}{
				"sensor_id":   sensorID,
				"sensor_name": sensorName,
				"sensor_type": sensorType,
				"value":       numericValue,
				"status":      statusValue,
			},
		}
		go s.notifSvc.CreateNotification(notif, "view_alerts")
	}

	return nil
}

func (s *AlertService) determineSeverity(reason string) notifications.NotificationSeverity {
	// High threshold exceeded = warning, other conditions = info
	if strings.Contains(strings.ToLower(reason), "high") || strings.Contains(strings.ToLower(reason), "exceeded") {
		return notifications.SeverityWarning
	}
	return notifications.SeverityInfo
}
```

**Step 3: Update AlertService initialization in main.go**

```go
notificationService := service.NewNotificationService(notificationRepo)
alertService := alerting.NewAlertService(alertRepo, smtpNotifier, notificationService)
```

**Step 4: Update tests with mock notification service**

Modify `sensor_hub/alerting/service_test.go` to add mock for NotificationCreator.

**Step 5: Run tests**

Run: `cd sensor_hub && go test ./alerting/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add sensor_hub/alerting/service.go sensor_hub/alerting/service_test.go sensor_hub/main.go
git commit -m "feat(alerting): create in-app notifications for threshold alerts"
```

<!-- Tasks 10-13 details below -->

---

### Task 10: Notification API Endpoints

**Files:**
- Create: `sensor_hub/api/notification_api.go`
- Create: `sensor_hub/api/notification_routes.go`
- Create: `sensor_hub/api/notification_api_test.go`

**Step 1: Create routes file**

Create `sensor_hub/api/notification_routes.go`:

```go
package api

import (
	"example/sensorHub/api/middleware"
	"example/sensorHub/service"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(r *gin.RouterGroup, notifSvc service.NotificationServiceInterface, roleRepo middleware.RoleRepository) {
	InitNotificationsAPI(notifSvc)

	notif := r.Group("/notifications")
	notif.Use(middleware.RequirePermission("view_notifications", roleRepo))
	{
		notif.GET("/", listNotificationsHandler)
		notif.GET("/unread-count", unreadCountHandler)
		notif.POST("/:id/read", markAsReadHandler)
		notif.POST("/:id/dismiss", dismissHandler)
		notif.POST("/bulk/read", bulkMarkAsReadHandler)
		notif.POST("/bulk/dismiss", bulkDismissHandler)
	}
}
```

**Step 2: Create API handlers**

Create `sensor_hub/api/notification_api.go`:

```go
package api

import (
	"example/sensorHub/service"
	"example/sensorHub/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var notificationService service.NotificationServiceInterface

func InitNotificationsAPI(n service.NotificationServiceInterface) {
	notificationService = n
}

func listNotificationsHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	includeDismissed := ctx.Query("include_dismissed") == "true"

	if limit > 100 {
		limit = 100
	}

	notifications, err := notificationService.GetNotificationsForUser(user.Id, limit, offset, includeDismissed)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, notifications)
}

func unreadCountHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	count, err := notificationService.GetUnreadCount(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

func markAsReadHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)
	notifID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := notificationService.MarkAsRead(user.Id, notifID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func dismissHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)
	notifID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := notificationService.Dismiss(user.Id, notifID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func bulkMarkAsReadHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	if err := notificationService.BulkMarkAsRead(user.Id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func bulkDismissHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	if err := notificationService.BulkDismiss(user.Id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
```

**Step 3: Write API test**

Create `sensor_hub/api/notification_api_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/sensorHub/notifications"
	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
)

type mockNotifService struct{}

func (m *mockNotifService) GetNotificationsForUser(userID, limit, offset int, includeDismissed bool) ([]notifications.UserNotification, error) {
	return []notifications.UserNotification{
		{ID: 1, NotificationID: 1, IsRead: false},
	}, nil
}
func (m *mockNotifService) GetUnreadCount(userID int) (int, error) { return 5, nil }
func (m *mockNotifService) MarkAsRead(userID, notificationID int) error { return nil }
func (m *mockNotifService) Dismiss(userID, notificationID int) error { return nil }
func (m *mockNotifService) BulkMarkAsRead(userID int) error { return nil }
func (m *mockNotifService) BulkDismiss(userID int) error { return nil }
func (m *mockNotifService) CreateNotification(n notifications.Notification, p string) (int, error) { return 1, nil }
func (m *mockNotifService) GetChannelPreferences(userID int) ([]notifications.ChannelPreference, error) { return nil, nil }
func (m *mockNotifService) SetChannelPreference(userID int, p notifications.ChannelPreference) error { return nil }
func (m *mockNotifService) ShouldNotifyChannel(userID int, cat notifications.NotificationCategory, ch string) (bool, error) { return true, nil }

func TestUnreadCountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})

	unreadCountHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]int
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"] != 5 {
		t.Errorf("expected count 5, got %d", resp["count"])
	}
}
```

**Step 4: Run tests**

Run: `cd sensor_hub && go test ./api/... -run TestNotification -v`
Expected: PASS

**Step 5: Register routes in main.go**

Add to `sensor_hub/main.go`:
```go
api.RegisterNotificationRoutes(apiGroup, notificationService, roleRepo)
```

**Step 6: Commit**

```bash
git add sensor_hub/api/notification_*.go sensor_hub/main.go
git commit -m "feat(api): add notification REST endpoints"
```

---

### Task 11: Channel Preference API Endpoints

**Files:**
- Modify: `sensor_hub/api/notification_routes.go`
- Modify: `sensor_hub/api/notification_api.go`

**Step 1: Add preference routes**

Append to `sensor_hub/api/notification_routes.go` inside RegisterNotificationRoutes:

```go
prefs := r.Group("/notification-preferences")
prefs.Use(middleware.RequirePermission("manage_notifications", roleRepo))
{
	prefs.GET("/", getPreferencesHandler)
	prefs.PUT("/", setPreferenceHandler)
}
```

**Step 2: Add preference handlers**

Append to `sensor_hub/api/notification_api.go`:

```go
func getPreferencesHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	prefs, err := notificationService.GetChannelPreferences(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, prefs)
}

type setPreferenceRequest struct {
	Category     string `json:"category" binding:"required"`
	EmailEnabled bool   `json:"email_enabled"`
	InAppEnabled bool   `json:"inapp_enabled"`
}

func setPreferenceHandler(ctx *gin.Context) {
	user := ctx.MustGet("currentUser").(*types.User)

	var req setPreferenceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pref := notifications.ChannelPreference{
		Category:     notifications.NotificationCategory(req.Category),
		EmailEnabled: req.EmailEnabled,
		InAppEnabled: req.InAppEnabled,
	}

	if err := notificationService.SetChannelPreference(user.Id, pref); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}
```

**Step 3: Run tests**

Run: `cd sensor_hub && go test ./api/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add sensor_hub/api/notification_*.go
git commit -m "feat(api): add channel preference endpoints"
```

---

### Task 12: WebSocket Push Integration

**Files:**
- Modify: `sensor_hub/ws/websocket.go`
- Modify: `sensor_hub/service/notification_service.go`

**Step 1: Add notification message type to WebSocket**

Check existing `sensor_hub/ws/websocket.go` for WebSocket hub structure. Add a broadcast method for notifications:

```go
type NotificationMessage struct {
	Type         string                        `json:"type"` // "notification"
	Notification *notifications.UserNotification `json:"notification"`
}

// Add to Hub struct or create new method
func (h *Hub) BroadcastToUser(userID int, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
```

**Step 2: Create WebSocket notifier interface**

Add to `sensor_hub/service/notification_service.go`:

```go
type WebSocketNotifier interface {
	BroadcastToUser(userID int, message interface{})
}

// Update NotificationService struct
type NotificationService struct {
	repo database.NotificationRepository
	ws   WebSocketNotifier
}

func NewNotificationService(repo database.NotificationRepository, ws WebSocketNotifier) *NotificationService {
	return &NotificationService{repo: repo, ws: ws}
}
```

**Step 3: Push notification on creation**

Modify `CreateNotification` in notification_service.go:

```go
func (s *NotificationService) CreateNotification(notif notifications.Notification, targetPermission string) (int, error) {
	id, err := s.repo.CreateNotification(notif)
	if err != nil {
		return 0, err
	}

	if err := s.repo.AssignNotificationToUsersWithPermission(id, targetPermission); err != nil {
		return 0, err
	}

	// Push to connected users via WebSocket
	if s.ws != nil {
		notif.ID = id
		go s.pushToConnectedUsers(id, &notif, targetPermission)
	}

	return id, nil
}

func (s *NotificationService) pushToConnectedUsers(notifID int, notif *notifications.Notification, permission string) {
	// Get user IDs with this permission from repo and push
	// For simplicity, the WebSocket hub can track which users have which permissions
	// Or we query and iterate
}
```

**Step 4: Update main.go to wire WebSocket hub**

```go
notificationService := service.NewNotificationService(notificationRepo, wsHub)
```

**Step 5: Run tests**

Run: `cd sensor_hub && go test ./... -v`
Expected: PASS (with mock WebSocket)

**Step 6: Commit**

```bash
git add sensor_hub/ws/websocket.go sensor_hub/service/notification_service.go sensor_hub/main.go
git commit -m "feat(ws): add WebSocket push for real-time notifications"
```

---

### Task 13: Cleanup Service Integration

**Files:**
- Modify: `sensor_hub/service/cleanup_service.go`
- Modify: `sensor_hub/application_properties/application_properties.go`

**Step 1: Add notification retention config**

Append to `application_properties.go` struct:

```go
NotificationRetentionDays int `properties:"notification.retention.days,default=90"`
```

**Step 2: Add notification cleanup to cleanup service**

Modify `sensor_hub/service/cleanup_service.go`:

```go
// Add to CleanupService struct
type CleanupService struct {
	// ... existing fields
	notifRepo database.NotificationRepository
}

// Add cleanup method
func (s *CleanupService) CleanupOldNotifications() (int64, error) {
	retentionDays := 90
	if appProps.AppConfig != nil && appProps.AppConfig.NotificationRetentionDays > 0 {
		retentionDays = appProps.AppConfig.NotificationRetentionDays
	}
	
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.notifRepo.DeleteOldNotifications(cutoff)
}

// Call in scheduled cleanup
func (s *CleanupService) RunScheduledCleanup() {
	// ... existing cleanup tasks ...
	
	deleted, err := s.CleanupOldNotifications()
	if err != nil {
		log.Printf("Failed to cleanup old notifications: %v", err)
	} else {
		log.Printf("Cleaned up %d old notifications", deleted)
	}
}
```

**Step 3: Run tests**

Run: `cd sensor_hub && go test ./service/... -run TestCleanup -v`
Expected: PASS

**Step 4: Commit**

```bash
git add sensor_hub/service/cleanup_service.go sensor_hub/application_properties/application_properties.go
git commit -m "feat(cleanup): add 90-day notification auto-purge"
```

<!-- Tasks 14-17 details below -->

---

### Task 14: Frontend - Notification Types and API Client

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/types/Notification.ts`
- Create: `sensor_hub/ui/sensor_hub_ui/src/api/Notifications.ts`

**Step 1: Create TypeScript types**

Create `sensor_hub/ui/sensor_hub_ui/src/types/Notification.ts`:

```typescript
export type NotificationCategory = 'threshold_alert' | 'user_management' | 'config_change';
export type NotificationSeverity = 'info' | 'warning' | 'error';

export interface Notification {
  id: number;
  category: NotificationCategory;
  severity: NotificationSeverity;
  title: string;
  message: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface UserNotification {
  id: number;
  user_id: number;
  notification_id: number;
  is_read: boolean;
  is_dismissed: boolean;
  read_at?: string;
  dismissed_at?: string;
  notification?: Notification;
}

export interface ChannelPreference {
  category: NotificationCategory;
  email_enabled: boolean;
  inapp_enabled: boolean;
}

export interface UnreadCountResponse {
  count: number;
}
```

**Step 2: Create API client**

Create `sensor_hub/ui/sensor_hub_ui/src/api/Notifications.ts`:

```typescript
import client from './Client';
import type { UserNotification, ChannelPreference, UnreadCountResponse } from '../types/Notification';

export async function getNotifications(
  limit = 50,
  offset = 0,
  includeDismissed = false
): Promise<UserNotification[]> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
  });
  if (includeDismissed) {
    params.set('include_dismissed', 'true');
  }
  const response = await client.get(`/notifications/?${params}`);
  return response.data;
}

export async function getUnreadCount(): Promise<number> {
  const response = await client.get<UnreadCountResponse>('/notifications/unread-count');
  return response.data.count;
}

export async function markAsRead(notificationId: number): Promise<void> {
  await client.post(`/notifications/${notificationId}/read`);
}

export async function dismissNotification(notificationId: number): Promise<void> {
  await client.post(`/notifications/${notificationId}/dismiss`);
}

export async function bulkMarkAsRead(): Promise<void> {
  await client.post('/notifications/bulk/read');
}

export async function bulkDismiss(): Promise<void> {
  await client.post('/notifications/bulk/dismiss');
}

export async function getChannelPreferences(): Promise<ChannelPreference[]> {
  const response = await client.get('/notification-preferences/');
  return response.data;
}

export async function setChannelPreference(pref: ChannelPreference): Promise<void> {
  await client.put('/notification-preferences/', pref);
}
```

**Step 3: Run type check**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds without type errors

**Step 4: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/types/Notification.ts sensor_hub/ui/sensor_hub_ui/src/api/Notifications.ts
git commit -m "feat(ui): add notification types and API client"
```

---

### Task 15: Frontend - Notification Context Provider

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/providers/NotificationContext.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/providers/NotificationProvider.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/ProviderWrapper.tsx`

**Step 1: Create context**

Create `sensor_hub/ui/sensor_hub_ui/src/providers/NotificationContext.tsx`:

```typescript
import { createContext, useContext } from 'react';
import type { UserNotification } from '../types/Notification';

interface NotificationContextType {
  notifications: UserNotification[];
  unreadCount: number;
  loading: boolean;
  refresh: () => Promise<void>;
  markAsRead: (notificationId: number) => Promise<void>;
  dismiss: (notificationId: number) => Promise<void>;
}

export const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within NotificationProvider');
  }
  return context;
}
```

**Step 2: Create provider with WebSocket subscription**

Create `sensor_hub/ui/sensor_hub_ui/src/providers/NotificationProvider.tsx`:

```typescript
import React, { useEffect, useState, useCallback, useRef } from 'react';
import { NotificationContext } from './NotificationContext';
import { useAuth } from './AuthContext';
import * as NotificationsApi from '../api/Notifications';
import type { UserNotification } from '../types/Notification';

export default function NotificationProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const [notifications, setNotifications] = useState<UserNotification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  const refresh = useCallback(async () => {
    if (!user) return;
    setLoading(true);
    try {
      const [notifs, count] = await Promise.all([
        NotificationsApi.getNotifications(5, 0, false),
        NotificationsApi.getUnreadCount(),
      ]);
      setNotifications(notifs);
      setUnreadCount(count);
    } catch (e) {
      console.error('Failed to fetch notifications:', e);
    } finally {
      setLoading(false);
    }
  }, [user]);

  const markAsRead = useCallback(async (notificationId: number) => {
    await NotificationsApi.markAsRead(notificationId);
    setNotifications(prev =>
      prev.map(n => n.notification_id === notificationId ? { ...n, is_read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, []);

  const dismiss = useCallback(async (notificationId: number) => {
    await NotificationsApi.dismissNotification(notificationId);
    setNotifications(prev => prev.filter(n => n.notification_id !== notificationId));
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, []);

  // WebSocket subscription for real-time updates
  useEffect(() => {
    if (!user) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'notification') {
          setNotifications(prev => [data.notification, ...prev.slice(0, 4)]);
          setUnreadCount(prev => prev + 1);
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    ws.onclose = () => {
      // Reconnect after delay
      setTimeout(() => {
        if (user) refresh();
      }, 5000);
    };

    return () => {
      ws.close();
    };
  }, [user, refresh]);

  // Initial fetch
  useEffect(() => {
    refresh();
  }, [refresh]);

  return (
    <NotificationContext.Provider value={{ notifications, unreadCount, loading, refresh, markAsRead, dismiss }}>
      {children}
    </NotificationContext.Provider>
  );
}
```

**Step 3: Add provider to wrapper**

Modify `sensor_hub/ui/sensor_hub_ui/src/ProviderWrapper.tsx`:

```typescript
import NotificationProvider from './providers/NotificationProvider';

// Wrap children with NotificationProvider inside AuthProvider
<AuthProvider>
  <NotificationProvider>
    {children}
  </NotificationProvider>
</AuthProvider>
```

**Step 4: Run build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/providers/Notification*.tsx sensor_hub/ui/sensor_hub_ui/src/ProviderWrapper.tsx
git commit -m "feat(ui): add notification context with WebSocket subscription"
```

---

### Task 16: Frontend - Bell Icon Component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/NotificationBell.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx`

**Step 1: Create bell icon component**

Create `sensor_hub/ui/sensor_hub_ui/src/components/NotificationBell.tsx`:

```typescript
import { Badge, IconButton } from '@mui/material';
import NotificationsIcon from '@mui/icons-material/Notifications';
import { useNotifications } from '../providers/NotificationContext';

interface NotificationBellProps {
  onClick: (event: React.MouseEvent<HTMLElement>) => void;
}

export default function NotificationBell({ onClick }: NotificationBellProps) {
  const { unreadCount } = useNotifications();

  return (
    <IconButton color="inherit" onClick={onClick} aria-label="notifications">
      <Badge badgeContent={unreadCount} color="error" max={99}>
        <NotificationsIcon />
      </Badge>
    </IconButton>
  );
}
```

**Step 2: Add to TopAppBar**

Modify `sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx`:

Add import:
```typescript
import NotificationBell from '../components/NotificationBell';
import { hasPerm } from '../tools/Utils';
```

Add state for notification dropdown:
```typescript
const [notifAnchor, setNotifAnchor] = useState<null | HTMLElement>(null);
const openNotif = Boolean(notifAnchor);
```

Add bell icon before theme switcher (around line 126, after pageTitle Typography):
```typescript
{user && hasPerm(user, 'view_notifications') && (
  <NotificationBell onClick={(e) => setNotifAnchor(e.currentTarget)} />
)}
```

**Step 3: Run build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/components/NotificationBell.tsx sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx
git commit -m "feat(ui): add notification bell icon with unread badge"
```

---

### Task 17: Frontend - Notification Dropdown

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/NotificationDropdown.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx`

**Step 1: Create dropdown component**

Create `sensor_hub/ui/sensor_hub_ui/src/components/NotificationDropdown.tsx`:

```typescript
import {
  Menu,
  MenuItem,
  ListItemText,
  ListItemIcon,
  Typography,
  Divider,
  Box,
  IconButton,
  Chip,
} from '@mui/material';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import ErrorIcon from '@mui/icons-material/Error';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import { useNavigate } from 'react-router';
import { useNotifications } from '../providers/NotificationContext';
import type { NotificationSeverity } from '../types/Notification';

interface NotificationDropdownProps {
  anchorEl: HTMLElement | null;
  open: boolean;
  onClose: () => void;
}

const severityIcons: Record<NotificationSeverity, React.ReactNode> = {
  info: <InfoIcon color="info" fontSize="small" />,
  warning: <WarningIcon color="warning" fontSize="small" />,
  error: <ErrorIcon color="error" fontSize="small" />,
};

export default function NotificationDropdown({ anchorEl, open, onClose }: NotificationDropdownProps) {
  const { notifications, markAsRead, dismiss } = useNotifications();
  const navigate = useNavigate();

  const handleViewAll = () => {
    onClose();
    navigate('/notifications');
  };

  const handleMarkRead = async (e: React.MouseEvent, notifId: number) => {
    e.stopPropagation();
    await markAsRead(notifId);
  };

  const handleDismiss = async (e: React.MouseEvent, notifId: number) => {
    e.stopPropagation();
    await dismiss(notifId);
  };

  return (
    <Menu
      anchorEl={anchorEl}
      open={open}
      onClose={onClose}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      transformOrigin={{ vertical: 'top', horizontal: 'right' }}
      PaperProps={{ sx: { width: 360, maxHeight: 400 } }}
    >
      <Box sx={{ px: 2, py: 1 }}>
        <Typography variant="subtitle1" fontWeight="bold">Notifications</Typography>
      </Box>
      <Divider />
      
      {notifications.length === 0 ? (
        <MenuItem disabled>
          <Typography variant="body2" color="text.secondary">No notifications</Typography>
        </MenuItem>
      ) : (
        notifications.slice(0, 5).map((un) => (
          <MenuItem
            key={un.id}
            sx={{
              opacity: un.is_read ? 0.7 : 1,
              backgroundColor: un.is_read ? 'transparent' : 'action.hover',
            }}
          >
            <ListItemIcon>
              {un.notification && severityIcons[un.notification.severity]}
            </ListItemIcon>
            <ListItemText
              primary={un.notification?.title}
              secondary={un.notification?.message}
              primaryTypographyProps={{ noWrap: true, variant: 'body2' }}
              secondaryTypographyProps={{ noWrap: true, variant: 'caption' }}
            />
            <Box sx={{ display: 'flex', gap: 0.5, ml: 1 }}>
              {!un.is_read && (
                <IconButton size="small" onClick={(e) => handleMarkRead(e, un.notification_id)} title="Mark read">
                  <CheckIcon fontSize="small" />
                </IconButton>
              )}
              <IconButton size="small" onClick={(e) => handleDismiss(e, un.notification_id)} title="Dismiss">
                <CloseIcon fontSize="small" />
              </IconButton>
            </Box>
          </MenuItem>
        ))
      )}
      
      <Divider />
      <MenuItem onClick={handleViewAll}>
        <Typography variant="body2" color="primary" sx={{ width: '100%', textAlign: 'center' }}>
          View all notifications
        </Typography>
      </MenuItem>
    </Menu>
  );
}
```

**Step 2: Add dropdown to TopAppBar**

Modify `sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx`:

Add import:
```typescript
import NotificationDropdown from '../components/NotificationDropdown';
```

Add dropdown component after NotificationBell:
```typescript
<NotificationDropdown
  anchorEl={notifAnchor}
  open={openNotif}
  onClose={() => setNotifAnchor(null)}
/>
```

**Step 3: Run build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/components/NotificationDropdown.tsx sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx
git commit -m "feat(ui): add notification dropdown with quick actions"
```

<!-- Tasks 18-20 details below -->

---

### Task 18: Frontend - Notifications Page

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/pages/notifications/NotificationsPage.tsx`

**Step 1: Create notifications page**

Create `sensor_hub/ui/sensor_hub_ui/src/pages/notifications/NotificationsPage.tsx`:

```typescript
import { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Typography,
  Paper,
  Button,
  ButtonGroup,
  Chip,
  IconButton,
  FormControlLabel,
  Switch,
  CircularProgress,
  Alert,
} from '@mui/material';
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import DoneAllIcon from '@mui/icons-material/DoneAll';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import ErrorIcon from '@mui/icons-material/Error';
import TopAppBar from '../../navigation/TopAppBar';
import * as NotificationsApi from '../../api/Notifications';
import type { UserNotification, NotificationCategory, NotificationSeverity } from '../../types/Notification';
import { useNotifications } from '../../providers/NotificationContext';
import { useIsMobile } from '../../hooks/useMobile';

const severityIcons: Record<NotificationSeverity, React.ReactNode> = {
  info: <InfoIcon color="info" />,
  warning: <WarningIcon color="warning" />,
  error: <ErrorIcon color="error" />,
};

const categoryLabels: Record<NotificationCategory, string> = {
  threshold_alert: 'Threshold Alert',
  user_management: 'User Management',
  config_change: 'Configuration',
};

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<UserNotification[]>([]);
  const [loading, setLoading] = useState(true);
  const [includeDismissed, setIncludeDismissed] = useState(false);
  const [categoryFilter, setCategoryFilter] = useState<NotificationCategory | 'all'>('all');
  const { refresh: refreshContext } = useNotifications();
  const isMobile = useIsMobile();

  const fetchNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const data = await NotificationsApi.getNotifications(100, 0, includeDismissed);
      setNotifications(data);
    } catch (e) {
      console.error('Failed to fetch notifications:', e);
    } finally {
      setLoading(false);
    }
  }, [includeDismissed]);

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  const handleMarkRead = async (notifId: number) => {
    await NotificationsApi.markAsRead(notifId);
    setNotifications(prev =>
      prev.map(n => n.notification_id === notifId ? { ...n, is_read: true } : n)
    );
    refreshContext();
  };

  const handleDismiss = async (notifId: number) => {
    await NotificationsApi.dismissNotification(notifId);
    if (includeDismissed) {
      setNotifications(prev =>
        prev.map(n => n.notification_id === notifId ? { ...n, is_dismissed: true } : n)
      );
    } else {
      setNotifications(prev => prev.filter(n => n.notification_id !== notifId));
    }
    refreshContext();
  };

  const handleBulkMarkRead = async () => {
    await NotificationsApi.bulkMarkAsRead();
    setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
    refreshContext();
  };

  const handleBulkDismiss = async () => {
    await NotificationsApi.bulkDismiss();
    if (includeDismissed) {
      setNotifications(prev => prev.map(n => ({ ...n, is_dismissed: true })));
    } else {
      setNotifications([]);
    }
    refreshContext();
  };

  const filteredNotifications = notifications.filter(n => {
    if (categoryFilter !== 'all' && n.notification?.category !== categoryFilter) {
      return false;
    }
    return true;
  });

  const columns: GridColDef[] = [
    {
      field: 'severity',
      headerName: '',
      width: 50,
      renderCell: (params: GridRenderCellParams<UserNotification>) =>
        params.row.notification ? severityIcons[params.row.notification.severity] : null,
    },
    {
      field: 'title',
      headerName: 'Title',
      flex: 1,
      minWidth: 200,
      valueGetter: (value, row) => row.notification?.title,
    },
    {
      field: 'message',
      headerName: 'Message',
      flex: 2,
      minWidth: 300,
      valueGetter: (value, row) => row.notification?.message,
    },
    {
      field: 'category',
      headerName: 'Category',
      width: 150,
      renderCell: (params: GridRenderCellParams<UserNotification>) => (
        <Chip
          label={params.row.notification ? categoryLabels[params.row.notification.category] : ''}
          size="small"
          variant="outlined"
        />
      ),
    },
    {
      field: 'created_at',
      headerName: 'Time',
      width: 180,
      valueGetter: (value, row) =>
        row.notification?.created_at
          ? new Date(row.notification.created_at).toLocaleString()
          : '',
    },
    {
      field: 'status',
      headerName: 'Status',
      width: 100,
      renderCell: (params: GridRenderCellParams<UserNotification>) => (
        <Chip
          label={params.row.is_dismissed ? 'Dismissed' : params.row.is_read ? 'Read' : 'Unread'}
          size="small"
          color={params.row.is_dismissed ? 'default' : params.row.is_read ? 'success' : 'primary'}
        />
      ),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 100,
      sortable: false,
      renderCell: (params: GridRenderCellParams<UserNotification>) => (
        <Box>
          {!params.row.is_read && (
            <IconButton size="small" onClick={() => handleMarkRead(params.row.notification_id)} title="Mark read">
              <CheckIcon fontSize="small" />
            </IconButton>
          )}
          {!params.row.is_dismissed && (
            <IconButton size="small" onClick={() => handleDismiss(params.row.notification_id)} title="Dismiss">
              <CloseIcon fontSize="small" />
            </IconButton>
          )}
        </Box>
      ),
    },
  ];

  return (
    <>
      <TopAppBar pageTitle="Notifications" />
      <Box sx={{ p: 2 }}>
        <Paper sx={{ p: 2, mb: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2 }}>
            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
              <ButtonGroup size="small">
                <Button
                  variant={categoryFilter === 'all' ? 'contained' : 'outlined'}
                  onClick={() => setCategoryFilter('all')}
                >
                  All
                </Button>
                <Button
                  variant={categoryFilter === 'threshold_alert' ? 'contained' : 'outlined'}
                  onClick={() => setCategoryFilter('threshold_alert')}
                >
                  Alerts
                </Button>
                <Button
                  variant={categoryFilter === 'user_management' ? 'contained' : 'outlined'}
                  onClick={() => setCategoryFilter('user_management')}
                >
                  Users
                </Button>
                <Button
                  variant={categoryFilter === 'config_change' ? 'contained' : 'outlined'}
                  onClick={() => setCategoryFilter('config_change')}
                >
                  Config
                </Button>
              </ButtonGroup>

              <FormControlLabel
                control={
                  <Switch
                    checked={includeDismissed}
                    onChange={(e) => setIncludeDismissed(e.target.checked)}
                  />
                }
                label="Show dismissed"
              />
            </Box>

            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                startIcon={<DoneAllIcon />}
                onClick={handleBulkMarkRead}
                variant="outlined"
                size="small"
              >
                Mark all read
              </Button>
              <Button
                startIcon={<DeleteSweepIcon />}
                onClick={handleBulkDismiss}
                variant="outlined"
                color="warning"
                size="small"
              >
                Dismiss all
              </Button>
            </Box>
          </Box>
        </Paper>

        <Paper sx={{ height: isMobile ? 'auto' : 600 }}>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress />
            </Box>
          ) : filteredNotifications.length === 0 ? (
            <Alert severity="info" sx={{ m: 2 }}>No notifications to display</Alert>
          ) : (
            <DataGrid
              rows={filteredNotifications}
              columns={columns}
              pageSizeOptions={[10, 25, 50, 100]}
              initialState={{
                pagination: { paginationModel: { pageSize: 25 } },
                sorting: { sortModel: [{ field: 'created_at', sort: 'desc' }] },
              }}
              disableRowSelectionOnClick
              getRowId={(row) => row.id}
              sx={{
                '& .MuiDataGrid-row': {
                  cursor: 'pointer',
                },
              }}
            />
          )}
        </Paper>
      </Box>
    </>
  );
}
```

**Step 2: Run build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/notifications/
git commit -m "feat(ui): add notifications page with filtering and bulk actions"
```

---

### Task 19: Frontend - Channel Preferences Settings UI

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/pages/account/NotificationPreferences.tsx`

**Step 1: Create preferences page**

Create `sensor_hub/ui/sensor_hub_ui/src/pages/account/NotificationPreferences.tsx`:

```typescript
import { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Switch,
  CircularProgress,
  Alert,
  Snackbar,
} from '@mui/material';
import TopAppBar from '../../navigation/TopAppBar';
import * as NotificationsApi from '../../api/Notifications';
import type { ChannelPreference, NotificationCategory } from '../../types/Notification';

const categoryLabels: Record<NotificationCategory, { name: string; description: string }> = {
  threshold_alert: {
    name: 'Threshold Alerts',
    description: 'Notifications when sensor readings exceed configured thresholds',
  },
  user_management: {
    name: 'User Management',
    description: 'Notifications about user additions, removals, and role changes',
  },
  config_change: {
    name: 'Configuration Changes',
    description: 'Notifications about sensor additions and removals',
  },
};

export default function NotificationPreferences() {
  const [preferences, setPreferences] = useState<ChannelPreference[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  useEffect(() => {
    async function fetchPreferences() {
      try {
        const data = await NotificationsApi.getChannelPreferences();
        setPreferences(data);
      } catch (e) {
        console.error('Failed to fetch preferences:', e);
        setSnackbar({ open: true, message: 'Failed to load preferences', severity: 'error' });
      } finally {
        setLoading(false);
      }
    }
    fetchPreferences();
  }, []);

  const handleToggle = async (category: NotificationCategory, channel: 'email' | 'inapp', value: boolean) => {
    const pref = preferences.find(p => p.category === category);
    if (!pref) return;

    const updatedPref: ChannelPreference = {
      ...pref,
      email_enabled: channel === 'email' ? value : pref.email_enabled,
      inapp_enabled: channel === 'inapp' ? value : pref.inapp_enabled,
    };

    // Optimistic update
    setPreferences(prev =>
      prev.map(p => p.category === category ? updatedPref : p)
    );

    setSaving(true);
    try {
      await NotificationsApi.setChannelPreference(updatedPref);
      setSnackbar({ open: true, message: 'Preferences saved', severity: 'success' });
    } catch (e) {
      // Revert on error
      setPreferences(prev =>
        prev.map(p => p.category === category ? pref : p)
      );
      setSnackbar({ open: true, message: 'Failed to save preferences', severity: 'error' });
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <TopAppBar pageTitle="Notification Preferences" />
      <Box sx={{ p: 2 }}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Channel Preferences
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Configure which notification categories are delivered to each channel.
          </Typography>

          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress />
            </Box>
          ) : preferences.length === 0 ? (
            <Alert severity="warning">No preferences available</Alert>
          ) : (
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Category</TableCell>
                    <TableCell align="center">Email</TableCell>
                    <TableCell align="center">In-App</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {preferences.map((pref) => (
                    <TableRow key={pref.category}>
                      <TableCell>
                        <Typography variant="body1">
                          {categoryLabels[pref.category]?.name || pref.category}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {categoryLabels[pref.category]?.description}
                        </Typography>
                      </TableCell>
                      <TableCell align="center">
                        <Switch
                          checked={pref.email_enabled}
                          onChange={(e) => handleToggle(pref.category, 'email', e.target.checked)}
                          disabled={saving}
                        />
                      </TableCell>
                      <TableCell align="center">
                        <Switch
                          checked={pref.inapp_enabled}
                          onChange={(e) => handleToggle(pref.category, 'inapp', e.target.checked)}
                          disabled={saving}
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Paper>
      </Box>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar(prev => ({ ...prev, open: false }))}
      >
        <Alert severity={snackbar.severity} onClose={() => setSnackbar(prev => ({ ...prev, open: false }))}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}
```

**Step 2: Run build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/account/NotificationPreferences.tsx
git commit -m "feat(ui): add notification channel preferences settings page"
```

---

### Task 20: Frontend - Navigation Integration

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx`

**Step 1: Add routes**

Modify `sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx`:

Add imports:
```typescript
import NotificationsPage from '../pages/notifications/NotificationsPage';
import NotificationPreferences from '../pages/account/NotificationPreferences';
```

Add routes inside protected routes section:
```typescript
<Route path="/notifications" element={<NotificationsPage />} />
<Route path="/account/notification-preferences" element={<NotificationPreferences />} />
```

**Step 2: Add to navigation sidebar**

Modify `sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx`:

Add import:
```typescript
import NotificationsIcon from '@mui/icons-material/Notifications';
```

Add menu item (check hasPerm for view_notifications):
```typescript
{hasPerm(user, 'view_notifications') && (
  <ListItem disablePadding>
    <ListItemButton onClick={() => navigate('/notifications')}>
      <ListItemIcon><NotificationsIcon /></ListItemIcon>
      <ListItemText primary="Notifications" />
    </ListItemButton>
  </ListItem>
)}
```

**Step 3: Add notification preferences to account menu in TopAppBar**

Modify `sensor_hub/ui/sensor_hub_ui/src/navigation/TopAppBar.tsx`:

Add menu item after "Change password" (around line 77-80):
```typescript
if (hasPerm(user, 'manage_notifications')) {
  accountMenuItems.push(
    <MenuItem key="notifprefs" onClick={() => { handleAccountClose(); navigate('/account/notification-preferences'); }}>
      <ListItemIcon><NotificationsIcon fontSize="small" /></ListItemIcon>
      Notification preferences
    </MenuItem>
  );
}
```

Add import if not present:
```typescript
import NotificationsIcon from '@mui/icons-material/Notifications';
```

**Step 4: Run build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/navigation/
git commit -m "feat(ui): integrate notifications into navigation and routing"
```

<!-- Tasks 21-24 details below -->

---

### Task 21: Backend - Repository Unit Tests

**Files:**
- Modify: `sensor_hub/db/notification_repository_test.go`

**Step 1: Add comprehensive repository tests**

Expand `sensor_hub/db/notification_repository_test.go`:

```go
func TestNotificationRepository_AssignNotificationToUsersWithPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectExec("INSERT IGNORE INTO user_notifications").
		WithArgs(1, "view_notifications").
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = repo.AssignNotificationToUsersWithPermission(1, "view_notifications")
	if err != nil {
		t.Fatalf("AssignNotificationToUsersWithPermission failed: %v", err)
	}
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectExec("UPDATE user_notifications SET is_read = TRUE").
		WithArgs(1, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.MarkAsRead(1, 5)
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}
}

func TestNotificationRepository_DismissNotification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectExec("UPDATE user_notifications SET is_dismissed = TRUE").
		WithArgs(1, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DismissNotification(1, 5)
	if err != nil {
		t.Fatalf("DismissNotification failed: %v", err)
	}
}

func TestNotificationRepository_BulkMarkAsRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	mock.ExpectExec("UPDATE user_notifications SET is_read = TRUE").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 10))

	err = repo.BulkMarkAsRead(1)
	if err != nil {
		t.Fatalf("BulkMarkAsRead failed: %v", err)
	}
}

func TestNotificationRepository_DeleteOldNotifications(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	cutoff := time.Now().AddDate(0, 0, -90)

	mock.ExpectExec("DELETE FROM notifications WHERE created_at").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 5))

	deleted, err := repo.DeleteOldNotifications(cutoff)
	if err != nil {
		t.Fatalf("DeleteOldNotifications failed: %v", err)
	}
	if deleted != 5 {
		t.Errorf("expected 5 deleted, got %d", deleted)
	}
}
```

**Step 2: Run tests**

Run: `cd sensor_hub && go test ./db/... -run TestNotificationRepository -v`
Expected: PASS (all tests)

**Step 3: Commit**

```bash
git add sensor_hub/db/notification_repository_test.go
git commit -m "test(db): add comprehensive notification repository tests"
```

---

### Task 22: Backend - Service Unit Tests

**Files:**
- Modify: `sensor_hub/service/notification_service_test.go`

**Step 1: Add comprehensive service tests**

Expand `sensor_hub/service/notification_service_test.go`:

```go
func TestNotificationService_GetUnreadCount(t *testing.T) {
	repo := &mockNotificationRepo{}
	repo.unreadCount = 7
	svc := NewNotificationService(repo, nil)

	count, err := svc.GetUnreadCount(1)
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 7 {
		t.Errorf("expected count 7, got %d", count)
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	err := svc.MarkAsRead(1, 5)
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}
	if !repo.markAsReadCalled {
		t.Error("expected MarkAsRead to be called on repo")
	}
}

func TestNotificationService_Dismiss(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := NewNotificationService(repo, nil)

	err := svc.Dismiss(1, 5)
	if err != nil {
		t.Fatalf("Dismiss failed: %v", err)
	}
	if !repo.dismissCalled {
		t.Error("expected DismissNotification to be called on repo")
	}
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
	if err != nil {
		t.Fatalf("ShouldNotifyChannel failed: %v", err)
	}
	if !shouldNotify {
		t.Error("expected shouldNotify to be true for email")
	}

	shouldNotify, err = svc.ShouldNotifyChannel(1, notifications.CategoryUserManagement, "inapp")
	if err != nil {
		t.Fatalf("ShouldNotifyChannel failed: %v", err)
	}
	if shouldNotify {
		t.Error("expected shouldNotify to be false for inapp")
	}
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
	if err != nil {
		t.Fatalf("SetChannelPreference failed: %v", err)
	}
	if !repo.setPreferenceCalled {
		t.Error("expected SetChannelPreference to be called on repo")
	}
}
```

Also update mock with required fields:

```go
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

func (m *mockNotificationRepo) GetUnreadCountForUser(userID int) (int, error) {
	return m.unreadCount, nil
}

func (m *mockNotificationRepo) MarkAsRead(userID, notificationID int) error {
	m.markAsReadCalled = true
	return nil
}

func (m *mockNotificationRepo) DismissNotification(userID, notificationID int) error {
	m.dismissCalled = true
	return nil
}

func (m *mockNotificationRepo) GetChannelPreference(userID int, cat notifications.NotificationCategory) (*notifications.ChannelPreference, error) {
	if m.channelPref != nil {
		return m.channelPref, nil
	}
	return &notifications.ChannelPreference{EmailEnabled: true, InAppEnabled: true}, nil
}

func (m *mockNotificationRepo) SetChannelPreference(pref notifications.ChannelPreference) error {
	m.setPreferenceCalled = true
	return nil
}
```

**Step 2: Run tests**

Run: `cd sensor_hub && go test ./service/... -run TestNotificationService -v`
Expected: PASS (all tests)

**Step 3: Commit**

```bash
git add sensor_hub/service/notification_service_test.go
git commit -m "test(service): add comprehensive notification service tests"
```

---

### Task 23: Backend - API Integration Tests

**Files:**
- Modify: `sensor_hub/api/notification_api_test.go`

**Step 1: Add comprehensive API tests**

Expand `sensor_hub/api/notification_api_test.go`:

```go
func TestListNotificationsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Request, _ = http.NewRequest("GET", "/notifications/?limit=10&offset=0", nil)

	listNotificationsHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp []notifications.UserNotification
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 notification, got %d", len(resp))
	}
}

func TestMarkAsReadHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	c.Request, _ = http.NewRequest("POST", "/notifications/5/read", nil)

	markAsReadHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestDismissHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	c.Request, _ = http.NewRequest("POST", "/notifications/5/dismiss", nil)

	dismissHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestBulkMarkAsReadHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Request, _ = http.NewRequest("POST", "/notifications/bulk/read", nil)

	bulkMarkAsReadHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestBulkDismissHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Request, _ = http.NewRequest("POST", "/notifications/bulk/dismiss", nil)

	bulkDismissHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetPreferencesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockNotifService{
		prefs: []notifications.ChannelPreference{
			{Category: notifications.CategoryUserManagement, EmailEnabled: false, InAppEnabled: true},
		},
	}
	InitNotificationsAPI(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Request, _ = http.NewRequest("GET", "/notification-preferences/", nil)

	getPreferencesHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSetPreferenceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	InitNotificationsAPI(&mockNotifService{})

	body := `{"category":"user_management","email_enabled":false,"inapp_enabled":true}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("currentUser", &types.User{Id: 1})
	c.Request, _ = http.NewRequest("PUT", "/notification-preferences/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	setPreferenceHandler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
```

Update mock to support preferences:

```go
type mockNotifService struct {
	prefs []notifications.ChannelPreference
}

func (m *mockNotifService) GetChannelPreferences(userID int) ([]notifications.ChannelPreference, error) {
	return m.prefs, nil
}
```

**Step 2: Run tests**

Run: `cd sensor_hub && go test ./api/... -run TestNotification -v && go test ./api/... -run TestPreference -v`
Expected: PASS (all tests)

**Step 3: Commit**

```bash
git add sensor_hub/api/notification_api_test.go
git commit -m "test(api): add comprehensive notification API tests"
```

---

### Task 24: Documentation

**Files:**
- Create: `docs/notification-system.md`
- Modify: `docs/alerting-system.md`

**Step 1: Create notification system documentation**

Create `docs/notification-system.md`:

```markdown
# In-App Notification System

## Overview

The notification system provides real-time in-app notifications for system events. Notifications are delivered via WebSocket for connected users and persisted in the database for viewing later.

## Features

- **Real-time delivery** via WebSocket
- **Persistent storage** in MySQL
- **Category-based filtering** (Threshold Alerts, User Management, Configuration Changes)
- **Severity levels** (Info, Warning, Error)
- **Dismissable notifications** with soft-delete (retained in history)
- **Per-user channel preferences** (email vs in-app per category)
- **RBAC integration** for permission-based notification distribution
- **90-day auto-purge** configurable via `notification.retention.days`

## Notification Categories

| Category | Permission Required | Description |
|----------|---------------------|-------------|
| `threshold_alert` | `view_alerts` | Sensor readings exceeding thresholds |
| `user_management` | `view_notifications_user_mgmt` | User additions, removals, role changes |
| `config_change` | `view_notifications_config` | Sensor additions and removals |

## Permissions

| Permission | Description |
|------------|-------------|
| `view_notifications` | Access bell icon and notification page |
| `view_notifications_user_mgmt` | Receive user management notifications |
| `view_notifications_config` | Receive configuration change notifications |
| `manage_notifications` | Configure channel preferences |

## API Endpoints

### Notifications

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| GET | `/notifications/` | `view_notifications` | List notifications (query: limit, offset, include_dismissed) |
| GET | `/notifications/unread-count` | `view_notifications` | Get unread notification count |
| POST | `/notifications/:id/read` | `view_notifications` | Mark notification as read |
| POST | `/notifications/:id/dismiss` | `view_notifications` | Dismiss notification |
| POST | `/notifications/bulk/read` | `view_notifications` | Mark all as read |
| POST | `/notifications/bulk/dismiss` | `view_notifications` | Dismiss all |

### Channel Preferences

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| GET | `/notification-preferences/` | `manage_notifications` | Get user's channel preferences |
| PUT | `/notification-preferences/` | `manage_notifications` | Set channel preference |

## Database Schema

### notifications
Stores notification content:
- `id`, `category`, `severity`, `title`, `message`, `metadata` (JSON), `created_at`

### user_notifications
Tracks per-user notification status:
- `user_id`, `notification_id`, `is_read`, `is_dismissed`, `read_at`, `dismissed_at`

### notification_channel_defaults
System-wide default channel settings per category.

### notification_channel_preferences
Per-user overrides of channel settings.

## UI Components

- **Bell Icon** (TopAppBar): Badge showing unread count
- **Dropdown**: Quick preview of last 5 notifications with mark-read/dismiss actions
- **Notifications Page** (`/notifications`): Full history with filtering, bulk actions
- **Preferences Page** (`/account/notification-preferences`): Configure email/in-app per category

## Configuration

In `application.properties`:
```properties
notification.retention.days=90
```

## WebSocket Integration

Notifications are pushed via WebSocket message:
```json
{
  "type": "notification",
  "notification": {
    "id": 1,
    "notification_id": 1,
    "notification": {
      "category": "user_management",
      "severity": "info",
      "title": "User added",
      "message": "User 'john' was added"
    }
  }
}
```
```

**Step 2: Update alerting-system.md**

Add to "Future Enhancements" section in `docs/alerting-system.md`:

```markdown
- ~~Alert acknowledgment system~~ ✅ **Implemented** - See [notification-system.md](notification-system.md) for in-app notifications
```

**Step 3: Commit**

```bash
git add docs/notification-system.md docs/alerting-system.md
git commit -m "docs: add notification system documentation"
```

---

## Final Integration Commit

After all tasks are complete:

```bash
git add .
git commit -m "feat: complete in-app notifications feature

- Database: notifications, user_notifications, channel preferences tables
- Permissions: view_notifications, view_notifications_user_mgmt, view_notifications_config, manage_notifications
- Backend: NotificationService, NotificationRepository with full CRUD
- Integration: User management, sensor config, and AlertService create notifications
- WebSocket: Real-time push to connected users
- Cleanup: 90-day auto-purge via cleanup service
- UI: Bell icon with badge, dropdown preview, full notifications page
- Preferences: Per-user channel configuration (email vs in-app)
- Tests: Repository, service, and API test coverage
- Docs: notification-system.md documentation"
```

---

**Plan complete and saved to `docs/plans/2026-01-17-in-app-notifications.md`. Ready to execute.**

