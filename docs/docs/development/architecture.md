# Architecture

This guide explains how the system is structured, how components interact, and
how data flows through the application.

## System Overview

The system consists of small Raspberry Pis running temperature sensors and a
central hub that aggregates data, stores it, and serves a web UI.

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Pi + Sensor  │  │  Pi + Sensor  │  │  Pi + Sensor  │
│  (Flask API)  │  │  (Flask API)  │  │  (Flask API)  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │ HTTP GET /temperature        │                │
       └──────────────┬───────────────┘                │
                      │                                │
              ┌───────▼────────┐                       │
              │   Sensor Hub   │◄──────────────────────┘
              │   (Go binary)  │
              │                │
              │  ┌──────────┐  │
              │  │  SQLite   │  │
              │  └──────────┘  │
              │                │
              │  ┌──────────┐  │
              │  │ React UI │  │  (embedded in the binary)
              │  └──────────┘  │
              └───────┬────────┘
                      │
              ┌───────▼────────┐
              │     Nginx      │  (optional TLS reverse proxy)
              └───────┬────────┘
                      │
              ┌───────▼────────┐
              │    Browser /   │
              │    CLI client  │
              └────────────────┘
```

Each sensor Pi runs a tiny Flask API (`temperature_sensor/`) that exposes a
`GET /temperature` endpoint returning the current reading. Sensor Hub polls
these endpoints on a configurable interval (default 300 seconds), stores
readings in SQLite, evaluates alert rules, and broadcasts updates over
WebSocket to connected UI clients.

The Go binary embeds the built React SPA via `//go:embed`, so a single binary
serves both the REST API and the frontend.

## Internal Layers

The Go backend follows a three-layer architecture:

```
  HTTP Request
      │
      ▼
┌─────────────────────────────────────────────┐
│  Router & Middleware (Gin)                   │
│  gin.Recovery → OTEL → Logger → CORS → CSRF │
│                                             │
│  Per-route: AuthRequired → RequirePermission │
└────────────────────┬────────────────────────┘
                     │
      ┌──────────────▼──────────────┐
      │  API Handlers (api/*.go)    │
      │  HTTP ↔ JSON, validation    │
      └──────────────┬──────────────┘
                     │
      ┌──────────────▼──────────────┐
      │  Services (service/*.go)    │
      │  Business logic, WebSocket  │
      │  broadcasts, alert checks   │
      └──────────────┬──────────────┘
                     │
      ┌──────────────▼──────────────┐
      │  Repositories (db/*.go)     │
      │  SQL queries, data mapping  │
      └──────────────┬──────────────┘
                     │
      ┌──────────────▼──────────────┐
      │  SQLite (WAL mode, FK on)   │
      └─────────────────────────────┘
```

### Handlers (`api/`)

Each resource area has a handler file and a routes file:

- `temperature_api.go` / `temperature_routes.go`
- `sensor_api.go` / `sensor_routes.go`
- `alert_api.go` / `alert_routes.go`
- etc.

Handler functions follow the naming convention `verbNounHandler` (e.g.
`getReadingsBetweenDatesHandler`, `addSensorHandler`). Each handler file has a
package-level service variable and an `InitXxxAPI(service)` function called at
startup.

### Services (`service/`)

Services contain business logic, coordinate repository calls, and broadcast
WebSocket updates. They are injected into handlers at startup. Constructor:

```go
func NewSensorService(
    sensorRepo database.SensorRepositoryInterface[types.Sensor],
    tempRepo   database.ReadingsRepository,
    alertRepo  database.AlertRepository,
    notifSvc   NotificationServiceInterface,
    logger     *slog.Logger,
) *SensorService
```

Public method naming convention: `ServiceVerbNoun` (e.g.
`ServiceGetSensorByName`, `ServiceCollectAndStoreAllSensorReadings`).

### Repositories (`db/`)

Repositories wrap SQL queries and return typed Go structs. Each defines an
interface (e.g. `SensorRepositoryInterface[T]`, `ReadingsRepository`) and a
concrete implementation. Constructor parameters are always `(db *sql.DB, ...deps, logger *slog.Logger)`.

All string-equality WHERE clauses use `LOWER(col) = LOWER(?)` for
case-insensitive matching (SQLite text comparison is case-sensitive by default).

## Data Flow: Collecting a Temperature Reading

Here is the end-to-end path of a single sensor reading:

```
1. Periodic task fires (every 300s by default)
   periodic.RunTask → SensorService.ServiceCollectAndStoreAllSensorReadings

2. Service calls SensorRepository.GetSensorsByDriver("sensor-hub-http-temperature")
   → returns all enabled sensors

3. For each sensor, HTTP GET to sensor.URL + "/temperature"
   → parses JSON response into Reading

4. ReadingsRepository.Add() inserts readings into readings table

5. SensorRepository.UpdateSensorHealthById() updates the sensor's health_status

6. ws.BroadcastToTopic("current-readings", readings)
   → pushes readings to all connected UI WebSocket clients

7. AlertService.ProcessReadingAlert() evaluates alert rules
   → if threshold breached and rate limit allows: create notification + send email
```

