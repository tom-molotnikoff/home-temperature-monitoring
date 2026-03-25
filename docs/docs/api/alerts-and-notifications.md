---
id: alerts-and-notifications
title: Alerts and Notifications
sidebar_position: 3
---

# Alerts and Notifications API

Endpoints for managing alert rules, viewing alert history, and managing notifications. All endpoints require authentication and the specified permission.

> All paths below are relative to the `/api` base path (e.g. `GET /alerts` is served at `GET /api/alerts`).

---

## Alert endpoints

### GET /alerts

List all alert rules.

Permission: `view_alerts`

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "sensor_id": 1,
    "alert_type": "numeric_range",
    "high_threshold": 30.0,
    "low_threshold": 15.0,
    "trigger_status": null,
    "enabled": true,
    "rate_limit_hours": 1,
    "last_alert_sent_at": "2026-01-15T12:00:00Z",
    "created_at": "2026-01-10T10:00:00Z",
    "updated_at": "2026-01-10T10:00:00Z"
  }
]
```

---

### GET /alerts/:sensorId

Get the alert rule for a specific sensor.

Permission: `view_alerts`

#### Path parameters

| Parameter   | Type    | Description  |
|-------------|---------|--------------|
| `sensorId`  | integer | Sensor ID    |

#### Response (200 OK)

Returns a single alert rule object in the same format as the list response.

#### Response (404 Not Found)

Returned if no alert rule exists for the sensor.

---

### GET /alerts/:sensorId/history

Get the alert history for a sensor, showing when alerts were triggered.

Permission: `view_alerts`

#### Path parameters

| Parameter  | Type    | Description  |
|------------|---------|--------------|
| `sensorId` | integer | Sensor ID    |

#### Query parameters

| Parameter  | Type    | Required  | Description                                       |
|------------|---------|-----------|---------------------------------------------------|
| `limit`    | integer | no        | Maximum records to return (default: 50, max: 100) |

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "sensor_id": 1,
    "alert_type": "numeric_range",
    "reading_value": "31.5",
    "sent_at": "2026-01-15T12:00:00Z"
  }
]
```

---

### POST /alerts

Create a new alert rule.

Permission: `manage_alerts`

#### Request body

```json
{
  "sensor_id": 1,
  "alert_type": "numeric_range",
  "high_threshold": 30.0,
  "low_threshold": 15.0,
  "enabled": true,
  "rate_limit_hours": 1
}
```

For status-based alerts, use `alert_type: "status_based"` and include the `trigger_status` field instead of thresholds.

#### Response (201 Created)

```json
{
  "message": "Alert rule created successfully"
}
```

---

### PUT /alerts/:sensorId

Update an existing alert rule.

Permission: `manage_alerts`

#### Path parameters

| Parameter  | Type    | Description  |
|------------|---------|--------------|
| `sensorId` | integer | Sensor ID    |

#### Request body

Same format as the create request.

#### Response (200 OK)

```json
{
  "message": "Alert rule updated successfully"
}
```

---

### DELETE /alerts/:sensorId

Delete an alert rule.

Permission: `manage_alerts`

#### Path parameters

| Parameter   | Type    | Description   |
|-------------|---------|---------------|
| `sensorId`  | integer | Sensor ID     |

#### Response (200 OK)

```json
{
  "message": "Alert rule deleted successfully"
}
```

---

## Notification endpoints

### GET /notifications

List notifications for the current user.

Permission: `view_notifications`

#### Query parameters

| Parameter           | Type    | Required  | Description                                      |
|---------------------|---------|-----------|--------------------------------------------------|
| `limit`             | integer | no        | Maximum records to return (default: 50)          |
| `offset`            | integer | no        | Number of records to skip (default: 0)           |
| `include_dismissed` | boolean | no        | Include dismissed notifications (default: false) |

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "user_id": 1,
    "notification_id": 10,
    "is_read": false,
    "is_dismissed": false,
    "read_at": null,
    "dismissed_at": null,
    "notification": {
      "id": 10,
      "category": "threshold_alert",
      "severity": "warning",
      "title": "High Temperature Alert",
      "message": "Downstairs sensor reading 31.5°C exceeds high threshold of 30.0°C",
      "metadata": {
        "sensor_id": 1,
        "reading_value": 31.5,
        "threshold": 30.0
      },
      "created_at": "2026-01-15T12:00:00Z"
    }
  }
]
```

---

### GET /notifications/unread-count

Get the count of unread notifications for the current user.

Permission: `view_notifications`

#### Response (200 OK)

```json
{
  "count": 3
}
```

---

### POST /notifications/:id/read

Mark a notification as read.

Permission: `view_notifications`

#### Path parameters

| Parameter  | Type    | Description     |
|------------|---------|-----------------|
| `id`       | integer | Notification ID |

#### Response (200 OK)

```json
{
  "message": "marked as read"
}
```

---

### POST /notifications/:id/dismiss

Dismiss a notification.

Permission: `manage_notifications`

#### Path parameters

| Parameter  | Type    | Description     |
|------------|---------|-----------------|
| `id`       | integer | Notification ID |

#### Response (200 OK)

```json
{
  "message": "dismissed"
}
```

---

### POST /notifications/bulk/read

Mark all notifications as read for the current user.

Permission: `view_notifications`

#### Response (200 OK)

```json
{
  "message": "all marked as read"
}
```

---

### POST /notifications/bulk/dismiss

Dismiss all notifications for the current user.

Permission: `manage_notifications`

#### Response (200 OK)

```json
{
  "message": "all dismissed"
}
```

---

### GET /notifications/preferences

Get notification channel preferences for the current user.

Permission: `view_notifications`

#### Response (200 OK)

```json
[
  {
    "user_id": 1,
    "category": "threshold_alert",
    "email_enabled": true,
    "inapp_enabled": true
  },
  {
    "user_id": 1,
    "category": "user_management",
    "email_enabled": false,
    "inapp_enabled": true
  },
  {
    "user_id": 1,
    "category": "config_change",
    "email_enabled": false,
    "inapp_enabled": true
  }
]
```

---

### POST /notifications/preferences

Set a notification channel preference.

Permission: `manage_notifications`

#### Request body

```json
{
  "category": "threshold_alert",
  "email_enabled": true,
  "inapp_enabled": true
}
```

Valid categories: `threshold_alert`, `user_management`, `config_change`.

#### Response (200 OK)

```json
{
  "message": "preference saved"
}
```

---

### WebSocket: user notifications

`GET /notifications/ws`

Permission: `view_notifications`

Streams real-time notifications for the authenticated user. New notifications are pushed as they are created.

#### Message format

```json
{
  "id": 1,
  "user_id": 1,
  "notification_id": 10,
  "is_read": false,
  "is_dismissed": false,
  "notification": {
    "id": 10,
    "category": "threshold_alert",
    "severity": "warning",
    "title": "High Temperature Alert",
    "message": "Downstairs sensor reading 31.5°C exceeds high threshold of 30.0°C",
    "created_at": "2026-01-15T12:00:00Z"
  }
}
```
