# Telemetry Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Add full OpenTelemetry-based observability (structured logging, distributed tracing, metrics with Prometheus endpoint) to Sensor Hub's server mode.

**Architecture:** A new `telemetry` package initialises slog + OTel providers at startup, before any other component. The `*slog.Logger` is passed explicitly to services/repos. If `OTEL_EXPORTER_OTLP_ENDPOINT` is set, logs/traces/metrics export to a collector; otherwise everything stays local (JSON logs to file, `/metrics` for Prometheus). CLI commands are unaffected.

**Tech Stack:** Go `log/slog`, OpenTelemetry Go SDK, `otelgin`, `otelsql`, `otelslog` bridge, Prometheus exporter

**Design Doc:** `docs/plans/2026-03-26-telemetry-design.md`

---

## Task Overview

| # | Task | Phase | Depends On |
|---|------|-------|------------|
| 1 | Add OTel dependencies | 1: Foundation | — |
| 2 | Create `telemetry` package — logger | 1: Foundation | 1 |
| 3 | Create `telemetry` package — tracing | 1: Foundation | 1 |
| 4 | Create `telemetry` package — metrics | 1: Foundation | 1 |
| 5 | Create `telemetry` package — init/shutdown | 1: Foundation | 2, 3, 4 |
| 6 | Wire telemetry into `cmd/serve.go` | 2: Integration | 5 |
| 7 | Replace Gin logger with OTel middleware | 2: Integration | 6 |
| 8 | Add Prometheus `/metrics` endpoint | 2: Integration | 7 |
| 9 | Instrument database layer with `otelsql` | 2: Integration | 6 |
| 10 | Add context parameter to repository interfaces | 3: Context Threading | 9 |
| 11 | Thread context through service layer | 3: Context Threading | 10 |
| 12 | Instrument `SensorService` | 4: Service Instrumentation | 11 |
| 13 | Instrument `CleanupService` | 4: Service Instrumentation | 11 |
| 14 | Instrument `AlertService` | 4: Service Instrumentation | 11 |
| 15 | Instrument `AuthService` | 4: Service Instrumentation | 11 |
| 16 | Instrument `NotificationService` | 4: Service Instrumentation | 11 |
| 17 | Instrument `ApiKeyService` | 4: Service Instrumentation | 11 |
| 18 | Instrument WebSocket hub | 4: Service Instrumentation | 11 |
| 19 | Instrument SMTP and OAuth | 4: Service Instrumentation | 11 |
| 20 | Instrument API handlers | 4: Service Instrumentation | 11 |
| 21 | Update configuration | 5: Config & Packaging | 6 |
| 22 | Update packaging | 5: Config & Packaging | 21 |
| 23 | Go backend logging audit | 6: Audit | 12-20 |
| 24 | React UI logging cleanup | 6: Audit | — |
| 25 | Docusaurus telemetry documentation | 7: Docs & Verify | 23, 24 |
| 26 | Final verification | 7: Docs & Verify | 25 |

---

## Phase 1: Foundation

### Task 1: Add OTel Dependencies

**Files:**
- Modify: `sensor_hub/go.mod`

**Step 1: Install OTel packages**

