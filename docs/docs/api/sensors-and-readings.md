---
id: sensors-and-readings
title: Sensors and Readings
sidebar_position: 2
---

# Sensors and Readings API

Endpoints for managing sensors and retrieving sensor data. All endpoints require authentication and the specified permission.

> All paths below are relative to the `/api` base path (e.g. `GET /sensors` is served at `GET /api/sensors`).

---

## GET /drivers

List all available sensor drivers and their config field schemas.

Permission: any authenticated user

### Response (200 OK)

```json
[
  {
    "type": "sensor-hub-http-temperature",
    "display_name": "Sensor Hub HTTP Temperature",
    "description": "Reads temperature from a Sensor Hub HTTP endpoint.",
    "supported_measurement_types": ["temperature"],
    "config_fields": [
      {
        "key": "url",
        "label": "Sensor URL",
        "description": "Base URL of the HTTP sensor (e.g. http://192.168.1.50:8080)",
        "required": true,
        "sensitive": false
      }
    ]
  }
]
```

---

## GET /sensors

List all registered sensors.

Permission: `view_sensors`

### Response (200 OK)

```json
[
  {
    "id": 1,
    "name": "Downstairs",
    "sensor_driver": "sensor-hub-http-temperature",
    "config": {
      "url": "http://192.168.1.50:5000"
    },
    "health_status": "good",
    "enabled": true
  }
]
```

---

## GET /sensors/driver/:driver

List sensors filtered by driver.

Permission: `view_sensors`

### Path parameters

| Parameter  | Type   | Description                                                        |
|------------|--------|--------------------------------------------------------------------|
| `driver`   | string | Sensor driver to filter by (e.g., `sensor-hub-http-temperature`)   |

### Response (200 OK)

Same format as `GET /sensors`, filtered to the specified driver.

---

## POST /sensors

Register a new sensor.

Permission: `manage_sensors`

### Request body

```json
{
  "name": "Upstairs",
  "sensor_driver": "sensor-hub-http-temperature",
  "config": {
    "url": "http://192.168.1.51:5000"
  }
}
```

### Response (201 Created)

```json
{
  "id": 2
}
```

---

## PUT /sensors/:id

Update an existing sensor.

Permission: `manage_sensors`

### Path parameters

| Parameter  | Type    | Description  |
|------------|---------|--------------|
| `id`       | integer | Sensor ID    |

### Request body

```json
{
  "name": "Upstairs Bedroom",
  "config": {
    "url": "http://192.168.1.51:5000"
  }
}
```

### Response (200 OK)

```json
{
  "message": "sensor updated"
}
```

---

## DELETE /sensors/:id

Delete a sensor and its associated data.

Permission: `delete_sensors`

### Path parameters

| Parameter  | Type    | Description |
|------------|---------|-------------|
| `id`       | integer | Sensor ID   |

### Response (200 OK)

```json
{
  "message": "sensor deleted"
}
```

---

## GET /readings/between

Get sensor readings between two dates/times for all sensors. The server automatically aggregates readings based on the requested time span to keep response sizes manageable for charting. Short ranges return raw data; longer ranges return time-bucketed aggregates. All timestamps are stored in UTC.

Permission: `view_readings`

### Query parameters

| Parameter              | Type   | Required | Description                                                                 |
|------------------------|--------|----------|-----------------------------------------------------------------------------|
| `start`                | string | yes      | Start date (`YYYY-MM-DD`, start of day) or ISO 8601 datetime (e.g. `2026-01-14T10:00:00Z`) |
| `end`                  | string | yes      | End date (`YYYY-MM-DD`, end of day) or ISO 8601 datetime (e.g. `2026-01-14T22:00:00Z`)     |
| `type`                 | string | no       | Filter by measurement type (e.g., `temperature`)                            |
| `sensor`               | string | no       | Filter by sensor name                                                       |
| `aggregation`          | string | no       | Override the auto-selected aggregation interval (ISO 8601 duration, e.g. `PT1H`, `PT5M`). Use `raw` for no aggregation. |
| `aggregation_function` | string | no       | Override the aggregation function (`avg`, `min`, `max`, `sum`, `count`, `last`). Defaults to the measurement type's configured function. |

### Auto-aggregation

The server selects an aggregation interval based on the time span of the query:

