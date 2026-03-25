---
id: installation
title: Installation
sidebar_position: 3
---

# Installation

:::note
The installation process is being refined and will improve in future releases. The steps below reflect the current deployment method.
:::

## Clone the repository

```bash
git clone <repo-url>
cd home-temperature-monitoring/sensor_hub
```

## Create configuration files

Create a `configuration/` directory inside `sensor_hub/` and add the following files.

### database.properties

```properties
database.path=data/sensor_hub.db
```

### application.properties

```properties
sensor.collection.interval=300
sensor.discovery.skip=true
health.history.retention.days=180
sensor.data.retention.days=365
data.cleanup.interval.hours=24
auth.bcrypt.cost=12
auth.session.ttl.minutes=43200
auth.session.cookie.name=sensor_hub_session
auth.login.backoff.window.minutes=15
auth.login.backoff.threshold=5
auth.login.backoff.base.seconds=2
auth.login.backoff.max.seconds=300
oauth.credentials.file.path=configuration/credentials.json
oauth.token.file.path=configuration/token.json
oauth.token.refresh.interval.minutes=30
```

See [Configuration Settings](configuration) for a description of each property.

### smtp.properties

```properties
smtp.user=your-email@gmail.com
```

This is the Gmail address used as the sender for email notifications. Only required if you plan to enable email alerts.

## Prepare TLS certificates

For local development with self-signed certificates using mkcert:

```bash
mkcert -install
mkcert home.sensor-hub
```

The default docker-compose configuration expects certificates at `/home/tom/cert/`. Update the volume mount paths in the compose file to match your setup.

## Configure Docker Compose

Edit `sensor_hub/docker/docker-compose.yml` and update the following:

- TLS certificate volume mount paths for the sensor-hub service
- `SENSOR_HUB_ALLOWED_ORIGIN` to match your domain and port (e.g., `https://home.sensor-hub:3443`)

## Set the initial admin user

Add the `SENSOR_HUB_INITIAL_ADMIN` environment variable to the `sensor-hub` service in docker-compose:

```yaml
environment:
  SENSOR_HUB_INITIAL_ADMIN: admin:your_admin_password
```

This creates an admin user with full permissions on first startup, only if no users exist in the database. Remove this variable after the initial deployment.

## Start the services

```bash
cd sensor_hub/docker
docker compose up -d
```

On first start:

1. Embedded migrations create the SQLite database and schema automatically
2. The Go binary starts serving both the API and the React UI
3. Nginx proxies incoming TLS requests to the Go process

## Set up OAuth for email notifications (optional)

Email notifications require Gmail OAuth 2.0 authorization.

### Option 1: Pre-authorization tool

1. Place your `credentials.json` file in `sensor_hub/configuration/`
2. Run the authorization tool:

```bash
cd sensor_hub/pre_authorise_application
go run pre_authorise_application.go
```

3. A browser window opens for Google authorization. Sign in and grant access.
4. The tool saves the token to `sensor_hub/configuration/token.json`

### Option 2: Web UI

After deployment, navigate to the Alerts and Notifications page and use the OAuth Configuration card to authorize Gmail access. This requires the `manage_oauth` permission.

## Verify the installation

1. Open the web UI at `https://<host>:3443` or `http://<host>:3000`
2. Log in with the admin credentials you configured
3. Verify the health endpoint returns a successful response:

```bash
curl http://<host>:8080/api/health
```

Expected response:

```json
{"status": "ok"}
```
