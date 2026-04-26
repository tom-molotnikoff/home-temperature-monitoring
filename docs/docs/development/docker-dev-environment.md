# Docker Dev Environment

The Docker Compose setup provides a complete development environment with
hot-reload, remote debugging, mock temperature sensors, and full
OpenTelemetry observability via Grafana.

## Start the Environment

```bash
cd sensor_hub
docker compose -f docker_tests/docker-compose.yml up --build
```

## Grafana — Observability Stack

The `grafana/otel-lgtm` container bundles the full Grafana observability
stack in a single image.

The sensor-hub container is pre-configured to export all OpenTelemetry data
(logs, traces, and metrics) to the Grafana LGTM collector via gRPC on port
4317. No additional application configuration is needed.

### Accessing Grafana

Open [http://localhost:4000](http://localhost:4000) in your browser. The
default credentials are **admin / admin** (skip the password change prompt
for local development).

## Air — Go Hot-Reload

Go source changes trigger an automatic rebuild and restart of the backend. The UI directory is excluded from file watching.

## Delve — Remote Debugging

Connect your IDE debugger to `localhost:2345` (DAP / Delve API v2). The application starts immediately — attach a debugger at any time without restarting.

## Vite — React HMR

The UI container runs the Vite dev server with hot module replacement. Source changes in `ui/sensor_hub_ui/src/` are reflected immediately in the browser at **localhost:3000**.

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

Values drift randomly within realistic ranges. The contact sensor and smart
plug toggle state occasionally to produce interesting binary data.

#### Setting Up MQTT Ingest

The mock sensor starts publishing immediately, but Sensor Hub won't process
the messages until you create a broker record and subscription. After the
stack is running:

**1. Log in and create a broker record pointing at the embedded broker:**

**2. Create a subscription that routes zigbee2mqtt topics to the driver:**

**3. Check that devices were auto-discovered as pending sensors:**

You should see some sensors listed as pending. Approve them to start recording readings.