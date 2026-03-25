# Docker Dev Environment

The Docker Compose setup provides a complete development environment with
hot-reload, remote debugging, and mock temperature sensors.

## Start the Environment

```bash
cd sensor_hub
docker compose -f docker_tests/docker-compose.yml up --build
```

## Services

| Service                  | Port(s)          | Description                                    |
|--------------------------|------------------|------------------------------------------------|
| **sensor-hub**           | 8080, 2345       | Go backend with Air hot-reload and Delve debug |
| **sensor-hub-ui**        | 3000 → 5173      | Vite React dev server with HMR                 |
| **mock-sensor-downstairs** | 5001 → 5000    | Python Flask mock sensor                       |
| **mock-sensor-upstairs**   | 5002 → 5000    | Python Flask mock sensor                       |

A named volume `sensor-hub-data` persists the SQLite database across restarts.

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

Two Python Flask containers simulate temperature sensors. Each returns random
temperature readings on port 5000 inside the container, exposed as ports 5001
and 5002 on the host. Register them in the Sensor Hub UI to test the full
pipeline.

## Environment Variables

| Variable                       | Service        | Default                            | Purpose                                |
|--------------------------------|----------------|------------------------------------|----------------------------------------|
| `SENSOR_HUB_INITIAL_ADMIN`    | sensor-hub     | `admin:adminpassword`              | Bootstrap admin user (user:password)   |
| `SENSOR_HUB_ALLOWED_ORIGIN`   | sensor-hub     | `http://localhost:3000`            | CORS allowed origin                    |
| `VITE_API_BASE`               | sensor-hub-ui  | `http://localhost:8080/api`        | Backend API URL for the React app      |
| `VITE_WEBSOCKET_BASE`         | sensor-hub-ui  | `ws://localhost:8080/api`          | WebSocket URL for the React app        |
| `HMR_POLLING`                 | sensor-hub-ui  | `true`                             | Enable Vite HMR polling in Docker      |
| `HMR_POLLING_INTERVAL`        | sensor-hub-ui  | `250`                              | Polling interval in milliseconds       |
