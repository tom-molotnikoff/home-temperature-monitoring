# Telemetry Unification & Distributed Tracing Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Push all metrics through OTLP (eliminating the Prometheus scrape requirement) and add distributed tracing for sensor HTTP calls — both Go client-side and Flask server-side.

**Architecture:** Replace `prometheus/client_golang` default collectors with `go.opentelemetry.io/contrib/instrumentation/runtime` for Go runtime metrics via OTLP. Wrap the sensor HTTP client with `otelhttp` to create traced outgoing requests. Instrument both the mock and real Flask sensors with `opentelemetry-instrumentation-flask` to participate in distributed traces. The `/metrics` endpoint stays for manual debugging but Prometheus scraping is removed from the dev environment.

**Tech Stack:** Go OTel SDK, otelhttp, OTel runtime instrumentation, Python opentelemetry-instrumentation-flask, Grafana LGTM

---

### Task 1: Add OTel Go runtime instrumentation

**Files:**
- Modify: `sensor_hub/go.mod` (add dependency)
- Modify: `sensor_hub/telemetry/telemetry.go` (start runtime instrumentation)

**Step 1: Add the runtime instrumentation dependency**

```bash
cd sensor_hub && go get go.opentelemetry.io/contrib/instrumentation/runtime
```

**Step 2: Start runtime instrumentation in `telemetry.go`**

After the meter provider is initialised (line ~85), add:

```go
import "go.opentelemetry.io/contrib/instrumentation/runtime"

// Inside Init(), after initMeterProvider:
if err := runtime.Start(); err != nil {
    return nil, fmt.Errorf("failed to start runtime instrumentation: %w", err)
}
```

This produces OTel-native Go runtime metrics (`go.goroutine.count`, `go.memory.used`, `go.memory.allocated`, `go.schedule.duration`) through the MeterProvider, which routes them to both the OTLP exporter and the Prometheus exporter.

**Step 3: Verify build**

```bash
cd sensor_hub && go build ./...
```

**Step 4: Run tests**

```bash
cd sensor_hub && go test ./telemetry/...
```

---

### Task 2: Remove Prometheus scrape config from Docker environment

**Files:**
- Delete: `sensor_hub/docker_tests/grafana/prometheus.yaml`
- Modify: `sensor_hub/docker_tests/docker-compose.yml` (remove scrape config mount, remove `OTEL_SEMCONV_STABILITY_OPT_IN`)

**Step 1: Remove the Prometheus scrape config volume mount from docker-compose.yml**

In the `grafana-lgtm` service volumes, remove the line:
```yaml
- ./grafana/prometheus.yaml:/otel-lgtm/prometheus.yaml
```

**Step 2: Remove `OTEL_SEMCONV_STABILITY_OPT_IN` from sensor-hub environment**

This was a temporary workaround for the Prometheus scrape path. With OTLP-only, otelsql metrics go through the OTel MeterProvider directly — no Prometheus naming conversion needed. The env var can stay if we want the newer semantic convention names at the `/metrics` debug endpoint, but it's not required for OTLP. **Keep it** for consistency.

**Step 3: Delete `sensor_hub/docker_tests/grafana/prometheus.yaml`**

```bash
rm sensor_hub/docker_tests/grafana/prometheus.yaml
```

---

### Task 3: Update Grafana dashboard to use OTel metric names

**Files:**
- Modify: `sensor_hub/docker_tests/grafana/dashboards/sensor-hub-overview.json`

The Go runtime panels currently query Prometheus-native metric names. After switching to OTel runtime instrumentation, the metrics arrive via OTLP with different names. In the LGTM Prometheus, OTLP metrics use dots-to-underscores conversion with unit suffixes.

**Metric name mapping:**

| Old (Prometheus native)             | New (OTel runtime via OTLP)              |
|--------------------------------------|------------------------------------------|
| `go_goroutines`                      | `go_goroutine_count`                     |
| `process_resident_memory_bytes`      | (remove — use `go_memory_used_bytes`)    |
| `go_memstats_heap_alloc_bytes`       | `go_memory_used_bytes`                   |
| `go_gc_duration_seconds_count`       | `go_memory_allocated_bytes_total` (rate) |