| Time span        | Aggregation interval |
|------------------|---------------------|
| ≤ 15 minutes     | raw (no aggregation) |
| ≤ 1 hour         | 10 seconds          |
| ≤ 6 hours        | 1 minute            |
| ≤ 1 day          | 5 minutes           |
| ≤ 7 days         | 15 minutes          |
| ≤ 30 days        | 1 hour              |
| > 30 days        | 1 day               |

These tiers are configurable via `application.properties`. Aggregation can be disabled entirely, in which case all queries return raw readings.

Each measurement type has a default aggregation function (e.g., `avg` for temperature, `last` for binary states). These defaults are stored in the `measurement_type_aggregations` database table and can be overridden per-request via the `aggregation_function` query parameter.

### Example requests

```bash
# Short range — returns raw readings automatically
curl "http://localhost:8080/api/readings/between?start=2026-01-14T10:00:00Z&end=2026-01-14T10:10:00Z" \
  -H "X-API-Key: shk_..."

# Longer range — automatically aggregated
curl "http://localhost:8080/api/readings/between?start=2026-01-14&end=2026-01-15" \
  -H "X-API-Key: shk_..."

# Override aggregation interval
curl "http://localhost:8080/api/readings/between?start=2026-01-01&end=2026-01-31&aggregation=PT1H" \
  -H "X-API-Key: shk_..."

# Force raw readings (no aggregation)
curl "http://localhost:8080/api/readings/between?start=2026-01-14&end=2026-01-15&aggregation=raw" \
  -H "X-API-Key: shk_..."
```

### Response (200 OK)

The response includes metadata about the aggregation applied, plus the array of readings:

```json
{
  "aggregation_interval": "PT5M",
  "aggregation_function": "avg",
  "readings": [
    {
      "id": 1234,
      "sensor_name": "Downstairs",
      "measurement_type": "temperature",
      "numeric_value": 21.56,
      "text_state": null,
      "unit": "°C",
      "time": "2026-01-14 10:00:00"
    }
  ]
}
```

When no aggregation is applied (short range or `aggregation=raw`), the response uses `"aggregation_interval": "raw"` and `"aggregation_function": "none"`.

---

## GET /measurement-types

Get all known measurement types.

Permission: `view_sensors`

### Query parameters

| Parameter      | Type    | Required | Description                                                      |
|----------------|---------|----------|------------------------------------------------------------------|
| `has_readings` | boolean | no       | When `true`, only return types that have at least one stored reading |

### Response (200 OK)

```json
[
  {
    "name": "temperature",
    "display_name": "Temperature",
    "category": "environment",
    "unit": "°C"
  },
  {
    "name": "humidity",
    "display_name": "Humidity",
    "category": "environment",
    "unit": "%"
  }
]
```

---

## GET /sensors/:id/health-history

Get the health status history for a sensor.

Permission: `view_sensors`

### Path parameters

| Parameter | Type    | Description |
|-----------|---------|-------------|
| `id`      | integer | Sensor ID   |

### Query parameters

| Parameter | Type    | Required | Description                                         |
|-----------|---------|----------|-----------------------------------------------------|
| `limit`   | integer | no       | Maximum number of records to return (default: 5000) |

### Response (200 OK)

```json
[
  {
    "sensor_id": 1,
    "status": "healthy",
    "checked_at": "2026-01-15T14:30:00Z"
  },
  {
    "sensor_id": 1,
    "status": "unhealthy",
    "checked_at": "2026-01-15T14:25:00Z"
  }
]
```

---

## WebSocket: current readings

`GET /readings/ws/current`

Permission: `view_readings`

Streams real-time sensor reading updates as they are collected. Messages are pushed each time a collection cycle completes.

### Message format

```json
[
  {
    "sensor_id": 1,
    "sensor_name": "Downstairs",
    "measurement_type": "temperature",
    "numeric_value": 21.56,
    "text_state": null,
    "unit": "°C",
    "time": "2026-01-15T14:30:00Z"
  }
]
```

---

## WebSocket: sensor metadata

`GET /sensors/ws/:driver`

Permission: `view_sensors`

Streams sensor metadata changes (additions, removals, status changes) for the specified sensor driver.

### Path parameters

| Parameter | Type   | Description                                                      |
|-----------|--------|------------------------------------------------------------------|
| `driver`  | string | Sensor driver (e.g., `sensor-hub-http-temperature`)              |

### Message format

```json
[
  {
    "id": 1,
    "name": "Downstairs",
    "sensor_driver": "sensor-hub-http-temperature",
    "config": {
      "url": "http://192.168.1.50:5000"
    },
    "health_status": "good"
  }
]
```
