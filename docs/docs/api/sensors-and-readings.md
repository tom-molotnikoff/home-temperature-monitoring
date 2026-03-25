---
id: sensors-and-readings
title: Sensors and Readings
sidebar_position: 2
---

# Sensors and Readings API

Endpoints for managing sensors and retrieving temperature data. All endpoints require authentication and the specified permission.

> All paths below are relative to the `/api` base path (e.g. `GET /sensors` is served at `GET /api/sensors`).

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
    "type": "temperature",
    "url": "http://192.168.1.50:5000",
    "status": "healthy",
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-15T10:00:00Z"
  }
]
```

---

## GET /sensors/:type

List sensors filtered by type.

Permission: `view_sensors`

### Path parameters

| Parameter  | Type   | Description                                    |
|------------|--------|------------------------------------------------|
| `type`     | string | Sensor type to filter by (e.g., `temperature`) |

### Response (200 OK)

Same format as `GET /sensors`, filtered to the specified type.

---

## POST /sensors

Register a new sensor.

Permission: `manage_sensors`

### Request body

```json
{
  "name": "Upstairs",
  "type": "temperature",
  "url": "http://192.168.1.51:5000"
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
  "url": "http://192.168.1.51:5000"
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

## GET /temperature/current

Get the most recent reading from all temperature sensors.

Permission: `view_readings`

### Response (200 OK)

```json
[
  {
    "sensor_id": 1,
    "sensor_name": "Downstairs",
    "temperature": 21.56,
    "time": "2026-01-15T14:30:00Z"
  }
]
```

---

## GET /temperature/hourly-average

Get hourly average temperature data for a sensor within a date range.

Permission: `view_readings`

### Query parameters

| Parameter  | Type    | Required  | Description                   |
|------------|---------|-----------|-------------------------------|
| `sensorId` | integer | yes       | Sensor ID                     |
| `from`     | string  | yes       | Start date in ISO 8601 format |
| `to`       | string  | yes       | End date in ISO 8601 format   |

### Example request

```bash
curl "http://localhost:8080/temperature/hourly-average?sensorId=1&from=2026-01-14T00:00:00Z&to=2026-01-15T00:00:00Z" \
  -b cookies.txt
```

### Response (200 OK)

```json
[
  {
    "sensor_id": 1,
    "hour": "2026-01-14T12:00:00Z",
    "average_temperature": 21.34
  },
  {
    "sensor_id": 1,
    "hour": "2026-01-14T13:00:00Z",
    "average_temperature": 21.78
  }
]
```

---

## POST /temperature/collect

Manually trigger a sensor reading collection cycle. This polls all registered sensors immediately, outside the normal collection interval.

Permission: `trigger_readings`

### Response (200 OK)

```json
{
  "message": "collection triggered"
}
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

## WebSocket: current temperatures

`GET /temperature/ws/current-temperatures`

Permission: `view_readings`

Streams real-time temperature updates as they are collected. Messages are pushed each time a collection cycle completes.

### Message format

```json
[
  {
    "sensor_id": 1,
    "sensor_name": "Downstairs",
    "temperature": 21.56,
    "time": "2026-01-15T14:30:00Z"
  }
]
```

---

## WebSocket: sensor metadata

`GET /sensors/ws/:type`

Permission: `view_sensors`

Streams sensor metadata changes (additions, removals, status changes) for the specified sensor type.

### Path parameters

| Parameter | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `type`    | string | Sensor type (e.g., `temperature`) |

### Message format

```json
[
  {
    "id": 1,
    "name": "Downstairs",
    "type": "temperature",
    "url": "http://192.168.1.50:5000",
    "status": "healthy"
  }
]
```