**Step 1: Update panel ID 4 (Active Go Routines)**

Change expr from `go_goroutines{job="sensor-hub"}` to `go_goroutine_count{job="sensor-hub"}`

**Step 2: Update panel ID 9 (Memory Usage)**

- Target A: Change from `process_resident_memory_bytes{job="sensor-hub"}` to `go_memory_used_bytes{job="sensor-hub"}` with legend "Heap Used"
- Target B: Change from `go_memstats_heap_alloc_bytes{job="sensor-hub"}` to `go_memory_allocated_bytes_total{job="sensor-hub"}` with legend "Total Allocated"

**Step 3: Update panel ID 10 (Goroutines & GC)**

- Target A: Change from `go_goroutines{job="sensor-hub"}` to `go_goroutine_count{job="sensor-hub"}`
- Target B: Change from `rate(go_gc_duration_seconds_count{job="sensor-hub"}[5m])` to `rate(go_memory_allocated_bytes_total{job="sensor-hub"}[5m])` with legend "Allocation Rate"

**Note:** The exact metric names must be verified after Task 1 by checking `curl localhost:8080/metrics` or Grafana Explore → Prometheus. The OTel runtime package metric names may vary by version. Adjust accordingly.
### Task 4: Instrument Go HTTP client with otelhttp for sensor calls

**Files:**
- Modify: `sensor_hub/go.mod` (add otelhttp dependency)
- Modify: `sensor_hub/service/sensor_service.go` (use instrumented HTTP client)

**Step 1: Add the otelhttp dependency**

```bash
cd sensor_hub && go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

**Step 2: Add an instrumented HTTP client to SensorService**

In `sensor_service.go`, add a package-level or struct-level HTTP client:

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// In SensorService struct, add:
httpClient *http.Client

// In NewSensorService, add:
httpClient: &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
    Timeout:   10 * time.Second,
},
```

**Step 3: Replace `http.Get` with context-aware instrumented call**

In `ServiceFetchTemperatureReadingFromSensor` (line ~287), replace:

```go
response, err := http.Get(sensor.URL + "/temperature")
```

With:

```go
req, err := http.NewRequestWithContext(ctx, "GET", sensor.URL+"/temperature", nil)
if err != nil {
    return tempReading, fmt.Errorf("error creating request to sensor at %s: %w", sensor.URL, err)
}
response, err := s.httpClient.Do(req)
```

This creates a child span for each outgoing sensor HTTP call, showing:
- Full round-trip duration
- HTTP method, URL, status code
- Trace context propagation (W3C `traceparent` header) to the sensor

**Step 4: Build and run tests**

```bash
cd sensor_hub && go build ./... && go test ./service/...
```

---

### Task 5: Instrument mock Flask sensors with OpenTelemetry

**Files:**
- Modify: `sensor_hub/docker_tests/requirements.txt` (add OTel packages)
- Modify: `sensor_hub/docker_tests/mock-temperature-sensor.py` (add OTel auto-instrumentation)
- Modify: `sensor_hub/docker_tests/mock-sensor.dockerfile` (no changes needed if using auto-instrument)

**Step 1: Add OTel packages to `sensor_hub/docker_tests/requirements.txt`**

Append these lines:

```
opentelemetry-distro==0.52b0
opentelemetry-exporter-otlp==1.31.0
opentelemetry-instrumentation-flask==0.52b0
```

**Step 2: Add OTel instrumentation to `mock-temperature-sensor.py`**

Add at the top of the file, before the Flask app is created:

```python
import os

# OpenTelemetry auto-instrumentation
from opentelemetry import trace
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

def init_telemetry():
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT")
    if not endpoint:
        return
    resource = Resource.create({
        "service.name": os.environ.get("OTEL_SERVICE_NAME", "mock-sensor")
    })
    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)

init_telemetry()

app = Flask(__name__)
FlaskInstrumentor().instrument_app(app)
```

