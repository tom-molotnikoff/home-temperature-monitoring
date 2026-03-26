# Telemetry Design — OpenTelemetry Integration

**Date:** 2026-03-26
**Status:** Approved

## Problem

Sensor Hub uses Go's basic `log` package with no structured logging, log levels, metrics, or tracing.
Diagnosing production issues is difficult because log output is unstructured plain text with no context (no trace IDs, no component attribution, no severity levels).
There is no way to monitor application health beyond the simple `/api/health` endpoint.

## Proposed Approach

Add full observability using OpenTelemetry (OTel) with three pillars:

1. **Structured logging** via Go's `log/slog` with JSON output and the OTel slog bridge
2. **Distributed tracing** via OTel TracerProvider with spans across HTTP, DB, sensor collection, and background tasks
3. **Metrics** via OTel MeterProvider with Prometheus `/metrics` endpoint

All telemetry is **local-first**: structured logs go to file and stdout, metrics are available via Prometheus scrape endpoint.
If `OTEL_EXPORTER_OTLP_ENDPOINT` is set, logs, traces, and metrics are also exported to an OTel Collector.

CLI commands are unaffected — only the server (`serve`) gets telemetry.

## Architecture

### Telemetry Package (`sensor_hub/telemetry/`)

A new `telemetry` package handles all initialisation and provides helpers.

**Files:**

- **`telemetry.go`** — Main entry point. `Init(configDir, logFile, version string) (*Telemetry, error)` creates all providers. `Shutdown(ctx)` flushes exporters with a 5-second timeout. Called early in `serve.go`, with `defer telemetry.Shutdown()`.

- **`logger.go`** — Configures slog with a multi-handler: `slog.JSONHandler` writing to the log file/stdout (structured JSON with timestamps, levels, trace IDs), plus the OTel slog bridge handler if a collector is configured. Exposes `Logger() *slog.Logger`. Also provides a `GinLoggerMiddleware()` that logs each HTTP request with method, path, status, latency, client IP, trace ID, and user ID.

- **`tracing.go`** — Sets up TracerProvider with OTLP exporter (or noop if no collector). Resource attributes: service name (`sensor-hub` default), version, environment. Exposes `Tracer(name string) trace.Tracer`.

- **`metrics.go`** — Sets up MeterProvider with OTLP exporter + Prometheus exporter. Registers Go runtime metrics (goroutines, memory, GC). Exposes `Meter(name string) metric.Meter` and `PrometheusHandler() http.Handler` for the `/metrics` endpoint.

### Logger Passing

The `*slog.Logger` is **explicitly passed** to services and repositories — not set as a global.
Each component creates a child logger with its own attributes:

```go
sensorLogger := logger.With("component", "sensor_service")
```

This keeps logging testable and gives every log line component attribution.

### Log Output Format

The log file at `/var/log/sensor-hub/sensor-hub.log` changes from unstructured text to **JSON lines**.

**Before:**
```
sensor-hub: Added sensor: kitchen
sensor-hub: Error fetching temperature from sensor kitchen at http://...: connection refused
```

**After:**
```json
{"time":"2026-03-26T13:40:00Z","level":"INFO","msg":"sensor added","component":"sensor_service","sensor":"kitchen","type":"Temperature"}
{"time":"2026-03-26T13:40:05Z","level":"ERROR","msg":"failed to fetch temperature","component":"sensor_service","sensor":"kitchen","url":"http://...","error":"connection refused","trace_id":"abc123"}
```

## Instrumentation Coverage

### HTTP Layer (`api/`)

- Replace `gin.Default()` with `gin.New()` + `gin.Recovery()` + `otelgin.Middleware("sensor-hub")` + slog request logger middleware
- Each request logged at INFO with: method, path, status, latency, client IP, trace ID, user ID (if authenticated)
- Metrics: `http_requests_total` (counter by method/path/status), `http_request_duration_seconds` (histogram by method/path/status)

### Database Layer (`db/`)

- Wrap `*sql.DB` with `otelsql` for automatic query span creation and duration recording
- Repository methods receive `context.Context` parameter to propagate trace context
- Logger added to repository structs for error/warn logging with query context
- Metrics: `db_query_duration_seconds` (histogram by operation)

### Service Layer (`service/`, `alerting/`)

- Each service receives a `*slog.Logger` with component attribute
- Sensor collection: span per collection cycle, child span per sensor HTTP call
- Alert evaluation: span per evaluation, logged at INFO (triggered) or DEBUG (not triggered)
- Cleanup operations: span per cleanup cycle, logged at INFO
- Metrics:
  - `sensor_collection_total` (counter by sensor name and status: success/failure)
  - `sensor_collection_duration_seconds` (histogram)
  - `alerts_triggered_total` (counter by sensor name)
  - `active_websocket_connections` (gauge)

### Background Tasks

