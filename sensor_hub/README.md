# Sensor Hub Application

This application also includes a Go component that aggregates readings from all sensors defined in an `openapi.yaml` file. These readings are stored in a MySQL database (separately running), and alerts are triggered when necessary.

## Configuration

- **Required:**

  - `database.properties` — Specify MySQL username, hostname, port, and password.

- **Optional:**
  - `smtp.properties` — Configure SMTP settings to enable email alerts for threshold temperatures.
  - `application.properties` — Optionally set threshold temperatures for alerts.

## Features

- Aggregates readings from multiple sensors
- Stores data in a MySQL database
- Sends alerts when temperature thresholds are exceeded (if SMTP configured)

## Dockerfile

There is an included Dockerfile for running Sensor Hub in a containerised form. To run:

```sh
docker build -f sensor-hub.dockerfile -t sensor-hub .
docker run sensor-hub
```
