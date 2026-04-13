---
id: alerts-and-notifications
title: Alerts and Notifications
sidebar_position: 7
---

# Alerts and Notifications

Sensor Hub provides an alerting system that monitors sensor readings against configurable rules and delivers notifications through in-app and email channels.

## Alert rules

Alert rules are configured per sensor. Each sensor can have multiple alert rules — one per measurement type per alert type. For example, a sensor that reports both temperature and humidity can have separate threshold rules for each.

There are two types of alert rules:

- Numeric range: triggers when a reading exceeds a high threshold or falls below a low threshold. Use this for temperature alerts (e.g., alert when temperature drops below 15 degrees or exceeds 30 degrees).
- Status-based: triggers when a sensor reports a specific status value.

Each alert rule includes:

- The sensor it applies to
- The measurement type it monitors (e.g. temperature, humidity, battery_low)
- The alert type (numeric range or status-based)
- Threshold values (high and/or low for numeric range, or a trigger status for status-based)
- A rate limit (configurable in seconds, minutes, or hours), which prevents the same alert from firing repeatedly within the specified window
- An enabled/disabled flag

## How alerts are evaluated

When Sensor Hub collects a new reading from a sensor:

1. The reading is checked against the sensor's alert rules for the corresponding measurement type, if any exist
2. If the reading meets the alert condition (e.g., temperature exceeds the high threshold), the system checks the rate limit
3. If sufficient time has elapsed since the last alert for that sensor, the alert fires
4. The alert is recorded in the alert history and a notification is created

## Notifications

Notifications are the delivery mechanism for alerts and system events. They are delivered through two channels:

- In-app: notifications appear in the notification bell in the UI header. New notifications are pushed to connected clients in real time via WebSocket.
- Email: notifications are sent to the user's email address via Gmail SMTP. This requires OAuth 2.0 configuration (see [Installation](installation#set-up-oauth-for-email-notifications-optional)).

## Notification categories

Sensor Hub generates notifications for three categories of events:

| Category              | Description                                            | Default channels  |
|-----------------------|--------------------------------------------------------|-------------------|
| Threshold alerts      | Sensor readings that exceed configured thresholds      | In-app and email  |
| User management       | User creation, deletion, and role changes              | In-app only       |
| Configuration changes | Sensors added, updated, or removed; properties changed | In-app only       |

## Notification preferences

Each user can configure which categories of notifications they receive and through which channels. Preferences are managed on the Alerts and Notifications page in the web UI.

For each category, you can independently enable or disable:

- In-app notifications
- Email notifications

Email notifications require a working OAuth configuration. If OAuth is not configured, email delivery is silently skipped.

## Notification lifecycle

- New notifications appear as unread in the notification bell
- Notifications can be marked as read individually or in bulk
- Notifications can be dismissed individually or in bulk
- Dismissed notifications are hidden by default but can be included in queries by setting the `include_dismissed` parameter
- Notifications older than 90 days are automatically purged

## Email notification setup

To enable email notifications:

1. Complete the OAuth setup as described in the [Installation guide](installation#set-up-oauth-for-email-notifications-optional)
2. Ensure `smtp.user` is set in `configuration/smtp.properties` to the Gmail address that will send notifications
3. The OAuth token is refreshed automatically in the background (default: every 30 minutes)

Once configured, the OAuth status is displayed on the Alerts and Notifications page under the OAuth Configuration card.

## Permissions

| Permission             | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `view_alerts`          | View alert rules and alert history                           |
| `manage_alerts`        | Create, edit, and delete alert rules                         |
| `view_notifications`   | View notifications and unread count                          |
| `manage_notifications` | Dismiss notifications and configure notification preferences |
| `manage_oauth`         | Configure OAuth settings for email delivery                  |
