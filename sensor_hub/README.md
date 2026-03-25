# Sensor Hub Application

This application is a web app that can collect readings from all sensors defined in an `openapi.yaml` file. These readings are stored in an embedded SQLite database, and alerts are triggered when necessary.

The React UI is embedded into the Go binary via `//go:embed all:dist` in `web/embed.go`. At runtime a single binary serves both the REST/WebSocket API (under `/api`) and the React SPA. In production, nginx sits in front purely as a TLS reverse proxy — all requests are forwarded to the Go process.

## Configuration

- **Required:**

  - `database.properties` — Specify the SQLite database file path.
  - `application.properties` — Specify the location of the sensor's openapi.yaml file. This is used to identify the number of available sensors and their URLs.

- **Optional:**
  - `smtp.properties` — Configure SMTP settings to enable email alerts for threshold temperatures.
  - `application.properties` — Optionally set threshold temperatures for alerts.
  - `credentials.json` — If using SMTP for alerts, an appropriate credentials.json with a redirect uri of <http://localhost:8080> must be provided. This file should be put through the pre-authorisation script to generate a token.json before running sensor-hub. Sensor-hub cannot do the interactive authorisation process.

## Features

- Single binary serves the REST API, WebSocket, and React SPA
- All API routes live under the `/api` prefix
- Stores data in a SQLite database
- Sends alerts when temperature thresholds are exceeded (if SMTP configured)

## Building

Use `scripts/build.sh` to build the React UI and Go binary in one step. The script runs `npm ci && npm run build` in `ui/sensor_hub_ui/`, copies the output to `web/dist/`, and then runs `go build`. The resulting binary has the UI embedded and needs no separate web server.

## Docker compose setup

The `docker/` folder contains a production Docker Compose file. The Dockerfile is a multi-stage build (node → go → alpine) that produces a single container with the Go binary and embedded UI. Nginx is configured as a TLS reverse proxy in front — it forwards all requests to the Go process.

```sh
cd docker
docker compose up -d --build
```

The `docker_tests/` folder has a development setup with mock sensors, Air + Delve for Go hot-reload and debugging, and a separate Vite dev server (`sensor-hub-ui` service) for React HMR.

```sh
cd docker_tests
docker compose up --build
```

## Debug setup

Since the project uses Air and Delve for the debugging of the Go application, you can attach a debugger to dive into the call stack and variables.

Create a launch.json in the .vscode directory with the following contents:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Connect to container",
      "type": "go",
      "debugAdapter": "dlv-dap",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "localhost",
      "trace": "verbose",
      "substitutePath": [
        {
          "from": "/Users/tommolotnikoff/Documents/personal/git/home-temperature-monitoring/sensor_hub",
          "to": "/app"
        }
      ]
    }
  ]
}
```

Replace the path to the "sensor_hub" folder with the appropriate path on the system you are on. Run the debugger and off u go.

Atm, exiting the debugger is closing the container for sensor_hub - to be fixed
