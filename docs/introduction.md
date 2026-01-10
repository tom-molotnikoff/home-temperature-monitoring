# Home Temperature Monitoring — Introduction

One-line summary
----------------
Home Temperature Monitoring — a small self-hosted sensor hub for collecting, storing and visualising temperature sensor data (REST + WebSocket API, MySQL-backed with a web UI).

Purpose & high-level features
-----------------------------
This project provides a backend "sensor hub" that collects temperature (and related) readings from sensors, stores readings and health history in a database, and exposes:
- a REST API (OpenAPI specification at `sensor_hub/openapi.yaml`) for reading sensors, readings, and properties,
- a WebSocket channel for real-time sensor updates (`sensor_hub/ws/hub.go` and `sensor_hub/api/websocket.go`),
- a simple UI (`sensor_hub/ui/sensor_hub_ui`) for dashboards and management,
- sensor implementation for temperature sensors - should be rehomed eventually (`temperature_sensor/`).

Key capabilities
- Persistent storage of sensor readings and sensor health history (migrations under `sensor_hub/db/changesets/`).
- Configurable application properties (defaults and files under `sensor_hub/application_properties/` and `sensor_hub/configuration/`).
- Docker-friendly setup (compose and Dockerfiles under `sensor_hub/docker/` and `sensor_hub/docker_tests/`).

Repository layout (important folders and files)
----------------------------------------------
- `sensor_hub/` — Main Go service (backend)
    - `main.go` — service entrypoint.
    - `openapi.yaml` — API specification.
    - `api/` — REST and WebSocket handlers (see `api.go`, `sensorApi.go`, `temperatureApi.go`, `websocket.go`).
    - `ws/` — websocket hub for registering and handling many websockets (`hub.go`).
    - `service/` — business logic layers (sensorService, temperatureService, propertiesService).
    - `db/` — DB helpers and repository interfaces (`sensorRepository.go`, `temperatureRepository.go`).
    - `db/changesets/` — SQL migrations (`V1__init_schema.sql` ... `V5__sensor_health_history.sql`).
    - `application_properties/` — defaults and property file helpers.
    - `configuration/` — example `application.properties`, credentials and other runtime config.
    - `docker/` — Dockerfiles, docker-compose and scripts for running the hub with a DB.
    - `ui/sensor_hub_ui/` — frontend (Vite/React/TS).
    - `integration/` — integration test(s).
- `temperature_sensor/` — sensor emulator and small Flask API; useful for local tests and dev. Contains `requirements.txt`, `sensor_api.py`, and `run-flask-api.sh`.
- `__notes__/todo` — developer TODOs and future improvements.

Quick start — local development (assumptions)
--------------------------------------------
Assumptions:
- Go toolchain installed (see `sensor_hub/go.mod` for Go version).
- Docker and docker-compose for container-based runs (optional but strongly recommended).
- Python 3 + pip for the `temperature_sensor` (optional).
- Node.js + npm for the UI (`sensor_hub/ui/sensor_hub_ui`) if running the SPA locally.

Common workflows

1) Build & run the Go backend (quick)
```bash
# from repo root
cd sensor_hub
# run directly (development)
go run ./main.go

# or build binary
go build -o sensor-hub ./...
./sensor-hub
```
2) Run production application with Docker
```bash
# requires database.properties and application.properties in sensor_hub/configuration/

# from repo root
# start sensor_hub and a database as defined in the repo's docker compose
docker-compose -f sensor_hub/docker/docker-compose.yml up --build
```
3) Run the full application in test mode with fake sensors
```bash
# from repo root
# start sensor_hub, database, and temperature_sensor emulator
docker-compose -f sensor_hub/docker_tests/docker-compose.yml up --build
```
4) Run the UI locally (optional, HMR is already in the docker setup)
```bash
# from repo root
cd sensor_hub/ui/sensor_hub_ui
npm install
npm run dev
```
5) Obtaining credentials.json and token.json for Gmail SMTP (optional, for email alerts)
```bash
# from repo root
cd sensor_hub/pre_authorise_application
go run ./main.go
# follow instructions to get credentials.json and token.json
# put them in sensor_hub/configuration/
```

API & WebSocket overview
-----------------------
API specification is available at sensor_hub/openapi.yaml. Use this as the authoritative contract for REST endpoints.
WebSocket hub is implemented under sensor_hub/ws/hub.go and integrated via handlers in sensor_hub/api/websocket.go — use it to subscribe to real-time sensor updates and events (the OpenAPI file documents REST only; check websocket.go and hub.go for message format and topics)

Database & schema
-----------------
The application uses MySQL for persistent storage.
SQL migrations are in sensor_hub/db/changesets/. These show the schema evolution (initial schema, sensors table, sensor health, disabling sensors, and sensor health history etc).
DB connection, helpers and repository implementations are in sensor_hub/db/ (see db.go, repository interfaces).
Check sensor_hub/docker_tests/ Docker Compose for details on how the DB is configured for local development.

Key files and folders
----------------------
- `sensor_hub/main.go` — program entrypoint and startup wiring.
- `sensor_hub/openapi.yaml` — API contract.
- `sensor_hub/api/` — handlers and API wiring.
- `sensor_hub/ws/hub.go` and sensor_hub/api/websocket.go — WebSocket implementation and usage.
- `sensor_hub/service/` — business logic (sensors, readings, properties).
- `sensor_hub/db/changesets/` — DB schema and migrations.
- `__notes__/todo` — in-repo TODOs and desired features.

Checks to run
-----------------
Exact Go version: check sensor_hub/go.mod for a go directive.
Node.js & npm versions: see sensor_hub/ui/sensor_hub_ui/package.json and toolchain requirements.