```bash
cd sensor_hub
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/sdk/log@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc@latest
go get go.opentelemetry.io/otel/exporters/prometheus@latest
go get go.opentelemetry.io/contrib/bridges/otelslog@latest
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin@latest
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@latest
go get github.com/XSAM/otelsql@latest
go get github.com/prometheus/client_golang/prometheus/promhttp@latest
go mod tidy
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: Success with no errors.

### Task 2: Create `telemetry` Package — Logger

**Files:**
- Create: `sensor_hub/telemetry/logger.go`

**Step 1: Create the logger module**

Create `sensor_hub/telemetry/logger.go` with:
- `multiHandler` struct implementing `slog.Handler` — fans out records to multiple handlers
- `ParseLogLevel(level string) slog.Level` — converts "debug"/"info"/"warn"/"error" strings
- `NewLogger(level slog.Level, writer io.Writer, logProvider *sdklog.LoggerProvider) *slog.Logger` — creates a JSON slog.Logger that writes to the given writer, with an optional OTel bridge handler if logProvider is non-nil
- `LogWriter(logFilePath string) (io.Writer, *os.File, error)` — returns an io.MultiWriter(stdout, file) or just stdout if path is empty

The multiHandler's `Enabled`, `Handle`, `WithAttrs`, `WithGroup` methods delegate to all child handlers.

**Step 2: Verify build**

Run: `cd sensor_hub && go build ./...`
Expected: Success.

### Task 3: Create `telemetry` Package — Tracing

**Files:**
- Create: `sensor_hub/telemetry/tracing.go`

**Step 1: Create the tracing module**

Create `sensor_hub/telemetry/tracing.go` with:
- `initTracerProvider(ctx, res *resource.Resource, exportEnabled bool) (*sdktrace.TracerProvider, error)` — creates a TracerProvider with the given resource; if exportEnabled, adds an OTLP gRPC batcher exporter; sets the global TracerProvider via `otel.SetTracerProvider(tp)`
- `Tracer(name string) trace.Tracer` — convenience wrapper around `otel.Tracer(name)`

**Step 2: Verify build**

Run: `cd sensor_hub && go build ./...`
Expected: Success.

### Task 4: Create `telemetry` Package — Metrics

**Files:**
- Create: `sensor_hub/telemetry/metrics.go`

**Step 1: Create the metrics module**

Create `sensor_hub/telemetry/metrics.go` with:
- `initMeterProvider(ctx, res *resource.Resource, exportEnabled bool) (*sdkmetric.MeterProvider, http.Handler, error)` — creates a MeterProvider with a Prometheus reader (always) and an OTLP gRPC periodic reader (if exportEnabled); sets the global MeterProvider; returns the Prometheus HTTP handler
- `Meter(name string) otelmetric.Meter` — convenience wrapper around `otel.Meter(name)`

**Step 2: Verify build**

Run: `cd sensor_hub && go build ./...`
Expected: Success.

### Task 5: Create `telemetry` Package — Init/Shutdown

**Files:**
- Create: `sensor_hub/telemetry/telemetry.go`

**Step 1: Create the orchestrator**

Create `sensor_hub/telemetry/telemetry.go` with:

```go
type Telemetry struct {
    Logger            *slog.Logger
    TracerProvider    *sdktrace.TracerProvider
    MeterProvider     *sdkmetric.MeterProvider
    LogProvider       *sdklog.LoggerProvider
    PrometheusHandler http.Handler
    logFile           *os.File
}

type Config struct {
    ServiceName string
    Version     string
    LogLevel    slog.Level
    LogFilePath string
}
```

- `Init(ctx context.Context, cfg Config) (*Telemetry, error)`:
  1. Check `OTEL_EXPORTER_OTLP_ENDPOINT` env var to determine exportEnabled
  2. Check `OTEL_SERVICE_NAME` env var to override service name
  3. Build OTel `resource.Resource` with service name, version, host, OS, process attributes
  4. If exportEnabled, create `sdklog.LoggerProvider` with OTLP log exporter
  5. Call `LogWriter()` and `NewLogger()` to create the slog logger
  6. Call `initTracerProvider()` and `initMeterProvider()`
  7. Log "telemetry initialised" with config summary
  8. Return `*Telemetry`

- `Shutdown()`:
  1. Create a 5-second timeout context
  2. Shutdown TracerProvider, MeterProvider, LogProvider (ignore errors)
  3. Close log file if open

**Step 2: Verify build**

Run: `cd sensor_hub && go mod tidy && go build ./...`
Expected: Success.

**Step 3: Run all tests**

Run: `cd sensor_hub && go test ./...`
Expected: All existing tests pass.

---

## Phase 2: Integration

### Task 6: Wire Telemetry into `cmd/serve.go`

**Files:**
- Modify: `sensor_hub/cmd/serve.go` (lines 25-164)

**Step 1: Initialise telemetry early in RunE**

In the `serveCmd` RunE function, replace the current log setup (lines 39-48 — `log.SetPrefix`, log file opening, `io.MultiWriter`) with a call to `telemetry.Init()`.

The telemetry init must happen BEFORE `appProps.InitialiseConfig()` because we need a logger for config loading. However, the log level comes from application.properties. Solution: init telemetry with default INFO level first, then after config loads, update the level if configured differently. Or: read only the log.level property early. Simplest approach: init at INFO, config loads, and if the user has set a different level, log a message about the active level.

Actually, the cleanest approach: parse the log level from application.properties BEFORE full telemetry init. The `appProps` package loads config from files. We can read the single property early:

1. Call `appProps.InitialiseConfig(configDir)` first (it uses basic log internally — that's fine, it'll be replaced in the audit)
2. Read `appProps.AppConfig.LogLevel` (new field we'll add in Task 21)
3. Call `telemetry.Init(ctx, telemetry.Config{...})` with the parsed level
4. Pass `tel.Logger` to all service constructors

Replace:
```go
log.SetPrefix("sensor-hub: ")
// ... log file setup ...
```

With:
```go
appProps.InitialiseConfig(configDir)

