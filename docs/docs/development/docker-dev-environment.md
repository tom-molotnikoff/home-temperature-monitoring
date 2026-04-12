# Docker Dev Environment

The Docker Compose setup provides a complete development environment with
hot-reload, remote debugging, mock temperature sensors, and full
OpenTelemetry observability via Grafana.

## Start the Environment

```bash
cd sensor_hub
docker compose -f docker_tests/docker-compose.yml up --build
```

Or use the helper script (also copies the sensor OpenAPI spec):

```bash
cd sensor_hub/docker_tests
./run_sensor_hub_docker.sh
```

## Services

| Service                    | Port(s)        | Description                                    |
|----------------------------|----------------|------------------------------------------------|
| **sensor-hub**             | 8080, 2345, 1883 | Go backend with Air hot-reload, Delve debug, and embedded MQTT broker |
| **sensor-hub-ui**          | 3000 → 5173    | Vite React dev server with HMR                 |
| **mock-sensor-downstairs** | 5001 → 5000    | Python Flask mock HTTP sensor (pull model)     |
| **mock-sensor-upstairs**   | 5002 → 5000    | Python Flask mock HTTP sensor (pull model)     |
| **mock-mqtt-sensor**       | —              | Simulated Zigbee2MQTT devices (push model)     |
| **grafana-lgtm**           | 4000, 4317, 4318 | Grafana + Loki + Tempo + Prometheus (all-in-one) |

### Named Volumes

| Volume              | Purpose                                           |
|---------------------|---------------------------------------------------|
| `sensor-hub-data`   | Persists the SQLite database across restarts       |
| `grafana-data`      | Persists Grafana dashboards and collected telemetry |

Use `docker compose down` to stop containers while keeping data.
Add `--volumes` to also wipe persisted data.

## Grafana — Observability Stack

The `grafana/otel-lgtm` container bundles the full Grafana observability
stack in a single image:

- **Grafana** — dashboards and alerting (port **4000**)
- **Loki** — log aggregation
- **Tempo** — distributed tracing
- **Prometheus** — metrics collection

The sensor-hub container is pre-configured to export all OpenTelemetry data
(logs, traces, and metrics) to the Grafana LGTM collector via gRPC on port
4317. No additional application configuration is needed.

### Accessing Grafana

