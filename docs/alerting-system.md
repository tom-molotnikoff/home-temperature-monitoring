# Generic Sensor Alerting System

## Overview

The alerting system provides generic, database-driven notifications for sensor readings that fall outside configured thresholds. The system is designed to support multiple sensor types (temperature, humidity, door status, etc.) without requiring code changes for each new sensor type.

## Architecture

### Components

1. **alerting/types.go** - Domain model for alert rules
   - `AlertRule` struct with validation and evaluation logic
   - Supports two alert types: `numeric_range` and `status_based`
   - Built-in rate limiting logic

2. **alerting/service.go** - Business logic orchestrator
   - `AlertService` processes sensor readings against alert rules
   - Coordinates repository lookups, rate limiting checks, and notifications
   - Non-blocking (uses goroutines for alert processing)

3. **db/alert_repository.go** - Database access layer
   - Fetches alert rules with last alert timestamp (for rate limiting)
   - Records alert history
   - Supports efficient LEFT JOIN queries

4. **smtp/smtp.go** - Generic notification system
   - `SMTPNotifier` sends formatted alert emails
   - Supports both numeric and status-based sensors
   - Uses OAuth2 for Gmail authentication

### Database Schema

#### sensor_alert_rules
Stores per-sensor alert configuration:
```sql
CREATE TABLE sensor_alert_rules (
    sensor_id INT PRIMARY KEY,
    alert_type VARCHAR(20) NOT NULL,
    high_threshold DECIMAL(10, 4),
    low_threshold DECIMAL(10, 4),
    status_trigger VARCHAR(100),
    rate_limit_hours INT NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (sensor_id) REFERENCES sensor(id) ON DELETE CASCADE
);
```

#### alert_sent_history
Tracks sent alerts for rate limiting:
```sql
CREATE TABLE alert_sent_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sensor_id INT NOT NULL,
    alert_type VARCHAR(20) NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reading_value VARCHAR(100),
    FOREIGN KEY (sensor_id) REFERENCES sensor(id) ON DELETE CASCADE
);
```

## API Endpoints

The alert management API provides RESTful endpoints for managing sensor alert rules with proper RBAC permissions.

### Alert Rule Management

#### GET /alerts
List all alert rules for all sensors.

**Permissions:** `view_alerts`

**Response:**
```json
[
  {
    "ID": 0,
    "SensorID": 1,
    "SensorName": "Bedroom",
    "AlertType": "numeric_range",
    "HighThreshold": 25.0,
    "LowThreshold": 12.0,
    "TriggerStatus": "",
    "Enabled": true,
    "RateLimitHours": 1,
    "LastAlertSentAt": "2026-01-13T10:00:00Z"
  }
]
```

#### GET /alerts/:sensorId
Get alert rule for a specific sensor by sensor ID.

**Permissions:** `view_alerts`

**Response:** Single AlertRule object (404 if not found)

#### POST /alerts
Create a new alert rule.

**Permissions:** `manage_alerts`

**Request Body (Numeric Range):**
```json
{
  "SensorID": 1,
  "AlertType": "numeric_range",
  "HighThreshold": 30.0,
  "LowThreshold": 10.0,
  "RateLimitHours": 1,
  "Enabled": true
}
```

**Request Body (Status-Based):**
```json
{
  "SensorID": 2,
  "AlertType": "status_based",
  "TriggerStatus": "open",
  "RateLimitHours": 1,
  "Enabled": true
}
```

**Validation:**
- `numeric_range`: HighThreshold must be > LowThreshold
- `status_based`: TriggerStatus must not be empty

**Response:** 201 Created with confirmation message

#### PUT /alerts/:sensorId
Update an existing alert rule. The sensor ID in the URL overrides any sensor ID in the request body.

**Permissions:** `manage_alerts`

**Request Body:** Same structure as POST

**Response:** 200 OK with confirmation message

#### DELETE /alerts/:sensorId
Delete an alert rule for a sensor.

**Permissions:** `manage_alerts`

**Response:** 200 OK with confirmation message

#### GET /alerts/:sensorId/history
Get alert history for a sensor.

**Permissions:** `view_alerts`

**Query Parameters:**
- `limit` (optional): Number of entries to return (1-100, default 50)

**Response:**
```json
[
  {
    "id": 1,
    "sensor_id": 1,
    "alert_type": "numeric_range",
    "reading_value": "35.5",
    "sent_at": "2026-01-13T10:00:00Z"
  }
]
```

### RBAC Permissions

Two permissions control access to alert management:

- **view_alerts**: View alert rules and history (read-only access)
- **manage_alerts**: Create, update, and delete alert rules (full CRUD access)

Both permissions are automatically granted to the `admin` role by the V15 database migration.

To grant permissions to other roles, use the roles API endpoints (requires `manage_roles` permission).

### Web UI

A web interface for managing alert rules is available at `/alerts` in the UI application.

**Features:**
- **View Alert Rules**: DataGrid showing all alert rules with sensor name, type, thresholds, rate limits, enabled status, and last alert timestamp (desktop). On mobile, displays a card-based list for better touch interaction.
- **Create Alert Rules**: Dynamic form that adapts based on alert type (numeric_range vs status_based)
  - Only shows sensors that don't already have alert rules
  - Validates input before submission
- **Edit Alert Rules**: Modify existing rules (sensor cannot be changed, only rule configuration)
- **Delete Alert Rules**: Remove alert rules with confirmation dialog
- **View Alert History**: See the 50 most recent alerts sent for each sensor