logLevel := telemetry.ParseLogLevel(appProps.AppConfig.LogLevel)
tel, err := telemetry.Init(cmd.Context(), telemetry.Config{
    ServiceName: "sensor-hub",
    Version:     Version,
    LogLevel:    logLevel,
    LogFilePath: logFile,
})
if err != nil {
    return fmt.Errorf("failed to initialise telemetry: %w", err)
}
defer tel.Shutdown()

logger := tel.Logger
```

**Step 2: Thread logger to service constructors**

For now, just add `logger` as the last parameter to each `New*Service()` call. The services won't accept it yet — we'll fix that in Tasks 12-19. For this task, just update serve.go so it compiles by temporarily ignoring the logger (or passing it but not wiring it through yet).

Actually, better approach: update serve.go to store the logger, and update service constructors one at a time in later tasks. For now, store `logger` in serve.go and log startup events with it:

```go
logger.Info("database connected")
logger.Info("starting API server")
```

Replace the existing `log.Printf(...)` calls in serve.go with `logger.Info(...)` / `logger.Error(...)`.

**Step 3: Verify build and tests**

Run: `cd sensor_hub && go build ./... && go test ./...`
Expected: All pass.

### Task 7: Replace Gin Logger with OTel Middleware

**Files:**
- Modify: `sensor_hub/api/api.go` (line 18 and surrounding)
- Create: `sensor_hub/telemetry/gin_middleware.go`

**Step 1: Create slog Gin middleware**

Create `sensor_hub/telemetry/gin_middleware.go`:

A `GinLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc` that:
1. Records the start time
2. Calls `c.Next()`
3. Logs at INFO (or WARN if status >= 400, ERROR if >= 500) with attributes: method, path, status, latency_ms, client_ip, user_agent, trace_id (from span context)

**Step 2: Replace gin.Default() with gin.New()**

In `api/api.go`, change:
```go
router := gin.Default()
```
to:
```go
router := gin.New()
router.Use(gin.Recovery())
router.Use(otelgin.Middleware("sensor-hub"))
router.Use(telemetry.GinLoggerMiddleware(logger))
```

This requires `api.go` to receive the logger. Add a package-level `var logger *slog.Logger` and an `InitLogger(l *slog.Logger)` function, or pass it through existing `InitialiseAndListen()`.

The cleanest approach: add a `logger` parameter to `Initialise()` (or create a new init function). The `Initialise` function currently creates the router and registers routes. Modify it to accept `*slog.Logger` and `http.Handler` (for prometheus).

**Step 3: Verify build**

Run: `cd sensor_hub && go build ./...`
Expected: Success.

### Task 8: Add Prometheus `/metrics` Endpoint

**Files:**
- Modify: `sensor_hub/api/api.go`

**Step 1: Register the metrics endpoint**

In `api/api.go`, after router creation, add:

```go
router.GET("/metrics", gin.WrapH(prometheusHandler))
```

The `prometheusHandler` is `tel.PrometheusHandler` passed from `serve.go` → `api.Initialise()`.

**Step 2: Verify build**

Run: `cd sensor_hub && go build ./...`
Expected: Success.

### Task 9: Instrument Database Layer with `otelsql`

**Files:**
- Modify: `sensor_hub/db/db.go` (lines 1-76)

**Step 1: Wrap the SQL driver with otelsql**

In `db.go`, the `InitialiseDatabase()` function currently calls `sql.Open("sqlite3", ...)`.

Replace with:
```go
import "github.com/XSAM/otelsql"

