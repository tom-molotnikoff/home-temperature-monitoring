---
id: properties-and-oauth
title: Properties and OAuth
sidebar_position: 5
---

# Properties and OAuth API

Endpoints for managing system configuration properties, OAuth settings, and checking service health. All endpoints except `/health` require authentication and the specified permission.

---

## Properties endpoints

### GET /properties

Get all system configuration properties.

Permission: `view_properties`

Sensitive values (such as `database.password`) are masked as `*****` in the response.

#### Response (200 OK)

```json
{
  "sensor.collection.interval": "300",
  "sensor.discovery.skip": "true",
  "health.history.retention.days": "180",
  "sensor.data.retention.days": "365",
  "data.cleanup.interval.hours": "24",
  "auth.bcrypt.cost": "12",
  "auth.session.ttl.minutes": "43200",
  "auth.session.cookie.name": "sensor_hub_session",
  "auth.login.backoff.window.minutes": "15",
  "auth.login.backoff.threshold": "5",
  "auth.login.backoff.base.seconds": "2",
  "auth.login.backoff.max.seconds": "300",
  "database.username": "root",
  "database.password": "*****",
  "database.hostname": "mysql",
  "database.port": "3306",
  "smtp.user": "alerts@example.com",
  "oauth.credentials.file.path": "configuration/credentials.json",
  "oauth.token.file.path": "configuration/token.json",
  "oauth.token.refresh.interval.minutes": "30"
}
```

---

### PATCH /properties

Update one or more configuration properties.

Permission: `manage_properties`

Changes are applied immediately in memory, saved to the configuration files asynchronously, and broadcast to all connected WebSocket clients.

To leave a sensitive property unchanged, send its masked value (`*****`).

#### Request body

```json
{
  "sensor.collection.interval": "600",
  "database.password": "*****"
}
```

In this example, the collection interval is changed to 600 seconds and the database password is left unchanged.

#### Response (202 Accepted)

```json
{
  "message": "Property updated successfully"
}
```

---

### WebSocket: properties updates

`GET /properties/ws`

Permission: `view_properties`

Streams property changes in real time. When any property is updated through the API or UI, the complete set of current properties is broadcast to all connected clients.

#### Message format

Same format as the `GET /properties` response.

---

## OAuth endpoints

These endpoints manage the Google OAuth 2.0 integration used for sending email notifications via Gmail SMTP.

### GET /oauth/status

Get the current OAuth configuration status.

Permission: `manage_oauth`

#### Response (200 OK)

```json
{
  "configured": true,
  "ready": true
}
```

| Field        | Description                                                 |
|--------------|-------------------------------------------------------------|
| `configured` | Whether OAuth credentials are present                       |
| `ready`      | Whether a valid token exists and the system can send emails |

---

### GET /oauth/authorize

Get the Google OAuth authorization URL. Direct the user to this URL to begin the authorization flow.

Permission: `manage_oauth`

#### Response (200 OK)

```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=urn:ietf:wg:oauth:2.0:oob&scope=https://mail.google.com/&response_type=code&state=abc123",
  "state": "abc123"
}
```

The `state` value must be included when submitting the authorization code to prevent CSRF attacks.

---

### POST /oauth/submit-code

Submit the authorization code received from Google after the user completes the OAuth consent flow.

Permission: `manage_oauth`

#### Request body

```json
{
  "code": "4/0AX4XfWhK...",
  "state": "abc123"
}
```

The `state` value must match the one returned by `/oauth/authorize`.

#### Response (200 OK)

```json
{
  "message": "OAuth authorization successful"
}
```

The token is saved to the configured token file path and the system begins refreshing it automatically.

---

### POST /oauth/reload

Reload OAuth credentials and token from disk. Use this if you have manually updated the credential or token files.

Permission: `manage_oauth`

#### Response (200 OK)

```json
{
  "message": "OAuth configuration reloaded"
}
```

---

## Health endpoint

### GET /health

Check that the service is running. This endpoint does not require authentication and is suitable for use as a Docker health check or load balancer probe.

#### Response (200 OK)

```json
{
  "status": "ok"
}
```
