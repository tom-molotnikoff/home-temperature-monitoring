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

| Type                 | Description                  | Config fields              |
|----------------------|------------------------------|----------------------------|
| `temperature-chart`  | Temperature line chart       | —                          |
| `live-readings`      | Live readings data grid      | —                          |
| `weather-forecast`   | Weather forecast card        | —                          |
| `sensor-health-pie`  | Sensor health pie chart      | —                          |
| `sensor-driver-pie`  | Sensor driver distribution pie | —                          |
| `health-timeline`    | Sensor health history chart  | `sensorName`               |
| `reading-stats`      | Reading statistics data grid | —                          |
| `notifications-feed` | Recent notifications list    | —                          |
| `markdown-note`      | Markdown text note           | `title`, `content`         |
| `current-reading`    | Single sensor current value  | `sensorName`               |
| `min-max-avg`        | Min/max/avg statistics       | `sensorName`, `hours`      |
| `gauge`              | Temperature gauge dial       | `sensorName`, `min`, `max` |
| `comparison-chart`   | Multi-sensor comparison      | `sensorIds`                |
| `group-summary`      | Sensor group summary table   | `sensorIds`                |
| `alert-summary`      | Active alert rules summary   | —                          |
| `uptime`             | Sensor uptime percentage     | `sensorName`, `days`       |
| `heatmap`            | Temperature heatmap grid     | `sensorName`, `days`       |
