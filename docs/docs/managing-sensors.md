---
id: managing-sensors
title: Managing Sensors
sidebar_position: 6
---

# Managing Sensors

Once sensors are deployed and accessible on your network, they need to be registered in Sensor Hub so that readings are collected and stored.

## Registering sensors

Sensors can be registered in two ways:

- Through the web UI on the Sensors Overview page, using the add sensor form (requires `manage_sensors` permission).
- Through the REST API by sending a `POST` request to `/sensors`. See the [Sensors and Readings API reference](api/sensors-and-readings) for details.

Each sensor registration requires:

- A name to identify the sensor (e.g., "Downstairs" or "Upstairs bedroom")
- A sensor driver, such as `sensor-hub-http-temperature`
- The URL of the sensor's HTTP endpoint (e.g., `http://192.168.1.50:5000`)

The URL must be reachable from the Sensor Hub host. If running in Docker, the sensor must be accessible from within the Docker network.

## Sensor drivers

The primary sensor driver is `sensor-hub-http-temperature`, which covers DS18B20 and compatible sensors that return temperature readings in Celsius. Additional sensor drivers can be supported by extending the driver definitions in the backend.

## Data collection

Sensor Hub polls each registered sensor at a configurable interval. The default is every 300 seconds (5 minutes). This interval is controlled by the `sensor.collection.interval` property (see [Configuration Settings](configuration)).

Each poll cycle:

1. Sensor Hub sends a `GET` request to the sensor's URL
2. The sensor returns the current reading as JSON
3. The reading is stored in the database and broadcast to connected UI clients via WebSocket
4. Hourly averages are computed and stored separately for efficient historical queries
5. If an alert rule exists for the sensor, the reading is evaluated against the rule

## Sensor health monitoring

Sensor Hub tracks the health status of each sensor based on whether it responds successfully when polled. Health status changes are recorded and displayed in the UI.

Health history is retained for a configurable period (default: 180 days), controlled by the `health.history.retention.days` property.

## Data retention

Temperature readings are retained for a configurable period (default: 365 days), controlled by the `sensor.data.retention.days` property. A cleanup task runs at a configurable interval (default: every 24 hours) to remove expired data.

## Managing sensors in the UI

The Sensors Overview page lists all registered sensors with their current status. From this page you can:

- View the list of all sensors, filtered by driver
- Add new sensors
- Edit sensor details (name, URL)
- Delete sensors

Individual sensor pages provide detailed views including:

- Current reading and status
- Temperature data charts with configurable date ranges
- Health history charts
- Sensor metadata

## Permissions

| Permission         | Description                                  |
|--------------------|----------------------------------------------|
| `view_sensors`     | View the sensor list and sensor data         |
| `manage_sensors`   | Add and edit sensors                         |
| `delete_sensors`   | Delete sensors                               |
| `view_readings`    | View sensor readings and charts               |
| `trigger_readings` | Manually trigger a sensor reading collection |