Open [http://localhost:4000](http://localhost:4000) in your browser. The
default credentials are **admin / admin** (skip the password change prompt
for local development).

### Pre-Built Dashboard

A **Sensor Hub Overview** dashboard is automatically provisioned with the
following panels:

| Section        | Panels                                                    |
|----------------|-----------------------------------------------------------|
| HTTP Overview  | Request rate, avg response time, 5xx error rate, goroutines |
| HTTP Detail    | Request rate by route, p95 latency by route               |
| Database       | Query rate, p95 query duration                            |
| Go Runtime     | Memory usage (RSS + heap), goroutines, GC rate            |
| Logs           | Live application log stream from Loki                     |

The dashboard is loaded from `docker_tests/grafana/dashboards/` and can be
edited in the Grafana UI. Changes made in the UI are saved to the
`grafana-data` volume.

### Exploring Traces

Navigate to **Explore → Tempo** in Grafana to search traces. Every HTTP
request and database query produces a trace span. Click on a trace to see
the full call tree including SQL queries.

### Viewing Logs

Navigate to **Explore → Loki** and query with:

```
{service_name="sensor-hub"} | json
```

Logs are structured JSON with fields like `level`, `msg`, `component`,
`trace_id`, and `span_id`. You can click a `trace_id` value to jump
directly to the corresponding trace in Tempo.

## Air — Go Hot-Reload

[Air](https://github.com/air-verse/air) watches for Go file changes and
rebuilds automatically. The config lives at `docker_tests/air.toml`:

- **Root**: `/app` (the mounted `sensor_hub/` directory)
- **Excludes**: the `ui/` directory
- **Full binary**: launches Delve instead of the bare binary, enabling
  hot-reload and remote debugging simultaneously

## Delve — Remote Debugging

The Air config starts the Go binary through Delve in headless mode:

```
dlv debug --headless --continue --listen=:2345 --api-version=2 --accept-multiclient
```

Connect your IDE debugger to `localhost:2345` (DAP / Delve API v2). The
`--continue` flag means the application starts immediately without waiting for a
debugger to attach.

## Vite — React HMR

The UI container runs `npm run dev` with polling enabled for Docker
compatibility. The Vite dev server inside the container listens on port 5173,
mapped to **localhost:3000** on the host.

## Mock Sensors

The dev stack includes mock sensors for both data collection models.

### HTTP Sensors (Pull Model)

Two Python Flask containers (`mock-sensor-downstairs`, `mock-sensor-upstairs`)
simulate HTTP temperature sensors. Each returns random temperature readings on
port 5000 inside the container, exposed as ports 5001 and 5002 on the host.
Register them in the Sensor Hub UI to test the full pull-based pipeline.

### MQTT Sensor (Push Model)

The `mock-mqtt-sensor` container simulates three Zigbee2MQTT devices that
publish to the embedded MQTT broker inside sensor-hub every 5 seconds:

| Device | Topic | Measurements |
|--------|-------|--------------|
| `living-room-sensor` | `zigbee2mqtt/living-room-sensor` | temperature, humidity, battery, linkquality |
| `front-door` | `zigbee2mqtt/front-door` | contact (binary), battery |
| `office-plug` | `zigbee2mqtt/office-plug` | power, energy, current, state (binary) |

Values drift randomly within realistic ranges. The contact sensor and smart
plug toggle state occasionally to produce interesting binary data.

#### Setting Up MQTT Ingest

The mock sensor starts publishing immediately, but Sensor Hub won't process
the messages until you create a broker record and subscription. After the
stack is running:

**1. Log in and create a broker record pointing at the embedded broker:**

```bash
curl -s -c cookies.txt -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"adminpassword"}'

CSRF=$(curl -s -b cookies.txt http://localhost:8080/api/me | python3 -c "import sys,json; print(json.load(sys.stdin)['csrf_token'])")

curl -s -b cookies.txt -X POST http://localhost:8080/api/mqtt/brokers \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"embedded","type":"embedded","host":"localhost","port":1883,"enabled":true}'
```

**2. Create a subscription that routes zigbee2mqtt topics to the driver:**

```bash
curl -s -b cookies.txt -X POST http://localhost:8080/api/mqtt/subscriptions \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"broker_id":1,"topic_pattern":"zigbee2mqtt/#","driver_type":"mqtt-zigbee2mqtt","enabled":true}'
```

**3. Check that devices were auto-discovered as pending sensors:**

```bash
curl -s -b cookies.txt http://localhost:8080/api/sensors/status/pending | python3 -m json.tool
```

You should see `living-room-sensor`, `front-door`, and `office-plug` listed
as pending. Approve them via the UI or API to start recording readings.

#### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `MQTT_BROKER_HOST` | `sensor-hub` | Hostname of the MQTT broker to publish to |
| `MQTT_BROKER_PORT` | `1883` | MQTT broker port |
| `PUBLISH_INTERVAL` | `5` | Seconds between publish cycles |

## Environment Variables

| Variable                       | Service        | Default                            | Purpose                                |
|--------------------------------|----------------|------------------------------------|----------------------------------------|
| `SENSOR_HUB_INITIAL_ADMIN`    | sensor-hub     | `admin:adminpassword`              | Bootstrap admin user (user:password)   |
| `SENSOR_HUB_ALLOWED_ORIGIN`   | sensor-hub     | `http://localhost:3000`            | CORS allowed origin                    |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | sensor-hub     | `http://grafana-lgtm:4317`        | OTel collector endpoint (gRPC)         |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | sensor-hub     | `grpc`                             | OTel export protocol                   |
| `OTEL_SERVICE_NAME`           | sensor-hub     | `sensor-hub`                       | Service name in telemetry data         |
| `VITE_API_BASE`               | sensor-hub-ui  | `http://localhost:8080/api`        | Backend API URL for the React app      |
| `VITE_WEBSOCKET_BASE`         | sensor-hub-ui  | `ws://localhost:8080/api`          | WebSocket URL for the React app        |
| `HMR_POLLING`                 | sensor-hub-ui  | `true`                             | Enable Vite HMR polling in Docker      |
| `HMR_POLLING_INTERVAL`        | sensor-hub-ui  | `250`                              | Polling interval in milliseconds       |