// Register the instrumented driver
driverName, err := otelsql.Register("sqlite3",
    otelsql.WithAttributes(semconv.DBSystemSqlite),
)
if err != nil { ... }

db, err := sql.Open(driverName, dsn)
```

This automatically creates spans for all DB queries and records `db.client.operation.duration` metrics.

Also add `otelsql.RecordStats(db)` to register DB connection pool metrics.

**Step 2: Verify build and tests**

Run: `cd sensor_hub && go build ./... && go test ./...`
Expected: All pass. The test DB uses the same `sql.Open` path — otelsql works with any driver, including the test mattn/sqlite3 driver. Note: If tests use sqlmock, otelsql wrapping may need to be conditional or the tests may need adjustment.

Check: If tests fail due to otelsql + sqlmock incompatibility, make the otelsql wrapping conditional (only in production, not in tests) by accepting a `*sql.DB` parameter instead of creating it internally.

---

## Phase 3: Context Threading

### Task 10: Add `context.Context` Parameter to Repository Interfaces

**Files:**
- Modify: `sensor_hub/db/sensor_repository.go` — interface + all methods
- Modify: `sensor_hub/db/temperature_repository.go` — interface + all methods
- Modify: `sensor_hub/db/alert_repository.go` — interface + all methods
- Modify: `sensor_hub/db/notification_repository.go` — interface + all methods
- Modify: `sensor_hub/db/user_repository.go` — interface + all methods
- Modify: `sensor_hub/db/session_repository.go` — interface + all methods
- Modify: `sensor_hub/db/failed_login_repository.go` — interface + all methods
- Modify: `sensor_hub/db/role_repository.go` — interface + all methods
- Modify: `sensor_hub/db/api_key_repository.go` — interface + all methods
- Modify: All corresponding `*_test.go` files
- Modify: `sensor_hub/service/test_mocks_test.go` — all mock implementations

**Step 1: Add `ctx context.Context` as the first parameter to every repository interface method and every implementation method**

This is a large mechanical change. For each repository:

1. Add `ctx context.Context` as first param in the interface definition
2. Add `ctx context.Context` as first param in the implementation
3. Change `r.db.Query(...)` → `r.db.QueryContext(ctx, ...)`
4. Change `r.db.Exec(...)` → `r.db.ExecContext(ctx, ...)`
5. Change `r.db.QueryRow(...)` → `r.db.QueryRowContext(ctx, ...)`

This enables otelsql to propagate trace context into DB spans.

**Step 2: Update all test mocks**

In `service/test_mocks_test.go` and any test files, update mock method signatures to include `ctx context.Context`.

**Step 3: Update all test call sites**

In `db/*_test.go`, pass `context.Background()` (or `context.TODO()`) to all repository method calls.

**Step 4: Verify build and tests**

Run: `cd sensor_hub && go build ./... && go test ./...`
Expected: Compilation errors until service layer is also updated (Task 11). Run build first to see what breaks, then proceed to Task 11 before running tests.

### Task 11: Thread Context Through Service Layer

**Files:**
- Modify: All files in `sensor_hub/service/` — every method that calls a repository
- Modify: `sensor_hub/alerting/service.go` — all methods
- Modify: `sensor_hub/api/*.go` — all handler functions (pass `c.Request.Context()` to services)
- Modify: All test files for services and API handlers

**Step 1: Add `ctx context.Context` to service method signatures**

For every service method that calls a repository method, add `ctx context.Context` as the first parameter. Pass it through to the repository calls.

For methods called from Gin handlers, the context comes from `c.Request.Context()`.

For background goroutines (periodic collection, cleanup), create a new `context.Background()` at the start of each cycle.

**Step 2: Update API handlers**

In every API handler, pass `c.Request.Context()` to service calls:

```go
// Before:
sensors, err := sensorService.ServiceGetAllSensors()
// After:
sensors, err := sensorService.ServiceGetAllSensors(c.Request.Context())
```

**Step 3: Update service tests**

Update all service test call sites to pass `context.Background()`.

**Step 4: Verify build and all tests pass**

Run: `cd sensor_hub && go build ./... && go test ./...`
Expected: All pass. This is the critical milestone — after this, the full context chain works from HTTP request → service → repository → DB with trace propagation.

---

## Phase 4: Service Instrumentation

For Tasks 12-20, the pattern is the same for each service:

1. Add `logger *slog.Logger` field to the service struct
2. Add `logger *slog.Logger` parameter to the constructor
3. Create a child logger: `logger.With("component", "service_name")`
4. Replace all `log.Printf/Println` calls with `logger.Info/Warn/Error/Debug`
5. Add trace spans for significant operations using `telemetry.Tracer("component")`
6. Add metrics for key operations
7. Update the constructor call in `cmd/serve.go`
8. Update test mocks/constructors
9. Verify build and tests

### Task 12: Instrument `SensorService`

**Files:**
- Modify: `sensor_hub/service/sensor_service.go` (23 log statements)
- Modify: `sensor_hub/cmd/serve.go` (constructor call)
- Modify: Service tests

**Step 1: Add logger and tracer to struct**

```go
type SensorService struct {
    sensorRepo   database.SensorRepositoryInterface[types.Sensor]
    tempRepo     database.ReadingsRepository[types.TemperatureReading]
    alertService *alerting.AlertService
    notifSvc     NotificationServiceInterface
    logger       *slog.Logger
}
```

Update `NewSensorService` to accept `logger *slog.Logger` and store `logger.With("component", "sensor_service")`.

**Step 2: Replace all 23 log statements**

Examples:
- `log.Printf("Added sensor: %v", sensor)` → `s.logger.Info("sensor added", "sensor", sensor.Name, "type", sensor.Type)`
- `log.Printf("Error fetching temperature from sensor %s at %s: %v", ...)` → `s.logger.Error("failed to fetch temperature", "sensor", sensor.Name, "url", sensor.URL, "error", err)`
- `log.Printf("Starting periodic sensor readings collection")` → `s.logger.Info("starting periodic sensor collection")`

**Step 3: Add trace spans**

Key spans:
- `ServiceCollectAndStoreAllSensorReadings` — root span for each collection cycle
- `ServiceCollectFromSensorByName` — child span per sensor
- `ServiceFetchTemperatureReadingFromSensor` — span for the external HTTP call
- `ServiceDiscoverSensors` — span for startup discovery

**Step 4: Add metrics**

```go
var (
    sensorCollectionCounter  metric.Int64Counter    // sensor_collection_total
    sensorCollectionDuration metric.Float64Histogram // sensor_collection_duration_seconds
)
```

Initialise in `NewSensorService` using `telemetry.Meter("sensor_service")`.
Record on each collection attempt: counter with sensor name + success/failure, histogram for duration.

**Step 5: Update serve.go constructor call and tests**

**Step 6: Verify build and tests**

Run: `cd sensor_hub && go build ./... && go test ./...`

### Task 13: Instrument `CleanupService`

**Files:**
- Modify: `sensor_hub/service/cleanup_service.go` (16 log statements)

Same pattern as Task 12. Key spans: `performCleanup`, `computeHourlyAverages`.
Replace all `log.Printf` with `cs.logger.Info/Error/Debug`.

### Task 14: Instrument `AlertService`

**Files:**
- Modify: `sensor_hub/alerting/service.go` (6 log statements)

Same pattern. Add `alerts_triggered_total` counter metric.
Key span: `ProcessReadingAlert`.

### Task 15: Instrument `AuthService`

**Files:**
- Modify: `sensor_hub/service/auth_service.go` (13 log statements)

Same pattern. Sensitive: do NOT log passwords or session tokens. Log: login attempts (success/failure), session creation/deletion, lockout events.

### Task 16: Instrument `NotificationService`

**Files:**
- Modify: `sensor_hub/service/notification_service.go` (5 log statements)

Same pattern. Key span: `SendNotification`.

### Task 17: Instrument `ApiKeyService`

**Files:**
- Modify: `sensor_hub/service/api_key_service.go` (1 log statement)

Same pattern. Log key creation, revocation, validation attempts. Do NOT log key values.

### Task 18: Instrument WebSocket Hub

**Files:**
- Modify: `sensor_hub/ws/hub.go` (5 log statements)

Add `active_websocket_connections` gauge metric. Log connections/disconnections at DEBUG, broadcast errors at ERROR.

### Task 19: Instrument SMTP and OAuth

**Files:**
- Modify: `sensor_hub/smtp/smtp.go` (3 log statements)
- Modify: `sensor_hub/oauth/oauth_service.go`

SMTP: span per email send, log success/failure. Do NOT log email content.
OAuth: span per token refresh, log refresh attempts.

### Task 20: Instrument API Handlers

**Files:**
- Modify: All files in `sensor_hub/api/` that contain `log.Printf` or `log.Println`
  - `api/alert_api.go` (5 statements)
  - `api/api.go` (3 statements)
  - `api/auth_api.go` (1 statement)
  - `api/sensor_api.go` (1 statement)
  - `api/temperature_api.go` (6 statements)
  - `api/websocket.go` (8 statements)

Each API file gets a package-level or handler-level logger. Replace all log.Printf with slog calls. The otelgin middleware handles request tracing — handlers don't need to create spans manually for the request itself, only for sub-operations.

---

## Phase 5: Config & Packaging

### Task 21: Update Configuration

**Files:**
- Modify: `sensor_hub/application_properties/application_configuration.go`
- Modify: `sensor_hub/application_properties/application_properties_defaults.go`
- Modify: `packaging/defaults/application.properties`

**Step 1: Add `LogLevel` field**

In `application_properties_defaults.go`, add to the AppConfiguration struct:
```go
LogLevel string
```

Set default to `"info"`.

**Step 2: Load the property**

In `application_configuration.go`, in the config loading function, read `log.level` from the properties and assign to `AppConfig.LogLevel`.

**Step 3: Add to defaults file**

In `packaging/defaults/application.properties`, add:
```properties
log.level=info
```

### Task 22: Update Packaging

**Files:**
- Modify: `packaging/environment`

**Step 1: Add OTel env var examples**

Append to `packaging/environment`:
```bash
# OpenTelemetry Collector endpoint (uncomment to enable telemetry export)
# OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# OTLP protocol: grpc (default, port 4317) or http/protobuf (port 4318)
# OTEL_EXPORTER_OTLP_PROTOCOL=grpc

# Override the default service name (default: sensor-hub)
# OTEL_SERVICE_NAME=sensor-hub
```

---

## Phase 6: Logging Audit

### Task 23: Go Backend Logging Audit

**Files:**
- Review and modify ALL Go files containing log statements (~138 across 24 files)
- Also review: `sensor_hub/application_properties/application_configuration.go` (16 statements)
- Also review: `sensor_hub/application_properties/application_properties_files.go` (5 statements)

**Audit checklist per file:**

1. ✅ Every `log.Printf`/`log.Println`/`log.Fatalf` replaced with `slog.Info`/`slog.Warn`/`slog.Error`/`slog.Debug`
2. ✅ Every error path has a WARN or ERROR log with the error as a structured attribute
3. ✅ Every significant operation (sensor added, reading collected, alert triggered, user created, config loaded) logged at INFO
4. ✅ Routine operations (periodic ticks, health checks passing, cache hits) logged at DEBUG
5. ✅ Structured context on every log: component name, entity IDs/names, trace IDs
6. ✅ No sensitive data logged: no passwords, no API key values, no session tokens, no email content
7. ✅ No remaining `log.` imports (standard library log package fully removed from server code)
8. ✅ No `fmt.Println` used for logging

**Level guide:**
- **DEBUG**: Periodic tick started, WS message broadcast, config property read, health check passed
- **INFO**: Sensor added/removed/enabled/disabled, reading collected, alert triggered, user created/deleted, API key created/revoked, server started/stopped, config loaded
- **WARN**: Retryable failures (sensor fetch timeout, email send retry), rate limiting triggered, deprecated config
- **ERROR**: Unrecoverable failures (DB connection lost, migration failed, sensor unreachable after retries)

### Task 24: React UI Logging Cleanup

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/tools/logger.ts`
- Modify: 38 console.* occurrences across ~18 files (see list below)

**Step 1: Create `logger.ts`**

```typescript
const isDev = import.meta.env.DEV;

export const logger = {
  debug: (...args: unknown[]) => { if (isDev) console.debug(...args); },
  info: (...args: unknown[]) => { if (isDev) console.log(...args); },
  warn: (...args: unknown[]) => console.warn(...args),
  error: (...args: unknown[]) => console.error(...args),
};
```

- `debug` and `info` are silent in production builds (Vite sets `import.meta.env.DEV = false`)
- `warn` and `error` always log (these are genuine issues)

**Step 2: Replace all console.* calls**

Files to modify (28 console.error → `logger.error`, 6 console.debug → `logger.debug`, 2 console.log → `logger.debug`):

- `api/Client.ts`: console.debug → `logger.debug`, console.error → `logger.error`
- `hooks/useCurrentTemperatures.ts`: console.error → `logger.error`, console.debug → `logger.debug`
- `hooks/useProperties.ts`: console.log → `logger.debug`, console.error → `logger.error`
- `hooks/useSensorForm.ts`: console.log → `logger.debug`
- `hooks/useSensors.ts`: console.error → `logger.error`, console.debug → `logger.debug`
- `hooks/useApiKeys.ts`: console.error → `logger.error`
- `hooks/useSensorHealthHistory.ts`: console.error → `logger.error`
- `hooks/useTotalReadingsForEachSensor.ts`: console.error → `logger.error`
- `hooks/useTemperatureData.ts`: console.error → `logger.error`
- `providers/NotificationProvider.tsx`: console.error → `logger.error`, console.debug → `logger.debug`
- `components/NotificationBell.tsx`: console.error → `logger.error`
- `components/AlertHistoryDialog.tsx`: console.error → `logger.error`
- `components/DeleteAlertDialog.tsx`: console.error → `logger.error`
- `components/EditAlertDialog.tsx`: console.error → `logger.error`
- `components/CreateAlertDialog.tsx`: console.error → `logger.error`
- `components/CreateUserDialog.tsx`: console.error → `logger.error`
- `components/DeleteUserDialog.tsx`: console.error → `logger.error`
- `components/EditUserDialog.tsx`: console.error → `logger.error`
- `components/SessionsCard.tsx`: console.error → `logger.error`
- `components/RolePermissionsCard.tsx`: console.error → `logger.error`
- `components/UserManagementCard.tsx`: console.error → `logger.error`
- `components/AlertRulesCard.tsx`: console.error → `logger.error`

**Step 3: Verify TypeScript and build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit && npx vite build`
Expected: Both pass.

---

## Phase 7: Docs & Verify

### Task 25: Docusaurus Telemetry Documentation

**Files:**
- Create: `docs/docs/telemetry.md`
- Modify: `docs/sidebars.ts`

**Step 1: Create telemetry.md**

Docusaurus page with frontmatter (`id: telemetry`, `title: Telemetry & Observability`, `sidebar_position: 9`).

Sections:
1. Overview — three pillars (logs, traces, metrics)
2. Configuration — log.level property, OTel env vars
3. Log Format — JSON schema with example lines
4. Prometheus Metrics — `/metrics` endpoint, all available metrics in a table
5. Exporting to OTel Collector — env var setup, gRPC vs HTTP, TLS
6. Grafana Integration — example Prometheus scrape config, example PromQL queries
7. Troubleshooting

**Step 2: Update sidebar**

Add `'telemetry'` to `docs/sidebars.ts` between `'llm-skills'` and the API Reference category.

**Step 3: Verify docs build**

Run: `cd docs && npx docusaurus build`
Expected: Success.

### Task 26: Final Verification

**Step 1: Full Go build and test**

```bash
cd sensor_hub && go build ./... && go test ./...
```

**Step 2: Cross-compile**

```bash
GOOS=darwin GOARCH=arm64 go build -o /dev/null ./...
GOOS=windows GOARCH=amd64 go build -o /dev/null ./...
```

**Step 3: Frontend build**

```bash
cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit && npx vite build
```

**Step 4: Docs build**

```bash
cd docs && npx docusaurus build
```

**Step 5: Verify no remaining `log.Printf`/`log.Println` in server code**

```bash
cd sensor_hub && grep -rn 'log\.Printf\|log\.Println\|log\.Fatalf' --include='*.go' | grep -v '_test.go' | grep -v 'cmd/' | grep -v vendor
```

Expected: Zero matches (CLI commands in cmd/ are exempt; test files are exempt).

**Step 6: Verify no remaining `console.log` in UI code**

```bash
cd sensor_hub/ui/sensor_hub_ui/src && grep -rn 'console\.log\b' --include='*.ts' --include='*.tsx'
```

Expected: Zero matches.

---

