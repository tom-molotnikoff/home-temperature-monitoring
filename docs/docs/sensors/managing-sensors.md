---
id: managing-sensors-ref
title: Managing Sensors
sidebar_position: 4
---

# Managing Sensors

Once sensors are connected and reporting data, Sensor Hub provides tools for managing them — from registration and data collection through to health monitoring, retention policies, and permissions.

## Registering sensors

Sensors can be registered in two ways depending on the driver type:

- **Pull sensors (HTTP):** Register manually through the web UI, the REST API, or the CLI. Each registration requires a name, a sensor driver, and driver-specific configuration (e.g. the sensor's URL).
- **Push sensors (MQTT/Zigbee):** Auto-discovered when a new device publishes an MQTT message. Created with **Pending** status for you to approve or dismiss.

Each driver declares which configuration fields it needs. You can see the available drivers and their config schemas via:

- The web UI **Add Sensor** form (config fields appear dynamically when you select a driver)
- The REST API: `GET /api/drivers`
- The CLI: `sensor-hub drivers list`

For example, the `sensor-hub-http-temperature` driver requires a single `url` field — the base URL of the sensor (e.g. `http://192.168.1.50:5000`).

The sensor's endpoint must be reachable from the Sensor Hub host. If running in Docker, the sensor must be accessible from within the Docker network.

## Data collection

### Pull sensors

Sensor Hub polls each registered pull sensor at a configurable interval. The default is every 300 seconds (5 minutes). This interval is controlled by the `sensor.collection.interval` property (see [Configuration Settings](../configuration)).

Each poll cycle:

1. Sensor Hub invokes the sensor's driver with its config to fetch readings
2. The sensor returns the current reading
3. The reading is stored in the database and broadcast to connected UI clients via WebSocket
4. Hourly averages are computed and stored separately for efficient historical queries
5. If an alert rule exists for the sensor, the reading is evaluated against the rule

### Push sensors

MQTT-based sensors report data as it changes — there is no polling interval. Messages arrive in real time and are processed immediately through the same pipeline (store, broadcast, evaluate alerts).

## Sensor health monitoring

Sensor Hub tracks the health status of each sensor based on whether it responds successfully when polled (pull sensors) or whether messages are arriving (push sensors). Health status changes are recorded and displayed in the UI.

Health history is retained for a configurable period (default: 180 days), controlled by the `health.history.retention.days` property.

## Data retention

Sensor readings are retained for a configurable period (default: 90 days), controlled by the `sensor.data.retention.days` property. A cleanup task runs at a configurable interval (default: every 1 hour) to remove expired data.

### Per-sensor retention

Individual sensors can override the global retention period with a custom value in hours. This is useful when some sensors produce high-volume data that should be kept for a shorter period, or when critical sensors need longer retention.

Per-sensor retention can be configured:

- Through the web UI on the individual Sensor page, using the Data Retention card (requires `manage_sensors` permission).
- Through the Data Retention overview page, which shows all sensors and their retention settings.
- Through the REST API by sending a `PUT` request to `/sensors/:id` with `retention_hours` in the body. Set to a positive integer to override, or `null` to revert to the global default. See the [Sensors and Readings API reference](../api/sensors-and-readings) for details.
- Through the CLI: `sensor-hub sensors update <id> --retention-hours <hours>` or `sensor-hub sensors update <id> --retention-hours null`

When a per-sensor retention is set, it always takes precedence over the global default. The cleanup task processes sensors with custom retention first, then applies the global retention to all remaining sensors.

The effective retention for a sensor can be seen via the `GET /sensors/:name` endpoint, which returns an `effective_retention_hours` field showing the retention that will actually be applied during cleanup.

## Managing sensors in the UI

The Sensors Overview page lists all registered sensors with their current status. From this page you can:

- View the list of all sensors, filtered by driver
- Add new sensors
- Edit sensor details (name, configuration)
- Delete sensors

Individual sensor pages provide detailed views including:

- Current reading and status
- Data charts with configurable date ranges
- Health history charts
- Sensor metadata

## Permissions

| Permission         | Description                                  |
|--------------------|----------------------------------------------|
| `view_sensors`     | View the sensor list and sensor data         |
| `manage_sensors`   | Add, edit, approve, and dismiss sensors      |
| `delete_sensors`   | Delete sensors                               |
| `view_readings`    | View sensor readings and charts              |
| `trigger_readings` | Manually trigger a sensor reading collection |
| `view_mqtt`        | View MQTT brokers and subscriptions          |
| `manage_mqtt`      | Create, update, and delete MQTT brokers and subscriptions |
