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

## GET /temperature/readings/between

Get raw temperature readings between two dates (inclusive) for all sensors.

Permission: `view_readings`

### Query parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| `start`   | string | yes      | Start date in `YYYY-MM-DD` format |
| `end`     | string | yes      | End date in `YYYY-MM-DD` format   |

### Example request

```bash
curl "http://localhost:8080/api/temperature/readings/between?start=2026-01-14&end=2026-01-15" \
  -H "X-API-Key: shk_..."
```

### Response (200 OK)

```json
[
  {
    "id": 1234,
    "sensor_name": "Downstairs",
    "time": "2026-01-14 10:00:00",
    "temperature": 21.56
  }
]
```

---

## GET /temperature/readings/hourly/between

Get hourly-averaged temperature readings between two dates (inclusive) for all sensors. Useful for plotting long-term trends.

Permission: `view_readings`

### Query parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| `start`   | string | yes      | Start date in `YYYY-MM-DD` format |
| `end`     | string | yes      | End date in `YYYY-MM-DD` format   |

### Example request

```bash
curl "http://localhost:8080/api/temperature/readings/hourly/between?start=2026-01-14&end=2026-01-15" \
  -H "X-API-Key: shk_..."
```

### Response (200 OK)

```json
[
  {
    "id": 5678,
    "sensor_name": "Downstairs",
    "time": "2026-01-14 12:00:00",
    "temperature": 21.34
  },
  {
    "id": 5679,
    "sensor_name": "Downstairs",
    "time": "2026-01-14 13:00:00",
    "temperature": 21.78
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
