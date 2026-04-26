# Home sensor monitoring

A small self hosted setup for collecting, storing and viewing sensor readings from networked devices. Supports pluggable sensor drivers for different hardware (temperature, humidity, power, etc.).

Since it began, it has evolved a bunch and I use this space to experiment with workflows and hardware. 

## Overview

- The backend that aggregates data and serves the UI is in `sensor_hub`. The React UI is embedded into the Go binary at build time (`sensor_hub/web/`), so a single binary serves both the API and the SPA.
- There's a simple Flask app at `temperature_sensor` for a temperature sensor I wired up to a few Pi Zeros. 
- Support for Zigbee devices through Zigbee2Mqtt

## Quick layout

- `sensor_hub` - Go backend, API handlers, DB code, service layer and the embedded UI.
- `temperature_sensor` - tiny Flask app that relies on a DS18B20 temperature sensor.
- `sensor_hub/docker_tests` - dev Docker Compose with mock sensors, a separate Vite dev server for HMR, and Delve for debugging.
- `sensor_hub/db/migrations` - SQL migrations that define the schema used by the hub.
- `packaging/` - systemd unit, logrotate config, scriptlets, and production config defaults for RPM/DEB packages.
- `.goreleaser.yml` - GoReleaser configuration for cross-compilation, packaging, and GPG signing.
- `.github/` - GitHub Actions for CI and releasing.

## Installation

The recommended way to install Sensor Hub is via RPM or DEB package from the [GitHub Releases](https://github.com/tom-molotnikoff/home-temperature-monitoring/releases) page. See the [installation guide](docs/docs/installation.md) for full instructions.

### Quick start (development)

Run the full dev stack with fake sensors and hot-reload:

```bash
docker compose -f sensor_hub/docker_tests/docker-compose.yml up --build
```

### Build packages locally

```bash
./scripts/build-packages.sh
```

Requires [GoReleaser](https://goreleaser.com/install/). Produces unsigned snapshot RPM and DEB packages in `dist/`.

## Documentation

User guide and developer documentation is available in the `docs/` directory. To build and serve locally:

```bash
cd docs
npm install
npm run start
```

![image showing the dashboard of the sensor hub user interface](readme-assets/dashboard.png "Sensor Hub Dashboard")
