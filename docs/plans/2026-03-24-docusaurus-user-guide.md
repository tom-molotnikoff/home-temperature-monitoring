# Docusaurus User Guide Documentation Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Replace the existing developer-focused markdown docs with a professional Docusaurus user guide covering all aspects of Sensor Hub.

**Architecture:** Initialize a Docusaurus site at `docs/` (replacing the existing markdown files). Each documentation page is a standalone `.md` file in `docs/docs/`. Sidebar navigation is defined in `sidebars.js`. The site is buildable and servable locally with `npm run build`. The existing `docs/plans/` directory is preserved outside the Docusaurus content tree.

**Tech Stack:** Docusaurus 3.x, Node.js, Markdown

---

### Task 1: Initialize Docusaurus project

**Files:**
- Create: `docs/` (Docusaurus root, replaces existing markdown files)
- Preserve: `docs/plans/` (move out temporarily, restore after init)

**Step 1: Back up plans directory**

```bash
cd /home/tommolotnikoff/Documents/repositories/home-temperature-monitoring
mv docs/plans /tmp/sensor-hub-plans-backup
```

**Step 2: Remove old doc files**

```bash
rm docs/alerting-system.md docs/introduction.md docs/mobile-responsiveness.md docs/notification-system.md docs/testing-guide.md
```

