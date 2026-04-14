---
id: dashboards
title: Dashboards
sidebar_position: 6
---

# Dashboards API

Endpoints for creating, reading, updating, and deleting customisable widget dashboards. Each dashboard stores a JSON configuration of widgets, their layout positions, and per-widget settings.

> All paths below are relative to the `/api` base path (e.g. `GET /dashboards/` is served at `GET /api/dashboards/`).

---

## Permissions

| Permission | Description |
|---|---|
| `view_dashboards` | List and get dashboards |
| `manage_dashboards` | Create, update, delete, share dashboards |

---

## List dashboards

### GET /dashboards/

Returns all dashboards owned by the authenticated user.

Permission: `view_dashboards`

#### Response (200 OK)

```json
[
  {
    "id": 1,
    "user_id": 1,
    "name": "My Dashboard",
    "config": "{\"widgets\":[],\"breakpoints\":{\"lg\":12,\"md\":10,\"sm\":6}}",
    "shared": false,
    "is_default": false,
    "created_at": "2026-01-15T10:30:00Z",
    "updated_at": "2026-01-15T10:30:00Z"
  }
]
```

Returns an empty array `[]` when no dashboards exist.

---

## Get dashboard

### GET /dashboards/:id

Returns a single dashboard by ID.

Permission: `view_dashboards`

#### Response (200 OK)

```json
{
  "id": 1,
  "user_id": 1,
  "name": "My Dashboard",
  "config": "{\"widgets\":[{\"id\":\"abc123\",\"type\":\"temperature-chart\",\"config\":{},\"layout\":{\"x\":0,\"y\":0,\"w\":6,\"h\":4}}],\"breakpoints\":{\"lg\":12,\"md\":10,\"sm\":6}}",
  "shared": false,
  "is_default": false,
  "created_at": "2026-01-15T10:30:00Z",
  "updated_at": "2026-01-15T12:00:00Z"
}
```

#### Response (404 Not Found)

```json
{ "message": "Dashboard not found" }
```

---

## Create dashboard

### POST /dashboards/

Creates a new dashboard for the authenticated user.

Permission: `manage_dashboards`

#### Request body

```json
{
  "name": "My Dashboard"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Dashboard display name |
| `config` | object | No | Initial widget configuration (defaults to empty) |

#### Response (201 Created)

```json
{ "id": 1 }
```

---

## Update dashboard

### PUT /dashboards/:id

Updates the name and/or widget configuration of a dashboard. Only the dashboard owner can update.

Permission: `manage_dashboards`

#### Request body

```json
{
  "name": "Renamed Dashboard",
  "config": {
    "widgets": [
      {
        "id": "abc123",
        "type": "temperature-chart",
        "config": {},
        "layout": { "x": 0, "y": 0, "w": 6, "h": 4 }
      }
    ],
    "breakpoints": { "lg": 12, "md": 10, "sm": 6 }
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | No | New dashboard name |
| `config` | object | No | Full widget configuration |

#### Response (200 OK)

```json
{ "message": "Dashboard updated" }
```

---

## Delete dashboard

### DELETE /dashboards/:id

Deletes a dashboard. Only the owner can delete.

Permission: `manage_dashboards`

#### Response (200 OK)

```json
{ "message": "Dashboard deleted" }
```

---

## Share dashboard

### POST /dashboards/:id/share

Creates a copy of the dashboard for another user.

Permission: `manage_dashboards`

#### Request body

```json
{ "target_user_id": 2 }
```

#### Response (200 OK)

```json
{ "message": "Dashboard shared" }
```

---

## Set default dashboard

### PUT /dashboards/:id/default

Sets a dashboard as the default for the authenticated user. The previously default dashboard (if any) is unset.

Permission: `manage_dashboards`

#### Response (200 OK)

```json
{ "message": "Default dashboard set" }
```

---

## Dashboard config structure

The `config` field is a JSON string containing the full widget layout. When decoded:

```json
{
  "widgets": [
    {
      "id": "unique-widget-id",
      "type": "temperature-chart",
      "config": {
        "sensorName": "Living Room"
      },
      "layout": {
        "x": 0,
        "y": 0,
        "w": 6,
        "h": 4
      }
    }
  ],
  "breakpoints": {
    "lg": 12,
    "md": 10,
    "sm": 6
  }
}
```

### Available widget types

| Type                 | Description                                                     | Config fields                                                                                       |
|----------------------|-----------------------------------------------------------------|-----------------------------------------------------------------------------------------------------|
| `readings-chart`     | Line chart for any measurement type with configurable date range | `measurementType`, `timeRange`, `refreshInterval` (number, default 30) |
| `live-readings`      | Real-time sensor readings data grid                             | —                                                                                                   |
| `weather-forecast`   | External weather forecast from configured provider              | —                                                                                                   |
| `sensor-health-pie`  | Pie chart showing sensor health status distribution             | —                                                                                                   |
| `sensor-type-pie`    | Pie chart showing sensor type distribution                      | —                                                                                                   |
| `health-timeline`    | Sensor health status history chart                              | `sensorId` (number), `limit` (number, default 1000)                                                 |
| `reading-stats`      | Total readings per sensor data grid                             | —                                                                                                   |
| `notifications-feed` | Recent notifications feed                                       | —                                                                                                   |
| `markdown-note`      | User-defined text block for notes or labels                     | `content` (textarea)                                                                                |
| `current-reading`    | Current value display for a sensor (numeric or binary/text)     | `sensorId` (number), `measurementType`                                                              |
| `min-max-avg`        | Period statistics (min, max, average) for a sensor              | `sensorId` (number), `measurementType`, `timeRange`                                                 |
| `gauge`              | Visual circular gauge for a single sensor                       | `sensorId` (number), `measurementType`, `min` (number, default 0), `max` (number, default 40)       |
| `comparison-chart`   | Multi-sensor overlay line chart for any measurement type        | `measurementType`, `sensorIds` (number[]), `timeRange`, `refreshInterval` (number, default 30) |
| `group-summary`      | Average reading for a measurement type across all sensors       | `measurementType`                                                                                   |
| `alert-summary`      | Compact list of configured alert rules                          | —                                                                                                   |
| `uptime`             | Uptime percentage for a sensor                                  | `sensorId` (number), `limit` (number, default 1000)                                                 |
| `heatmap`            | Colour-coded 30-day grid for any measurement type               | `sensorId` (number), `measurementType`, `scaleMin` (number, default 10), `scaleMax` (number, default 30) |
| `sensor-detail`      | Latest readings grid for all measurement types of a sensor      | `sensorId` (number)                                                                                 |

> **Note:** `temperature-chart` is an alias for `readings-chart` (kept for backward compatibility).