This reads the incoming `traceparent` header from the sensor-hub's otelhttp client, creates a child span for the Flask request handling, and exports it to the same OTLP endpoint.

**Step 3: Verify mock sensor still works standalone**

```bash
cd sensor_hub/docker_tests && python -c "import flask; print('Flask OK')"
```

---

### Task 6: Instrument real Flask sensor with OpenTelemetry

**Files:**
- Modify: `temperature_sensor/requirements.txt` (add OTel packages)
- Modify: `temperature_sensor/sensor_api.py` (add OTel auto-instrumentation)

**Step 1: Add OTel packages to `temperature_sensor/requirements.txt`**

Append the same OTel packages as the mock sensor:

```
opentelemetry-distro==0.52b0
opentelemetry-exporter-otlp==1.31.0
opentelemetry-instrumentation-flask==0.52b0
```

**Step 2: Add OTel instrumentation to `sensor_api.py`**

Add the same telemetry init block before the Flask app. Use `OTEL_SERVICE_NAME` defaulting to the hostname (matching current sensor naming convention):

```python
import os
import socket
from opentelemetry import trace
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

def init_telemetry():
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT")
    if not endpoint:
        return
    resource = Resource.create({
        "service.name": os.environ.get("OTEL_SERVICE_NAME", socket.gethostname())
    })
    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)

init_telemetry()
```

Then after creating the Flask app, add:

```python
FlaskInstrumentor().instrument_app(app)
```

**Note:** The real sensor OTel export is opt-in — if `OTEL_EXPORTER_OTLP_ENDPOINT` is not set, no tracing overhead is added.

---

### Task 7: Update docker-compose for mock sensor OTel export

**Files:**
- Modify: `sensor_hub/docker_tests/docker-compose.yml` (add OTel env vars to mock sensors)

**Step 1: Add OTel environment variables to both mock sensor services**

For `mock-sensor-downstairs`:
```yaml
environment:
  OTEL_EXPORTER_OTLP_ENDPOINT: http://grafana-lgtm:4317
  OTEL_SERVICE_NAME: mock-sensor-downstairs
```

For `mock-sensor-upstairs`:
```yaml
environment:
  OTEL_EXPORTER_OTLP_ENDPOINT: http://grafana-lgtm:4317
  OTEL_SERVICE_NAME: mock-sensor-upstairs
```

**Step 2: Add `depends_on: grafana-lgtm` to mock sensors**

So they don't try to export traces before the collector is ready.

---

### Task 8: Build, test, and verify end-to-end

**Step 1: Run Go tests**

```bash
cd sensor_hub && go build ./... && go test ./...
```

**Step 2: Rebuild Docker environment**

```bash
cd sensor_hub/docker_tests && docker compose down && docker compose up --build -d
```

**Step 3: Verify OTel runtime metrics appear**

```bash
curl -s localhost:8080/metrics | grep go_goroutine_count
curl -s localhost:8080/metrics | grep go_memory
```

**Step 4: Verify distributed traces in Grafana**

1. Open http://localhost:4000 (Grafana)
2. Go to Explore → Tempo
3. Search for traces from `sensor-hub` service
4. Find a sensor reading trace — it should show:
   - Parent span: `GET /api/...` (sensor-hub API)
   - Child span: `GET` to sensor URL (otelhttp client)
   - Grandchild span: `GET /temperature` (Flask sensor)

**Step 5: Verify dashboard panels populate**

1. Open the Sensor Hub Overview dashboard
2. Goroutines panel should show data from `go_goroutine_count`
3. Memory panel should show data from `go_memory_used_bytes`
4. DB panels should show data from `db_client_operation_duration_seconds`

**Step 6: If metric names differ from plan, update dashboard accordingly**

The OTel runtime metric names depend on the exact package version. Check `curl localhost:8080/metrics` for actual names and update the dashboard JSON if needed.
