# In-App Notifications System

## Overview

The notifications system provides real-time, persistent notifications to users based on system events. It supports multiple notification channels (email, in-app), per-user channel preferences, and role-based access control for viewing notifications. All notification types (threshold alerts, user management, config changes) are delivered through a unified system that respects individual user preferences.

## Features

- **Real-time delivery** via WebSocket push notifications
- **Persistent storage** in database for notification history
- **Per-user channel preferences** - each user controls which notifications they receive via email vs in-app
- **Email notifications** sent to individual user email addresses based on their preferences
- **Notification categories**: threshold_alert, user_management, config_change
- **Severity levels**: info, warning, error
- **RBAC integration** with view_notifications, manage_notifications permissions
- **Soft dismiss** with 90-day auto-purge
- **Rate limiting** for threshold alerts

## Architecture

### Database Schema

#### Tables

1. **notifications** - Stores notification content
   - `id`, `category`, `severity`, `title`, `message`, `metadata`, `created_at`

2. **user_notifications** - Links notifications to users
   - `id`, `user_id`, `notification_id`, `is_read`, `is_dismissed`, `read_at`, `dismissed_at`

3. **notification_channel_defaults** - Global default preferences
   - `category`, `email_enabled`, `inapp_enabled`

4. **user_notification_channel_preferences** - Per-user overrides
   - `user_id`, `category`, `email_enabled`, `inapp_enabled`

### Permissions

| Permission | Description |
|------------|-------------|
| `view_notifications` | View and mark as read own notifications |
| `manage_notifications` | Dismiss notifications and configure preferences |
| `view_notifications_user_mgmt` | Receive user management notifications |
| `view_notifications_config` | Receive configuration change notifications |

Threshold alerts use the existing `view_alerts` permission.

## API Endpoints

### Notifications

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| GET | `/notifications/` | view_notifications | List notifications (supports pagination) |
| GET | `/notifications/unread-count` | view_notifications | Get unread count |
| POST | `/notifications/:id/read` | view_notifications | Mark as read |
| POST | `/notifications/:id/dismiss` | manage_notifications | Dismiss notification |
| POST | `/notifications/bulk/read` | view_notifications | Mark all as read |
| POST | `/notifications/bulk/dismiss` | manage_notifications | Dismiss all |
| GET | `/notifications/ws` | view_notifications | WebSocket for real-time updates |

### Channel Preferences

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| GET | `/notifications/preferences` | view_notifications | Get channel preferences |
| POST | `/notifications/preferences` | manage_notifications | Set channel preference |

## WebSocket Integration

Clients connect to `/notifications/ws` after authentication. Each user subscribes to their personal topic: `notifications:user:<user_id>`.

New notifications are pushed in real-time as JSON:

```json
{
  "id": 123,
  "category": "user_management",
  "severity": "info",
  "title": "User Added",
  "message": "User john was added by admin",
  "metadata": {"username": "john", "actor": "admin"},
  "created_at": "2024-01-17T10:30:00Z"
}
```

## Email Notifications

All notification types support email delivery based on per-user preferences. Emails are sent to individual users who have:

1. A valid email address in their user profile
2. Email notifications enabled for that category in their preferences
3. The appropriate permission to receive that notification type

### Email Flow

1. **Notification created** → `NotificationService.CreateNotification()`
2. **In-app delivery** → Stored in database + pushed via WebSocket
3. **Email delivery** → For each user with permission:
   - Check if user has a valid email address
   - Check user's email preference for the notification category
   - If enabled, send email via SMTP

### SMTP Configuration

Email sending requires OAuth2 configuration for Gmail:

1. Set up OAuth credentials in Google Cloud Console
2. Configure `credentials.json` and authorize to get `token.json`
3. Configure `smtp.user` in `application.properties`

**Note:** The `smtp.recipient` configuration is deprecated. Emails are now sent to individual user email addresses based on their preferences.

### Per-User Email Preferences

Each user can configure their email preferences via the UI at `/notifications/preferences` or via the API. Preferences are stored per-category:

| Category | Default Email | Default In-App |
|----------|---------------|----------------|
| threshold_alert | ✅ Enabled | ✅ Enabled |
| user_management | ❌ Disabled | ✅ Enabled |
| config_change | ❌ Disabled | ✅ Enabled |

Users can override defaults via their notification preferences.

## Notification Categories

### Threshold Alerts (`threshold_alert`)
- Triggered when sensor readings exceed configured thresholds
- Assigned to users with `view_alerts` permission
- Severity: warning

### User Management (`user_management`)
- User created, deleted, role changed
- Assigned to users with `view_notifications_user_mgmt` permission
- Severity: info

### Configuration Changes (`config_change`)
- Sensor added, updated, removed
- Assigned to users with `view_notifications_config` permission
- Severity: info

## Frontend Components

### NotificationBell
- Located in app bar
- Shows unread count badge
- Dropdown with last 5 notifications
- "View all" link to notifications page

### NotificationsPage (`/notifications`)
- Full notification history
- Tabs: Unread / All
- Bulk actions: Mark All Read, Dismiss All
- Individual actions via context menu

### NotificationPreferencesPage (`/notifications/preferences`)
- Toggle email/in-app per category
- Auto-saves on change

## Integration Points

### UserService
- `CreateUser()` → user_management notification
- `DeleteUser()` → user_management notification
- `SetUserRoles()` → user_management notification

### SensorService
- `ServiceAddSensor()` → config_change notification
- `ServiceUpdateSensorById()` → config_change notification
- `ServiceDeleteSensorByName()` → config_change notification

### AlertService
- `ProcessReadingAlert()` → threshold_alert notification (via callback)

## Cleanup

Old notifications are automatically purged after 90 days by the cleanup service.

## Configuration

No additional configuration required. The notification system uses the same rate limiting as email notifications.
