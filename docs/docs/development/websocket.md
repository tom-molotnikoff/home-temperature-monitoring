# WebSocket

This guide explains the real-time update system, how the WebSocket hub works,
and how the frontend integrates with it.

## Overview

Sensor Hub uses WebSocket connections to push real-time updates to connected
browser clients. When sensor readings are collected, sensors are updated, or
notifications are created, the backend broadcasts the change over WebSocket so
the UI updates without polling.

## Hub Architecture

The WebSocket system lives in `ws/` and follows a hub-and-spoke pattern:

```
                    ┌──────────────┐
  Service ────────► │  Hub         │
  broadcasts       │              │
  via topic        │  conns map:  │
                    │  conn1 → [current-temperatures, sensors:Temperature]
                    │  conn2 → [current-temperatures, notifications:user:5]
                    │  conn3 → [sensors:Temperature]
                    └──┬───┬───┬──┘
                       │   │   │
                    write pumps (one goroutine per connection)
                       │   │   │
                    ┌──▼┐ ┌▼──┐┌▼──┐
                    │ws1│ │ws2││ws3│  Browser clients
                    └───┘ └───┘└───┘
```

### Core Types

```go
type Hub struct {
    mu           sync.Mutex
    conns        map[*websocket.Conn]*connInfo
    writeTimeout time.Duration  // 5 seconds
    logger       *slog.Logger
}

type connInfo struct {
    conn   *websocket.Conn
    send   chan any            // 16-message buffer
    topics map[string]bool
}
```

A global `DefaultHub` singleton is used throughout the application.

### Connection Lifecycle

1. **Upgrade** — HTTP GET upgrades to WebSocket via Gin handler
2. **Register** — connection added to hub with initial topic subscriptions.
   Two goroutines start: write pump (sends from channel) and read pump
   (detects disconnect)
3. **Broadcast** — services call `BroadcastToTopic(topic, data)` which queues
   messages to all subscribed connections
4. **Unregister** — on disconnect or write error, the connection is removed and
   its channel closed

### Backpressure

Each connection has a 16-message buffered channel. If a client falls behind and
the buffer fills, the connection is dropped to prevent slow clients from
blocking others.

## Topics

Topics are string identifiers. Services broadcast to specific topics, and
clients subscribe when their WebSocket connection opens.

| Topic | Data Type | Triggered By |
|-------|-----------|-------------|
| `current-temperatures` | `[]TemperatureReading` | Sensor reading collection |
| `sensors:Temperature` | `[]Sensor` | Sensor add/update/delete/enable/disable |
| `sensors:Humidity` | `[]Sensor` | Same (per sensor type) |
| `notifications:user:{id}` | `Notification` | New notification created |

### Service Broadcasting Examples

**After collecting readings** (`sensor_service.go`):

```go
ws.BroadcastToTopic("current-temperatures", []types.TemperatureReading{reading})
```

**After sensor state changes** (`sensor_service.go`):

```go
func (s *SensorService) broadcastSensors(ctx context.Context) {
    sensors, _ := s.sensorRepo.GetAllSensors(ctx)
    byType := make(map[string][]types.Sensor)
    for _, sensor := range sensors {
        byType[sensor.Type] = append(byType[sensor.Type], sensor)
    }
    for t, list := range byType {
        ws.BroadcastToTopic("sensors:"+t, list)
    }
}
```

**After creating a notification** (`ws/notification_broadcaster.go`):

```go
func (b *NotificationBroadcaster) BroadcastToUser(userID int, message interface{}) {
    topic := fmt.Sprintf("notifications:user:%d", userID)
    DefaultHub.BroadcastToTopic(topic, message)
}
```

## Message Formats

All messages are sent as JSON via `conn.WriteJSON()`.

### Temperature Readings

```json
[
  {
    "sensor_name": "Living Room",
    "time": "2026-03-27 09:30:45",
    "temperature": 22.5
  }
]
```

### Sensors

```json
[
  {
    "id": 1,
    "name": "Living Room",
    "type": "Temperature",
    "url": "http://192.168.1.100:8080",
    "health_status": "good",
    "health_reason": "successful reading",
    "enabled": true
  }
]
```

### Notifications

```json
{
  "id": 456,
  "category": "threshold_alert",
  "severity": "warning",
  "title": "Alert: Living Room Temperature High",
  "message": "Temperature above threshold (value: 28.50)",
  "metadata": {
    "sensor_name": "Living Room",
    "sensor_type": "Temperature",
    "numeric_value": 28.5
  },
  "created_at": "2026-03-27T09:30:45Z"
}
```

## WebSocket Endpoints

All endpoints require authentication (session cookie is sent on the WebSocket
upgrade request).

| Endpoint | Permission | Topic |
|----------|-----------|-------|
| `GET /api/temperature/ws/current-temperatures` | `view_readings` | `current-temperatures` |
| `GET /api/sensors/ws/{type}` | (auth only) | `sensors:{type}` |
| `GET /api/notifications/ws` | `view_notifications` | `notifications:user:{userId}` |

### Push vs Interval WebSocket

The codebase supports two WebSocket patterns:

**Push WebSocket** (used in production) — services broadcast when events occur:

```go
func createPushWebSocket(ctx *gin.Context, topic string) {
    conn, _ := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
    ws.Register(conn, []string{topic})
}
```

**Interval WebSocket** (available but not currently used) — calls a function on
a timer and pushes the result:

```go
func createIntervalBasedWebSocket(ctx *gin.Context, topic string,
    methodToCall func() (any, error), intervalSeconds int)
```

## Frontend Integration

The React UI connects to WebSocket endpoints using the native `WebSocket` API.

### Connection Pattern

```typescript
useEffect(() => {
    if (!user) return;

    const ws = new WebSocket(`${WEBSOCKET_BASE}/temperature/ws/current-temperatures`);

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        // Update React state
    };

    ws.onerror = (err) => logger.error(err);

    return () => ws.close();  // Cleanup on unmount
}, [user]);
```

### Key Implementation Details

- **Notifications** (`NotificationProvider.tsx`) — connects when user has
  `view_notifications` permission. Parses incoming notifications and updates
  both the notification list and unread count in context
- **Current temperatures** (`useCurrentTemperatures.ts`) — handles both array
  snapshots (initial load) and single reading updates
- **Sensor list** (`useSensors.ts`) — subscribes to `sensors:{type}` topics to
  keep the sensor list current

### Environment Configuration

```typescript
// environment/Environment.ts
export const WEBSOCKET_BASE = import.meta.env.VITE_WEBSOCKET_BASE || '/api';
```

In production the WebSocket connects through the same origin. In development
with Vite, set `VITE_WEBSOCKET_BASE` to point to the Go backend (the Vite dev
server proxies API requests but WebSocket connections need the direct URL).
