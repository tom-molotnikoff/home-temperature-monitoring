---
id: configuration
title: Configuration Settings
sidebar_position: 10
---

# Configuration Settings

Sensor Hub is configured through property files and environment variables. Properties can be updated at runtime through the web UI or the REST API.

## Configuration files

Configuration files are located in the `sensor_hub/configuration/` directory. There are three files:

| File                     | Purpose                                                                  |
|--------------------------|--------------------------------------------------------------------------|
| `application.properties` | Application behavior, sensor polling, authentication, and OAuth settings |
| `database.properties`    | Database connection details                                              |
| `smtp.properties`        | Email sender configuration                                               |

Files use a simple `KEY=VALUE` format, one property per line.

## Runtime configuration updates

Properties can be updated at runtime through the Properties page in the web UI or via the `PATCH /api/properties` API endpoint. Runtime updates are:

- Applied immediately in memory
- Saved back to the configuration files asynchronously
- Broadcast to all connected UI clients via WebSocket

This means changes take effect without restarting the services.

## Application properties

| Property                               | Default                          | Description                                                                        |
|----------------------------------------|----------------------------------|------------------------------------------------------------------------------------|
| `sensor.collection.interval`           | `300`                            | Seconds between sensor polling cycles                                              |
| `sensor.discovery.skip`                | `true`                           | When set to `true`, skips automatic sensor discovery on startup                    |
| `health.history.retention.days`        | `180`                            | Number of days to retain sensor health history records                             |
| `sensor.data.retention.days`           | `365`                            | Number of days to retain temperature reading data                                  |
| `data.cleanup.interval.hours`          | `24`                             | Hours between data cleanup runs                                                    |
| `auth.bcrypt.cost`                     | `12`                             | Bcrypt cost factor for password hashing (higher values are more secure but slower) |
| `auth.session.ttl.minutes`             | `43200`                          | Session duration in minutes (default is 30 days)                                   |
| `auth.session.cookie.name`             | `sensor_hub_session`             | Name of the session cookie                                                         |
| `auth.login.backoff.window.minutes`    | `15`                             | Time window in minutes for counting failed login attempts                          |
| `auth.login.backoff.threshold`         | `5`                              | Number of failed login attempts before backoff is applied                          |
| `auth.login.backoff.base.seconds`      | `2`                              | Base duration in seconds for exponential login backoff                             |
| `auth.login.backoff.max.seconds`       | `300`                            | Maximum backoff duration in seconds                                                |
| `oauth.credentials.file.path`          | `configuration/credentials.json` | Path to the Google OAuth credentials file                                          |
| `oauth.token.file.path`                | `configuration/token.json`       | Path to the stored OAuth token file                                                |
| `oauth.token.refresh.interval.minutes` | `30`                             | Interval in minutes for background OAuth token refresh                             |

## Database properties

| Property        | Type   | Default             | Description                    |
|-----------------|--------|---------------------|--------------------------------|
| `database.path` | string | `data/sensor_hub.db` | Path to the SQLite database file |

## SMTP properties

| Property    | Description                                              |
|-------------|----------------------------------------------------------|
| `smtp.user` | Gmail address used as the sender for email notifications |

## Environment variables

The following environment variables are used by the Docker Compose deployment and override their corresponding properties when set.

| Variable                    | Description                                                                                   |
|-----------------------------|-----------------------------------------------------------------------------------------------|
| `TLS_CERT_FILE`             | Path to the TLS certificate file; when set alongside `TLS_KEY_FILE`, the backend serves HTTPS |
| `TLS_KEY_FILE`              | Path to the TLS private key file                                                              |
| `SENSOR_HUB_ALLOWED_ORIGIN` | The allowed CORS origin for the web UI (e.g., `https://home.sensor-hub:3443`)                 |
| `SENSOR_HUB_INITIAL_ADMIN`  | Creates an initial admin user on first startup; format is `username:password`                 |

## Sensitive properties

There are currently no sensitive database properties. When updating properties through the API, all values are stored as provided.