**Step 3: Initialize Docusaurus in docs/**

```bash
cd /home/tommolotnikoff/Documents/repositories/home-temperature-monitoring
npx create-docusaurus@latest docs classic --typescript
```

Select defaults when prompted. This creates the Docusaurus scaffold.

**Step 4: Restore plans directory**

```bash
mv /tmp/sensor-hub-plans-backup docs/plans
```

**Step 5: Clean up Docusaurus boilerplate**

Remove the default blog, tutorial docs, and example pages that ship with Docusaurus:
- Delete `docs/blog/`
- Delete `docs/docs/tutorial-*` files
- Delete `docs/src/pages/index.tsx` (replace with redirect to docs)
- Clear `docs/src/components/HomepageFeatures/`

**Step 6: Configure docusaurus.config.ts**

Edit `docs/docusaurus.config.ts`:
- Set `title` to `Sensor Hub`
- Set `tagline` to `Home Temperature Monitoring System`
- Set `url` to `https://sensor-hub.docs` (placeholder)
- Disable blog plugin (`blog: false`)
- Set docs `routeBasePath` to `/` so docs are the landing page
- Update navbar with title "Sensor Hub Documentation"
- Remove social links from footer

**Step 7: Verify scaffold builds**

```bash
cd docs && npm run build
```

Expected: Build succeeds (may warn about missing docs, that is fine).

---

### Task 2: Write Overview page

**Files:**
- Create: `docs/docs/overview.md`

**Step 1: Create the overview page**

This is the landing page. Content must cover:
- What Sensor Hub is (one paragraph)
- Key capabilities (bullet list, no bold overuse): temperature monitoring, alert rules, email and in-app notifications, user management with role-based access control, real-time WebSocket data streaming, REST API, responsive web UI
- Architecture summary: Go backend, React/TypeScript SPA, MySQL database, Flyway migrations, Docker Compose deployment
- System diagram description (text, no image): sensors expose HTTP endpoints, Sensor Hub polls them on a schedule, stores readings in MySQL, serves data via REST API and WebSocket, React UI consumes both
- Link to Prerequisites and Installation pages

Frontmatter:
```yaml
---
id: overview
title: Overview
sidebar_position: 1
slug: /
---
```

**Step 2: Review for tone**

Ensure: no emoji, no overuse of bold, professional and concise. Match Cirata docs tone.

---

### Task 3: Write Prerequisites page

**Files:**
- Create: `docs/docs/prerequisites.md`

**Step 1: Create the prerequisites page**

Content must cover:
- Docker and Docker Compose (required for deployment)
- MySQL 8 (provided by Docker Compose, or bring your own)
- TLS certificates (for production; optional for development)
  - mkcert for local development certificates
  - Certificate and key in PEM format
  - CA certificate for Nginx to verify backend
- Network requirements: ports 8080 (API), 3000/3443 (UI), 3306 (MySQL, internal)
- For deploying temperature sensors:
  - Raspberry Pi (or similar Linux SBC)
  - DS18B20 temperature sensor connected via 1-wire protocol
  - Python 3.11+ with pip and venv
  - Network connectivity between sensor and Sensor Hub host
- Google Cloud project with OAuth 2.0 credentials (for Gmail email notifications, optional)
  - Desktop application type credential
  - Gmail API enabled

Frontmatter:
```yaml
---
id: prerequisites
title: Prerequisites
sidebar_position: 2
---
```

---

### Task 4: Write Installation page

**Files:**
- Create: `docs/docs/installation.md`

**Step 1: Create the installation page**

Note in the page intro that the installation process is being refined and will improve in future releases.

Content must cover:

**Clone the repository:**
```bash
git clone <repo-url>
cd home-temperature-monitoring/sensor_hub
```

**Configuration files:**
- Create `configuration/` directory
- `configuration/database.properties` with keys: `database.username`, `database.password`, `database.hostname`, `database.port`
- `configuration/application.properties` with all keys and their defaults (reference Configuration Settings page)
- `configuration/smtp.properties` with key: `smtp.user` (Gmail address for sending notifications)

**TLS certificates (production):**
- Generate with mkcert or use real certificates
- Place cert and key files where docker-compose volume mounts expect them
- Default paths in docker-compose: `/home/tom/cert/` (must be customised)

**Docker Compose deployment:**
- Edit `sensor_hub/docker/docker-compose.yml`:
  - Update MySQL password
  - Update volume mount paths for TLS certs and MySQL data
  - Update `SENSOR_HUB_ALLOWED_ORIGIN` to match your domain
- Run: `docker compose -f sensor_hub/docker/docker-compose.yml up -d`
- Flyway runs migrations automatically on first start
- Backend starts after migrations complete
- UI served on ports 3000 (HTTP) and 3443 (HTTPS)

**Initial admin user:**
- Set `SENSOR_HUB_INITIAL_ADMIN=username:password` environment variable on the sensor-hub service
- Only creates the admin user if no users exist in the database
- Remove the environment variable after first start (credentials are persisted in the database)

**OAuth setup for email notifications (optional):**
- Create Google Cloud OAuth 2.0 credentials (Desktop application type)
- Save credentials JSON to `configuration/credentials.json`
- Run the pre-authorization tool: `cd sensor_hub/pre_authorise_application && go run pre_authorise_application.go`
- Authorize in browser, token saved to `configuration/token.json`
- Alternatively, configure OAuth through the web UI after deployment

**Verify installation:**
- Access UI at `https://<host>:3443` or `http://<host>:3000`
- Log in with admin credentials
- Check health endpoint: `GET /health` returns `{"status": "ok"}`

Frontmatter:
```yaml
---
id: installation
title: Installation
sidebar_position: 3
---
```

---

### Task 5: Write Upgrading page

**Files:**
- Create: `docs/docs/upgrading.md`

**Step 1: Create the upgrading page**

Content must cover:

**Upgrade process:**
1. Pull the latest changes: `git pull`
2. Rebuild and restart containers: `docker compose -f sensor_hub/docker/docker-compose.yml up -d --build`
3. Flyway automatically applies any new database migrations on startup
4. The backend waits for Flyway to complete before starting

**Database migrations:**
- Migrations are located in `sensor_hub/db/changesets/`
- Flyway tracks applied migrations in the `flyway_schema_history` table
- Migrations are versioned sequentially (V1 through V18 at time of writing)
- `baselineOnMigrate=true` ensures Flyway handles existing databases correctly
- Migrations are forward-only; rollback is not automated

**Configuration changes:**
- Review release notes for any new configuration properties
- New properties use sensible defaults; existing configuration files do not need changes unless you want to override defaults
- Configuration can also be updated at runtime through the UI or API (see Configuration Settings)

**Backup recommendations:**
- Back up the MySQL data volume before upgrading
- Back up the `configuration/` directory

Frontmatter:
```yaml
---
id: upgrading
title: Upgrading
sidebar_position: 4
---
```

---

### Task 6: Write Deploying Sensors page

**Files:**
- Create: `docs/docs/deploying-sensors.md`

**Step 1: Create the deploying sensors page**

Content must cover:

**How sensors work:**
- Each sensor is a lightweight HTTP service running on a device (typically a Raspberry Pi)
- Sensor Hub polls each registered sensor on a configurable interval (default: 300 seconds)
- Sensors expose a single `GET /temperature` endpoint that returns the current reading
- Communication is plain HTTP; sensors are expected to be on a trusted local network

**Hardware setup:**
- DS18B20 temperature sensor connected to Raspberry Pi via 1-wire protocol
- Enable 1-wire interface on the Raspberry Pi (`raspi-config` or `/boot/config.txt`)

**Software installation:**
1. Install prerequisites: `sudo apt install python3-pip python3.11-venv`
2. Clone the repository: `git clone <repo-url>` and `cd temperature_sensor`
3. Create a Python virtual environment: `python3 -m venv ./venv`
4. Install dependencies: `venv/bin/pip3 install -r requirements.txt`
5. Create a `.env` file: `FLASK_APP=/path/to/sensor_api.py`

**Running as a service:**
1. Edit `temp_sensor_api.service` — update paths and user/group to match your system
2. Copy to systemd: `sudo cp temp_sensor_api.service /etc/systemd/system/`
3. Reload and start: `sudo systemctl daemon-reload && sudo systemctl enable --now temp_sensor_api`
4. The sensor API runs on port 5000 by default

**Sensor API response format:**
```json
{
  "time": "2026-01-15 14:30:00",
  "temperature": 21.56
}
```

**Updating a sensor:**
1. `git pull` on the Raspberry Pi
2. `venv/bin/pip3 install -r requirements.txt`
3. `sudo systemctl restart temp_sensor_api`

**Verifying a sensor:**
```bash
curl http://<sensor-ip>:5000/temperature
```

Frontmatter:
```yaml
---
id: deploying-sensors
title: Deploying Sensors
sidebar_position: 5
---
```

---

### Task 7: Write Managing Sensors page

**Files:**
- Create: `docs/docs/managing-sensors.md`

**Step 1: Create the managing sensors page**

Content must cover:

**Registering sensors:**
- Sensors can be auto-discovered from the OpenAPI spec on startup (if `sensor.discovery.skip` is not set to `true`)
- Sensors can be added manually through the UI or API
- Each sensor record includes: name, type (e.g., `temperature`), URL (the HTTP endpoint to poll)
- The URL must be reachable from the Sensor Hub host

**Sensor types:**
- `temperature` — DS18B20 or compatible temperature sensors
- Additional types can be added by extending the sensor type definitions

**Sensor health monitoring:**
- Sensor Hub tracks the health status of each sensor
- Health is determined by whether the sensor responds successfully when polled
- Health history is retained for a configurable number of days (default: 180)
- The Sensors Overview page and individual Sensor pages show health status and history charts

**Sensor data collection:**
- Readings are collected at the configured interval (default: every 300 seconds / 5 minutes)
- Hourly averages are computed and stored separately for efficient historical queries
- Sensor data is retained for a configurable number of days (default: 365)
- Data cleanup runs on a configurable schedule (default: every 24 hours)

**Managing sensors in the UI:**
- Sensors Overview page lists all registered sensors with their current status
- Individual Sensor pages show detailed health history and temperature data
- Sensors can be added, edited, and deleted through the UI (requires `manage_sensors` permission)

**Relevant permissions:**
- `view_sensors` — view sensor list and data
- `manage_sensors` — add, edit sensors
- `delete_sensors` — delete sensors
- `view_readings` — view temperature readings
- `trigger_readings` — manually trigger a sensor reading

Frontmatter:
```yaml
---
id: managing-sensors
title: Managing Sensors
sidebar_position: 6
---
```

---

### Task 8: Write Alerts and Notifications page

**Files:**
- Create: `docs/docs/alerts-and-notifications.md`

**Step 1: Create the alerts and notifications page**

Content must cover:

**Alert rules:**
- Alert rules are configured per sensor
- Two alert types:
  - Numeric range: triggers when a reading exceeds a high threshold or falls below a low threshold
  - Status-based: triggers when a sensor reports a specific status value
- Each rule has a rate limit (in hours) to prevent repeated alerts for the same condition
- Rules can be enabled or disabled independently

**How alerts are evaluated:**
- When a new reading is collected from a sensor, it is evaluated against the sensor's alert rule
- If the reading meets the alert condition and the rate limit window has elapsed, an alert is triggered
- Triggered alerts are recorded in the alert history and generate a notification

**Notifications:**
- Notifications are delivered through two channels: in-app and email
- In-app notifications appear in the notification bell in the UI header and are delivered in real time via WebSocket
- Email notifications are sent via Gmail SMTP using OAuth 2.0 (requires OAuth configuration)

**Notification categories:**
- Threshold alerts — sensor readings that exceed configured thresholds
- User management — user creation, deletion, and role changes
- Configuration changes — sensors added, updated, or removed; properties changed

**Notification preferences:**
- Each user can configure which notification categories they receive via email and in-app
- Preferences are managed on the Alerts and Notifications page in the UI
- Default: threshold alerts sent via both channels; user management and configuration changes sent in-app only

**Notification lifecycle:**
- Notifications can be marked as read or dismissed
- Dismissed notifications are hidden by default but can be included in queries
- Notifications older than 90 days are automatically purged

**Email notification setup:**
- Requires OAuth configuration (see Installation — OAuth setup)
- SMTP user configured in `configuration/smtp.properties` (`smtp.user=your-email@gmail.com`)
- OAuth token is refreshed automatically in the background (default: every 30 minutes)

**Relevant permissions:**
- `view_alerts` — view alert rules and history
- `manage_alerts` — create, edit, delete alert rules
- `view_notifications` — view notifications and unread count
- `manage_notifications` — dismiss notifications, configure preferences

Frontmatter:
```yaml
---
id: alerts-and-notifications
title: Alerts and Notifications
sidebar_position: 7
---
```

---

### Task 9: Write Session Management page

**Files:**
- Create: `docs/docs/session-management.md`

**Step 1: Create the session management page**

Content must cover:

**How sessions work:**
- Sensor Hub uses session-based authentication
- On successful login, a session token is issued as an HTTP-only cookie
- Each session records the client IP address, user agent, creation time, and expiry time
- Sessions expire after a configurable period (default: 30 days)
- A CSRF token is issued alongside the session to protect against cross-site request forgery on state-changing requests

**Viewing active sessions:**
- The Sessions page in the UI lists all active sessions for the current user
- Each session entry shows the IP address, browser user agent, creation time, last accessed time, and expiry time
- The current session is indicated in the list

**Revoking sessions:**
- Users can revoke any of their own sessions from the Sessions page
- Revoking a session immediately invalidates it; the affected client must log in again
- Session revocation is recorded in an audit trail

**Login rate limiting:**
- Failed login attempts are tracked per username and per IP address
- After exceeding a configurable threshold (default: 5 attempts within 15 minutes), an exponential backoff is applied
- The API returns a `429 Too Many Requests` response with a `Retry-After` header indicating when the next attempt is allowed
- Rate limiting parameters are configurable (see Configuration Settings)

**Session configuration:**
- `auth.session.ttl.minutes` — session duration (default: 43200 minutes / 30 days)
- `auth.session.cookie.name` — cookie name (default: `sensor_hub_session`)
- `auth.login.backoff.window.minutes` — rate limit window (default: 15)
- `auth.login.backoff.threshold` — failed attempts before backoff (default: 5)
- `auth.login.backoff.base.seconds` — base backoff time (default: 2)
- `auth.login.backoff.max.seconds` — maximum backoff time (default: 300)

Frontmatter:
```yaml
---
id: session-management
title: Session Management
sidebar_position: 8
---
```

---

### Task 10: Write User Management and RBAC page

**Files:**
- Create: `docs/docs/user-management.md`

**Step 1: Create the user management and RBAC page**

Content must cover:

**User accounts:**
- Users are managed through the User Management page in the UI or via the API
- Each user has a username, email address, and one or more roles
- The initial admin user is created on first startup via the `SENSOR_HUB_INITIAL_ADMIN` environment variable
- Users can be disabled without deletion

**Creating and managing users:**
- Requires `manage_users` permission
- When creating a user, specify username, email, password, and roles
- The "must change password" flag forces the user to change their password on next login
- While the flag is active, the user can only access the password change endpoint and basic auth endpoints

**Password management:**
- Users can change their own password from the UI
- Passwords are hashed using bcrypt with a configurable cost factor (default: 12)
- Administrators can set the "must change password" flag on any user

**Roles:**
- Three built-in roles: admin, user, viewer
- Roles are containers for permissions; a user's effective permissions are the union of all permissions granted to their assigned roles
- Roles can be assigned to users through the User Management page

**Permissions:**
- Permissions are granular access controls assigned to roles
- Administrators can add or remove permissions from roles through the User Management page

Document the full permission set in a table:

| Permission | Description |
|---|---|
| `manage_sensors` | Add and edit sensors |
| `delete_sensors` | Delete sensors |
| `view_sensors` | View sensor list and data |
| `trigger_readings` | Manually trigger sensor readings |
| `view_alerts` | View alert rules and alert history |
| `manage_alerts` | Create, edit, and delete alert rules |
| `view_roles` | View roles and their permissions |
| `manage_roles` | Assign and remove permissions from roles |
| `manage_users` | Create, delete, and modify user accounts |
| `view_users` | View the user list |
| `view_readings` | View temperature readings |
| `view_notifications` | View notifications |
| `manage_notifications` | Dismiss notifications and configure preferences |
| `manage_properties` | Update system configuration properties |
| `view_properties` | View system configuration properties |
| `manage_oauth` | Configure OAuth settings |

Frontmatter:
```yaml
---
id: user-management
title: User Management and RBAC
sidebar_position: 9
---
```

---

### Task 11: Write API Reference — Authentication

**Files:**
- Create: `docs/docs/api/authentication.md`

**Step 1: Create the authentication API reference page**

Structure: brief intro explaining session-based auth with CSRF, then endpoint-by-endpoint reference.

For each endpoint, document: method, path, authentication requirement, required permission (if any), request parameters/body (with types), response body (with example JSON), error responses.

**Endpoints to document:**

1. `POST /auth/login` — no auth, no CSRF. Request: `{"username": "string", "password": "string"}`. Response 200: `{"must_change_password": false, "csrf_token": "abc123"}` plus `Set-Cookie` header. Error 401: invalid credentials. Error 429: rate limited with `retry_after` field and `Retry-After` header.

2. `POST /auth/logout` — auth required, CSRF bypassed. Response 200: empty body, clears session cookie.

3. `GET /auth/me` — auth required. Response 200: `{"user": {user object}, "csrf_token": "abc123"}`.

**CSRF protection section:**
- State-changing requests (POST, PUT, PATCH, DELETE) require `X-CSRF-Token` header
- CSRF token is returned by `/auth/login` and `/auth/me`
- Exceptions: `/auth/login` and `/auth/logout` bypass CSRF

**Session cookie details:**
- Name: configurable (default `sensor_hub_session`)
- Flags: HttpOnly, Secure (production), SameSite=Lax

Frontmatter:
```yaml
---
id: api-authentication
title: Authentication
sidebar_position: 1
---
```

---

### Task 12: Write API Reference — Sensors and Readings

**Files:**
- Create: `docs/docs/api/sensors-and-readings.md`

**Step 1: Create the sensors and readings API reference page**

**Endpoints to document:**

1. `GET /sensors` — permission: `view_sensors`. Response: array of sensor objects.
2. `GET /sensors/:type` — permission: `view_sensors`. Path param: `type` (e.g., `temperature`). Response: array of sensors filtered by type.
3. `POST /sensors` — permission: `manage_sensors`. Request body: sensor creation fields. Response 201.
4. `PUT /sensors/:id` — permission: `manage_sensors`. Request body: sensor update fields.
5. `DELETE /sensors/:id` — permission: `delete_sensors`.
6. `GET /temperature/current` — permission: `view_readings`. Response: array of current readings from all temperature sensors.
7. `GET /temperature/hourly-average` — permission: `view_readings`. Query params: `sensorId`, `from`, `to`. Response: array of hourly average data points.
8. `POST /temperature/collect` — permission: `trigger_readings`. Manually triggers a collection cycle.
9. `GET /sensors/:id/health-history` — permission: `view_sensors`. Query params: `limit`. Response: array of health status records.

**WebSocket endpoints:**
- `GET /temperature/ws/current-temperatures` — real-time temperature updates, topic `current_temperatures`
- `GET /sensors/ws/:type` — sensor metadata changes, topic `sensors:{type}`

Document WebSocket message format for each.

Frontmatter:
```yaml
---
id: api-sensors-and-readings
title: Sensors and Readings
sidebar_position: 2
---
```

---

### Task 13: Write API Reference — Alerts and Notifications

**Files:**
- Create: `docs/docs/api/alerts-and-notifications.md`

**Step 1: Create the alerts and notifications API reference page**

**Alert endpoints:**

1. `GET /alerts` — permission: `view_alerts`. Response: array of alert rule objects.
2. `GET /alerts/:sensorId` — permission: `view_alerts`. Response: single alert rule.
3. `GET /alerts/:sensorId/history` — permission: `view_alerts`. Query param: `limit` (default 50, max 100). Response: array of alert history records.
4. `POST /alerts` — permission: `manage_alerts`. Request body: alert rule object with fields: `sensor_id`, `alert_type` (`numeric_range` or `status_based`), `high_threshold`, `low_threshold`, `trigger_status`, `enabled`, `rate_limit_hours`.
5. `PUT /alerts/:sensorId` — permission: `manage_alerts`. Request body: alert rule object.
6. `DELETE /alerts/:sensorId` — permission: `manage_alerts`.

**Notification endpoints:**

1. `GET /notifications` — permission: `view_notifications`. Query params: `limit` (default 50), `offset` (default 0), `include_dismissed` (default false). Response: array of user notification objects with nested notification details.
2. `GET /notifications/unread-count` — permission: `view_notifications`. Response: `{"count": 3}`.
3. `POST /notifications/:id/read` — permission: `view_notifications`.
4. `POST /notifications/:id/dismiss` — permission: `manage_notifications`.
5. `POST /notifications/bulk/read` — permission: `view_notifications`. Marks all as read.
6. `POST /notifications/bulk/dismiss` — permission: `manage_notifications`. Dismisses all.
7. `GET /notifications/preferences` — permission: `view_notifications`. Response: array of channel preference objects.
8. `POST /notifications/preferences` — permission: `manage_notifications`. Request body: `{"category": "threshold_alert", "email_enabled": true, "inapp_enabled": true}`.

**WebSocket endpoint:**
- `GET /notifications/ws` — permission: `view_notifications`. Per-user notification feed.

Frontmatter:
```yaml
---
id: api-alerts-and-notifications
title: Alerts and Notifications
sidebar_position: 3
---
```

---

### Task 14: Write API Reference — Users, Roles, and Sessions

**Files:**
- Create: `docs/docs/api/users-roles-sessions.md`

**Step 1: Create the users, roles, and sessions API reference page**

**User endpoints:**

1. `POST /users` — permission: `manage_users`. Request body: `{"username": "string", "email": "string", "password": "string", "roles": ["string"]}`. Response 201: `{"id": 1}`.
2. `GET /users` — permission: `view_users`. Response: array of user objects.
3. `PUT /users/password` — auth required (any user). Request body: `{"user_id": 1, "new_password": "string"}`. `user_id` optional, defaults to current user.
4. `DELETE /users/:id` — permission: `manage_users`. Cannot delete current user.
5. `PATCH /users/:id/must_change` — permission: `manage_users`. Request body: `{"must_change": true}`.
6. `POST /users/:id/roles` — permission: `manage_users`. Request body: `{"roles": ["admin", "user"]}`.

**Role endpoints:**

1. `GET /roles` — permission: `view_roles`. Response: array of role objects.
2. `GET /roles/permissions` — permission: `view_roles`. Response: array of all available permissions.
3. `GET /roles/:id/permissions` — permission: `view_roles`. Response: permissions assigned to the role.
4. `POST /roles/:id/permissions/:permissionId` — permission: `manage_roles`. Grants permission to role.
5. `DELETE /roles/:id/permissions/:permissionId` — permission: `manage_roles`. Revokes permission from role.

**Session endpoints:**

1. `GET /auth/sessions` — auth required. Response: array of session objects for the current user. Each includes `id`, `created_at`, `expires_at`, `last_accessed_at`, `ip_address`, `user_agent`, `current` (boolean).
2. `DELETE /auth/sessions/:id` — auth required. Revokes the specified session.

Frontmatter:
```yaml
---
id: api-users-roles-sessions
title: Users, Roles, and Sessions
sidebar_position: 4
---
```

---

### Task 15: Write API Reference — Properties and OAuth

**Files:**
- Create: `docs/docs/api/properties-and-oauth.md`

**Step 1: Create the properties and OAuth API reference page**

**Properties endpoints:**

1. `GET /properties` — permission: `view_properties`. Response: key-value object of all properties. Sensitive values (e.g., `database.password`) are masked as `*****`.
2. `PATCH /properties` — permission: `manage_properties`. Request body: key-value object of properties to update. Sending `*****` for a sensitive property leaves it unchanged. Response 202: `{"message": "Property updated successfully"}`. Changes are applied immediately in memory, saved to configuration files asynchronously, and broadcast to all connected WebSocket clients.

**Properties WebSocket:**
- `GET /properties/ws` — permission: `view_properties`. Receives property updates in real time.

**OAuth endpoints:**

1. `GET /oauth/status` — permission: `manage_oauth`. Response: `{"configured": true, "ready": true}`.
2. `GET /oauth/authorize` — permission: `manage_oauth`. Response: `{"auth_url": "https://accounts.google.com/...", "state": "csrf-state"}`. Returns the Google OAuth authorization URL.
3. `POST /oauth/submit-code` — permission: `manage_oauth`. Request body: `{"code": "auth-code", "state": "csrf-state"}`. Exchanges the authorization code for tokens.
4. `POST /oauth/reload` — permission: `manage_oauth`. Reloads OAuth credentials from disk.

**Health endpoint:**
- `GET /health` — no authentication required. Response: `{"status": "ok"}`.

Frontmatter:
```yaml
---
id: api-properties-and-oauth
title: Properties and OAuth
sidebar_position: 5
---
```

---

### Task 16: Write Configuration Settings page

**Files:**
- Create: `docs/docs/configuration.md`

**Step 1: Create the configuration settings page**

Content must cover:

**Configuration file locations:**
- `configuration/application.properties` — application and sensor settings
- `configuration/database.properties` — database connection details
- `configuration/smtp.properties` — email sender configuration

**How configuration works:**
- Properties are loaded from files on startup
- Properties can be updated at runtime through the UI (Properties page) or via the API (`PATCH /properties`)
- Runtime updates are applied immediately in memory and saved back to the configuration files asynchronously
- All connected clients are notified of changes via WebSocket

**Application properties reference table:**

| Property | Default | Description |
|---|---|---|
| `sensor.collection.interval` | `300` | Seconds between sensor polling cycles |
| `sensor.discovery.skip` | `true` | Skip automatic sensor discovery on startup |
| `health.history.retention.days` | `180` | Days to retain sensor health history records |
| `sensor.data.retention.days` | `365` | Days to retain temperature reading data |
| `data.cleanup.interval.hours` | `24` | Hours between data cleanup runs |
| `auth.bcrypt.cost` | `12` | Bcrypt cost factor for password hashing |
| `auth.session.ttl.minutes` | `43200` | Session expiry time in minutes (default: 30 days) |
| `auth.session.cookie.name` | `sensor_hub_session` | Name of the session cookie |
| `auth.login.backoff.window.minutes` | `15` | Time window for counting failed login attempts |
| `auth.login.backoff.threshold` | `5` | Failed attempts before backoff is applied |
| `auth.login.backoff.base.seconds` | `2` | Base duration for exponential backoff |
| `auth.login.backoff.max.seconds` | `300` | Maximum backoff duration |
| `oauth.credentials.file.path` | `configuration/credentials.json` | Path to Google OAuth credentials file |
| `oauth.token.file.path` | `configuration/token.json` | Path to stored OAuth token file |
| `oauth.token.refresh.interval.minutes` | `30` | Interval for background OAuth token refresh |

**Database properties reference table:**

| Property | Description |
|---|---|
| `database.username` | MySQL username |
| `database.password` | MySQL password (masked in API responses) |
| `database.hostname` | MySQL host |
| `database.port` | MySQL port |

**SMTP properties reference table:**

| Property | Description |
|---|---|
| `smtp.user` | Gmail address used as the sender for email notifications |

**Environment variables:**

| Variable | Description |
|---|---|
| `DB_HOST` | Overrides `database.hostname` (used in Docker) |
| `DB_PORT` | Overrides `database.port` (used in Docker) |
| `DB_USER` | Overrides `database.username` (used in Docker) |
| `DB_PASS` | Overrides `database.password` (used in Docker) |
| `TLS_CERT_FILE` | Path to TLS certificate (enables HTTPS on backend) |
| `TLS_KEY_FILE` | Path to TLS private key (enables HTTPS on backend) |
| `SENSOR_HUB_ALLOWED_ORIGIN` | CORS allowed origin for the UI |
| `SENSOR_HUB_INITIAL_ADMIN` | `username:password` to create initial admin user |

Frontmatter:
```yaml
---
id: configuration
title: Configuration Settings
sidebar_position: 10
---
```

---

### Task 17: Configure sidebar navigation

**Files:**
- Modify: `docs/sidebars.ts`

**Step 1: Configure the sidebar**

```typescript
const sidebars = {
  docs: [
    'overview',
    'prerequisites',
    'installation',
    'upgrading',
    'deploying-sensors',
    'managing-sensors',
    'alerts-and-notifications',
    'session-management',
    'user-management',
    {
      type: 'category',
      label: 'API Reference',
      items: [
        'api/api-authentication',
        'api/api-sensors-and-readings',
        'api/api-alerts-and-notifications',
        'api/api-users-roles-sessions',
        'api/api-properties-and-oauth',
      ],
    },
    'configuration',
  ],
};

export default sidebars;
```

---

### Task 18: Build and verify

**Step 1: Build the Docusaurus site**

```bash
cd docs && npm run build
```

Expected: Build succeeds with no errors.

**Step 2: Serve locally and verify**

```bash
cd docs && npm run serve
```

Verify:
- Landing page loads and shows the Overview content
- Sidebar navigation works and shows all pages
- All page links resolve correctly
- API Reference category expands to show sub-pages
- No broken links

---

### Task 19: Clean up old documentation files

**Step 1: Verify old docs are no longer needed**

Confirm the following files were removed in Task 1:
- `docs/alerting-system.md`
- `docs/introduction.md`
- `docs/notification-system.md`
- `docs/mobile-responsiveness.md`
- `docs/testing-guide.md`

**Step 2: Update root README.md**

Add a documentation section to the root README pointing to the Docusaurus site:

```markdown
## Documentation

User guide documentation is available in the `docs/` directory. To build and serve locally:

\`\`\`bash
cd docs
npm install
npm run start
\`\`\`
```