**Mobile Support:**
- On screens below 950px, the DataGrid is replaced with a card-based list
- Each card shows: Sensor name, Enabled/Disabled status chip, threshold range or trigger status
- Tapping a card opens the same context menu (Edit, Delete, View History) as the desktop DataGrid row click

**Permissions:**
- Users with `view_alerts` can view alert rules and history
- Users with `manage_alerts` can create, edit, and delete alert rules
- Create, Edit, and Delete buttons are disabled for users without `manage_alerts`

**Access:**
Navigate to the "Alerts" menu item in the sidebar (requires `view_alerts` permission).

## Usage

### Setting Up Alert Rules

Alert rules are configured per-sensor in the database. The V14 migration (`V14__add_sensor_alert_config.sql`) migrates existing temperature thresholds from `application.properties`.

#### Numeric Range Alerts (Temperature, Humidity, Pressure)
```sql
INSERT INTO sensor_alert_rules (sensor_id, alert_type, high_threshold, low_threshold, rate_limit_hours, enabled)
VALUES (1, 'numeric_range', 25.0, 12.0, 1, TRUE);
```

#### Status-Based Alerts (Door, Motion)
```sql
INSERT INTO sensor_alert_rules (sensor_id, alert_type, status_trigger, rate_limit_hours, enabled)
VALUES (2, 'status_based', 'open', 1, TRUE);
```

### How Alerts Are Triggered

1. **Sensor Collection** - `SensorService.ServiceCollectAndStoreTemperatureReadings()` collects readings
2. **Per-Reading Processing** - For each reading, `AlertService.ProcessReadingAlert()` is called in a goroutine
3. **Rule Evaluation** - Alert rule is fetched and evaluated against the reading
4. **Rate Limiting** - System checks if an alert was sent recently (within `rate_limit_hours`)
5. **Notification** - If conditions are met, email is sent via `SMTPNotifier`
6. **History Recording** - Alert is logged in `alert_sent_history` table

### Rate Limiting

Alerts are rate-limited per sensor to prevent notification spam. The `rate_limit_hours` field controls the minimum time between alerts for each sensor.

**Example**: If `rate_limit_hours = 1`, and an alert is sent at 10:00 AM, no further alerts for that sensor will be sent until after 11:00 AM, even if readings continue to exceed thresholds.

## Adding Support for New Sensor Types

The system is designed to handle new sensor types without code changes:

1. **Numeric Sensors (e.g., Humidity, Pressure)**
   - Create sensor entry in `sensor` table with appropriate `type`
   - Add alert rule with `alert_type = 'numeric_range'`
   - Set `high_threshold` and `low_threshold` values
   - No code changes required

2. **Status-Based Sensors (e.g., Door, Motion)**
   - Create sensor entry in `sensor` table with appropriate `type`
   - Add alert rule with `alert_type = 'status_based'`
   - Set `status_trigger` to the status that should trigger alerts (e.g., 'open', 'motion_detected')
   - No code changes required

3. **New Collection Service**
   - When implementing collection for a new sensor type, call `AlertService.ProcessReadingAlert()` after storing each reading
   - Follow the pattern in `sensorService.go:193-220`

## Testing

### Unit Tests
- `alerting/types_test.go` - Domain model validation (12 tests, 100% coverage)
- `alerting/service_test.go` - Business logic with mocks (6 tests)
- `db/alert_repository_test.go` - Repository with mocks (2 tests)
- `smtp/smtp_test.go` - Notifier with test server (6 tests, 94.1% coverage)

### Running Tests
```bash
cd sensor_hub
go test ./alerting/... -v
go test ./smtp/... -v
```

## Configuration

### SMTP Settings
Email configuration is still managed via `application.properties`:
```
smtp.user=your-email@gmail.com
smtp.recipient=alert-recipient@example.com
```

OAuth2 credentials are configured in the `oauth` package.

### Deprecated Configuration
The following fields in `application.properties` are **deprecated** and no longer used:
- `email.alert.high.temperature.threshold` - Use database `sensor_alert_rules` table
- `email.alert.low.temperature.threshold` - Use database `sensor_alert_rules` table

These fields are retained for backwards compatibility with existing config files.

## Migration from Old System

The V14 migration automatically:
1. Creates `sensor_alert_rules` and `alert_sent_history` tables
2. Migrates temperature thresholds from `application.properties` to database
3. Applies rules to all temperature sensors

### Manual Migration Steps
1. Apply the V14 migration using Flyway (see `__notes__/flyway.md`)
2. Verify alert rules: `SELECT * FROM sensor_alert_rules;`
3. Update or add rules as needed for your sensors
4. Monitor `alert_sent_history` table to verify alerts are being sent

## Troubleshooting

### Alerts Not Sending
1. Check if rule exists: `SELECT * FROM sensor_alert_rules WHERE sensor_id = X;`
2. Check if rule is enabled: `enabled = TRUE`
3. Check rate limiting: `SELECT * FROM alert_sent_history WHERE sensor_id = X ORDER BY sent_at DESC LIMIT 1;`
4. Check SMTP configuration in `application.properties`
5. Review application logs for error messages

### Too Many Alerts
- Increase `rate_limit_hours` in the alert rule
- Adjust `high_threshold` and `low_threshold` to reduce sensitivity

### Testing Email Configuration
Use the existing SMTP tests:
```bash
cd sensor_hub
go test ./smtp/... -v -run TestSendAlert
```

## Future Enhancements

Possible improvements for future iterations:
- ~~Web UI for managing alert rules~~ ✅ **Implemented** - See /alerts page in the UI
- Multiple notification channels (SMS, Slack, etc.)
- Alert templates per sensor type
- Configurable alert messages
- Alert acknowledgment system
- Grouped/batched alerts for multiple sensors