- Periodic collection goroutine creates a new root span per cycle (no parent HTTP request)
- Cleanup goroutine gets its own root spans
- All logged with appropriate levels (INFO for routine, WARN for retryable failures, ERROR for critical)

### WebSocket (`ws/`)

- Span per message broadcast
- Gauge metric for active connection count
- Connection/disconnection logged at DEBUG

### SMTP (`smtp/`)

- Span per email send attempt
- Success/failure logged at INFO/ERROR with recipient and subject context

### OAuth (`oauth/`)

- Span per OAuth flow step
- Token refresh logged at DEBUG, failures at ERROR

## Configuration

### application.properties

New entry:
```properties
log.level=info
```

Valid values: `debug`, `info`, `warn`, `error`. Default: `info`.

### packaging/environment

New commented examples:
```bash
# OpenTelemetry Collector endpoint (uncomment to enable telemetry export)
# OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# OTLP protocol: grpc (default, port 4317) or http/protobuf (port 4318)
# OTEL_EXPORTER_OTLP_PROTOCOL=grpc

# Override the default service name (default: sensor-hub)
# OTEL_SERVICE_NAME=sensor-hub
```

The OTel SDK natively reads these env vars — no custom parsing needed.
TLS is handled by `OTEL_EXPORTER_OTLP_CERTIFICATE` or using `https://` in the endpoint URL.

### Behaviour

| Collector configured? | Logs | Traces | Metrics |
|---|---|---|---|
| No | JSON to file + stdout | Discarded (noop) | `/metrics` endpoint only |
| Yes | JSON to file + stdout + OTel export | Exported to collector | `/metrics` + OTel export |

Existing deployments with no `OTEL_*` vars continue working unchanged, just with better structured log output.

## Prometheus `/metrics` Endpoint

Available at `GET /metrics` on port 8080 (same as the API server).
No authentication required (standard for Prometheus scraping).

### Metrics Reference

| Metric | Type | Labels | Description |
|---|---|---|---|
| `http_requests_total` | Counter | method, path, status | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | method, path, status | Request latency |
| `db_query_duration_seconds` | Histogram | operation | Database query latency |
| `sensor_collection_total` | Counter | sensor, status | Sensor reading collection attempts |
| `sensor_collection_duration_seconds` | Histogram | sensor | Sensor collection latency |
| `alerts_triggered_total` | Counter | sensor | Alert trigger count |
| `active_websocket_connections` | Gauge | — | Current WebSocket connections |
| `go_*` | Various | — | Go runtime metrics (goroutines, memory, GC) |

## Logging Audit

A dedicated final phase systematically reviews every Go package and the React UI:

### Go Backend Audit

For each package (`db`, `service`, `alerting`, `api`, `ws`, `oauth`, `smtp`, `notifications`, `cleanup`):

1. Replace every `log.Printf`/`log.Println` with the appropriate `slog.Info`/`slog.Warn`/`slog.Error`/`slog.Debug`
2. Ensure every error path has a log at WARN or ERROR with the error value as a structured attribute
3. Ensure every significant operation (sensor added, reading collected, alert triggered, user created) logs at INFO
4. Add DEBUG logging for routine operations (periodic ticks, cache hits, health checks passing)
5. Add structured context to every log: component name, relevant entity IDs/names, trace IDs where applicable
6. Verify no sensitive data is logged (passwords, API key values, session tokens)

### React UI Audit

1. Create a `logger.ts` utility that respects a debug flag (silent in production, verbose in development)
2. Replace all `console.log` and `console.debug` calls with the logger utility
3. Keep `console.error` and `console.warn` for genuinely unexpected errors
4. Remove the verbose `[Client]` debug logging from `Client.ts` in production mode

## Dependencies

New Go dependencies:
- `go.opentelemetry.io/otel` — Core OTel API
- `go.opentelemetry.io/otel/sdk` — OTel SDK (TracerProvider, MeterProvider)
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` — OTLP trace exporter
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc` — OTLP metric exporter
- `go.opentelemetry.io/otel/exporters/prometheus` — Prometheus metric exporter
- `go.opentelemetry.io/contrib/bridges/otelslog` — slog-to-OTel log bridge
- `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` — Gin middleware
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` — HTTP client tracing (for sensor collection)
- `github.com/XSAM/otelsql` — Database SQL tracing

No new React dependencies.

## Documentation

New Docusaurus page `docs/docs/telemetry.md`:
1. Overview — what telemetry is collected
2. Configuration — log level, OTel env vars
3. Local-only mode — JSON logs, Prometheus endpoint
4. Exporting to a collector — endpoint setup, gRPC vs HTTP, TLS
5. Grafana/Prometheus integration — example scrape config, dashboard queries
6. Log format reference — JSON schema, available fields, trace ID correlation
7. Metrics reference — table of all metrics with descriptions
8. Troubleshooting — verifying export, checking collector connectivity