## Application Startup Sequence

The `serve` command (`cmd/serve.go`) initialises everything in this order:

1. **Signal context** — `signal.NotifyContext` for SIGINT/SIGTERM
2. **Configuration** — `InitialiseConfig(configDir)` loads application.properties,
   smtp.properties, database.properties
3. **Telemetry** — `telemetry.Init()` sets up slog and Prometheus metrics
4. **Database** — `InitialiseDatabase()` opens SQLite, runs migrations
5. **Repositories** — created in dependency order (sensor → temperature → alert → etc.)
6. **Services** — each receives its repository dependencies via constructor injection
7. **API handlers** — `InitXxxAPI(service)` wires each handler to its service
8. **Middleware** — `InitAuthMiddleware`, `InitPermissionMiddleware`, `InitApiKeyMiddleware`
9. **Initial admin** — creates admin user from `SENSOR_HUB_INITIAL_ADMIN` env var if no users exist
10. **Sensor discovery** — reads `openapi.yaml` to auto-register sensors (if configured)
11. **OAuth** — initialises Gmail OAuth (optional, failure is non-fatal)
12. **Periodic tasks** — starts sensor collection and data cleanup goroutines
13. **HTTP server** — `api.InitialiseAndListen()` starts Gin on the configured port

## Graceful Shutdown

When a SIGINT or SIGTERM is received:

1. The signal context is cancelled
2. Periodic tasks detect `ctx.Done()` and exit their loops
3. Deferred cleanup runs in reverse order: database close, telemetry shutdown

## Periodic Task Supervision

Background tasks (sensor collection, data cleanup, hourly averages) run via the
`periodic` package (`periodic/periodic.go`). Each task is a supervised
goroutine with:

- **Panic recovery** — catches panics with `defer/recover`, logs full stack trace
- **Exponential backoff** — 5s → 10s → 20s → ... → 5min cap on consecutive panics
- **Automatic restart** — re-enters the event loop after backoff
- **Success reset** — consecutive panic counter resets to 0 on a successful execution
- **Context cancellation** — clean exit when the application shuts down

Usage:

```go
periodic.RunTask(ctx, periodic.TaskConfig{
    Name:           "sensor_collection",
    Interval:       300 * time.Second,
    Logger:         logger,
    RunImmediately: true,
}, func(ctx context.Context) error {
    return sensorService.ServiceCollectAndStoreAllSensorReadings(ctx)
})
```

## Middleware Chain

Middleware is applied in this order for every request:

| Order | Middleware | Purpose |
|-------|-----------|---------|
| 1 | `gin.Recovery()` | Catch panics in handlers |
| 2 | `otelgin.Middleware` | OpenTelemetry trace spans |
| 3 | `GinLoggerMiddleware` | Structured request logging |
| 4 | `cors.New()` | CORS headers (if enabled) |
| 5 | `CSRFMiddleware` | Validate X-CSRF-Token on state-changing requests |

Then per-route:

| Middleware | Purpose |
|-----------|---------|
| `AuthRequired()` | Validates session cookie or API key, sets `currentUser` in context |
| `RequirePermission(perm)` | Checks user has the named permission |

`AuthRequired` checks the `X-API-Key` header first, then falls back to the
session cookie. If the user has `MustChangePassword` set, only login, logout,
current-user, and change-password endpoints are accessible.

## Key Directories

| Directory | Contents |
|-----------|----------|
| `cmd/` | CLI commands (Cobra). `serve.go` is the main server entry point |
| `api/` | HTTP handlers and route registration |
| `api/middleware/` | Auth, CSRF, and permission middleware |
| `service/` | Business logic layer |
| `db/` | Repository layer and SQL migrations |
| `db/migrations/` | golang-migrate SQL migration files |
| `periodic/` | Supervised periodic task runner |
| `ws/` | WebSocket hub and connection management |
| `types/` | Shared data types |
| `alerting/` | Alert rule evaluation logic |
| `notifications/` | Notification dispatch (in-app + email) |
| `telemetry/` | Logging, tracing, metrics setup |
| `application_properties/` | Configuration loading and parsing |
| `ui/sensor_hub_ui/` | React SPA source |
| `ui/sensor_hub_ui/src/dashboard/` | Dashboard engine, widget registry, and page components |
| `web/` | Embedded UI assets (built by npm, included via `//go:embed`) |
| `integration/` | Integration tests |
| `testharness/` | Test infrastructure for integration tests |
