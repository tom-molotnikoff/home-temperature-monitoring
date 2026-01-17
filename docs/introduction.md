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
    - `api/` — REST and WebSocket handlers (see `api.go`, `sensor_api.go`, `temperature_api.go`, `websocket.go`).
    - `ws/` — websocket hub for registering and handling many websockets (`hub.go`).
    - `service/` — business logic layers (sensorService, temperatureService, propertiesService).
    - `db/` — DB helpers and repository interfaces (`sensor_repository.go`, `temperature_repository.go`).
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
5) Obtaining credentials.json for Gmail SMTP (optional, for email alerts)
```bash
# Download credentials.json from Google Cloud Console (OAuth 2.0 Client)
# Needs to be credentials.json for "Desktop app" type with Gmail API enabled.
# Need to make sure your email is a test user in OAuth consent screen settings.
# Place it in sensor_hub/configuration/credentials.json

# Configure the paths in application.properties:
# oauth.credentials.file.path=configuration/credentials.json
# oauth.token.file.path=configuration/token.json

# Token can be obtained via:
# - CLI: Run sensor_hub/pre_authorise_application (go run ./main.go)
# - UI: Navigate to /admin/oauth and click "Start OAuth Flow" (requires manage_oauth permission)
```

OAuth Management via UI
-----------------------
The OAuth page at `/admin/oauth` (requires `manage_oauth` permission) provides:
- **Status display**: Shows if OAuth is configured, token validity, auto-refresh status
- **Reload Config**: Reloads credentials.json from disk without restart
- **Start OAuth Flow**: Opens Google consent screen in a new tab for token authorization
- **Copy Authorization URL**: Manually copy the URL if popup is blocked

After completing the Google consent flow, the callback automatically stores the token. The OAuth service auto-refreshes tokens based on the configured interval.

Configuration in `application.properties`:
```
oauth.credentials.file.path=configuration/credentials.json
oauth.token.file.path=configuration/token.json  
oauth.token.refresh.interval.minutes=30
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

Frontend architecture & UI implementation details
------------------------------------------------
The frontend is a single-page React application (Vite + TypeScript) located at `sensor_hub/ui/sensor_hub_ui`. The UI is implemented using Material UI (MUI) components and follows the project's existing conventions for "cards" and layout helpers. This section documents the architecture and the key files to make it easier to understand and extend the UI.

Key libraries and patterns
- React + TypeScript (Vite) for the SPA.
- Material UI (MUI) for components and styling (see `package.json` dependencies).
- `@mui/x-data-grid` used for tabular views (sensor lists, users list, etc.).
- Small local UI helpers/components (under `src/tools/`) provide consistent layout and typography across pages (for example `LayoutCard.tsx`, `PageContainer.tsx`, `Typography.tsx`).
- A lightweight API client (`src/api/Client.ts`) centralises fetch logic; it is configured to send cookies (credentials: 'include') and to include a CSRF token header when available.

Routing and pages
- Routing is handled by `src/navigation/AppRoutes.tsx` (uses `react-router`). Key routes:
  - `/` — Temperature dashboard (`src/pages/temperature-dashboard/TemperatureDashboard.tsx`)
  - `/sensors-overview` — sensors listing
  - `/sensor/:id` — sensor details page
  - `/properties-overview` — application properties UI
  - `/alerts` — alert rules management (`src/pages/alerts/AlertsPage.tsx`)
  - `/login` — login page (`src/pages/Login.tsx`)
  - `/account/change-password` — forced password change page (`src/pages/Account/ChangePassword.tsx`)
  - `/account/sessions` — session management and revoke page (`src/pages/Account/SessionsPage.tsx`)
  - `/admin/users` — admin users management (`src/pages/admin/UsersPage.tsx`)
  - `/admin/roles` — admin roles management (`src/pages/admin/RolesPage.tsx`)
  - `/admin/oauth` — OAuth credentials management (`src/pages/admin/OAuthPage.tsx`)

Auth flow and session handling
- The backend issues session tokens stored in an HttpOnly cookie (name configurable in application properties). The SPA uses fetch() with credentials included so the cookie is sent automatically.
- CSRF protection uses a per-session token: when a session is created (login) the server generates a random CSRF token and returns it in the JSON response. The SPA stores the CSRF token in memory (not localStorage) and sends it as `X-CSRF-Token` on state-changing requests. The server validates this token against the server-stored token for the session.
- On application load the SPA calls `GET /auth/me` to obtain the current user and CSRF token (if logged-in), so pages and the client are populated immediately.
- Session audit: when sessions are revoked (via API or UI), the server records an audit row in `session_audit` (migration V10) describing who revoked which session and when.

How UI components are organized
- `src/pages/*` contains page-level components which compose smaller components.
- `src/components/*` contains reusable components like `SensorsDataGrid.tsx`, `CurrentTemperatures.tsx`, and `SensorHealthHistory.tsx`. These use MUI primitives (DataGrid, Box, Snackbar, Dialogs) to match the app look and feel.
- `src/providers/*` contains React context providers (sensor and date contexts) used across pages.
- `src/api/*` contains typed API clients for the backend endpoints (Client.ts, Auth.ts, Users.ts, Sensors.ts, etc.).
- `src/hooks/useMobile.ts` provides the `useIsMobile()` hook for responsive design (breakpoint: 950px).
- `src/tools/DesktopRowMobileColumn.tsx` is a layout helper that switches between row (desktop) and column (mobile) flex layouts.

Mobile responsiveness
- The UI is fully responsive for screens below 950px (mobile breakpoint).
- Key mobile behaviors:
  - **Grids**: Pages use `Grid size={isMobile ? 12 : 6}` to stack items vertically on mobile.
  - **Charts**: Recharts graphs use a `compact` prop that reduces height, rotates X-axis labels 45°, and increases tick gap.
  - **DataGrids**: Non-essential columns are hidden on mobile (e.g., timestamps, IDs). Sessions page shows a short device name parsed from User-Agent.
  - **Alerts page**: Uses a card-based list on mobile instead of DataGrid for better touch interaction.
  - **Date pickers**: Stack vertically and use full width on mobile. Default date range is 2 days on mobile (vs 7 days desktop).
  - **Buttons**: Button groups stack vertically with full width on mobile (e.g., OAuth page).
- See `docs/mobile-responsiveness.md` for detailed implementation guide.

Files to look at for the UI implementation
- `sensor_hub/ui/sensor_hub_ui/src/api/Client.ts` — centralised client; sets credentials and CSRF header.
- `sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx` — SPA routing.
- `sensor_hub/ui/sensor_hub_ui/src/pages/Login.tsx` — MUI-based login form.
- `sensor_hub/ui/sensor_hub_ui/src/pages/Account/ChangePassword.tsx` — MUI-based change-password page.
- `sensor_hub/ui/sensor_hub_ui/src/pages/admin/UsersPage.tsx` — admin users UI using MUI DataGrid and Dialog.
- `sensor_hub/ui/sensor_hub_ui/src/components/SensorsDataGrid.tsx` — example of how DataGrid, Menu and Snackbar are used in other components.
- `sensor_hub/ui/sensor_hub_ui/src/tools/LayoutCard.tsx` and `Typography.tsx` — site-wide layout helpers.

Developer notes
- When adding new state-changing endpoints that the SPA will call, ensure the client includes the `X-Requested-With` header (the default client already does this). If you need strict CSRF tokens instead, a token-based CSRF approach can be implemented (server issues token via JSON then SPA includes it in a header).
- To add or modify pages that require admin-only access, update the backend RBAC (roles) and protect routes with the `RequireAdmin()` middleware.
- The UI assumes the API is hosted under the same domain or a properly configured CORS policy and that cookies are accepted; adjust `API_BASE` in `src/environment/Environment` accordingly.

Operational notes
- To create an initial admin account during deployment set the env var `SENSOR_HUB_INITIAL_ADMIN=username:password`; Flyway (docker-compose) will run DB migrations, then the service will create the initial admin if no users exist.
- The UI can be started locally (HMR) from `sensor_hub/ui/sensor_hub_ui` with `npm install` and `npm run dev` (see Quick start in this document).


## Developer summary & architecture (detailed)

This section records a detailed, up-to-date mental model of the project and implementation decisions so a new developer can orient themselves from a single file.

High-level service architecture
- The backend is a single Go process exposing both HTTP REST handlers (Gin framework) and WebSocket upgrade handlers. Core concerns are split into small packages:
  - `api/` implements request handlers, request validation and response shaping (controller layer).
  - `service/` contains business logic used by handlers and background jobs (for example `properties_service` and `temperatureService`).
  - `db/` holds repository implementations and interfaces; SQL schema lives in `db/changesets` and Flyway manages migrations in docker runs.
  - `application_properties/` encapsulates loading defaults, reading/writing configuration maps and persisting them to `sensor_hub/configuration/` files.

Authentication, sessions, CSRF and RBAC
---------------------------------------
- Authentication uses cookie-based sessions. Session tokens are stored in an HttpOnly cookie (name configurable via application properties). The SPA uses fetch() with credentials included so the cookie is sent automatically.
- The server stores session state (user id, roles, CSRF token, expiry) server-side and sends minimal session info to the client.
- CSRF protection uses a token-based flow: after login the server returns a freshly generated CSRF token in the JSON response. The SPA stores the CSRF token in memory and includes it as `X-CSRF-Token` for state-changing requests. The server validates the token against the server-side session.
- On application load the SPA calls `GET /auth/me` to obtain the current user and CSRF token (if logged-in), so pages and the client are populated immediately.
- Role-based access control (RBAC) is implemented server-side. Middleware enforces permission checks (`RequirePermission`) and a convenience `RequireAdmin()` middleware is available for admin-only routes.

Recent auth & security changes (brief)
-------------------------------------
- New user & RBAC features: the service now includes user management, roles, permissions, session listing/revocation, and session audit (see `sensor_hub/service/*`, `sensor_hub/db/*`, `sensor_hub/api/usersApi.go`).
- CSRF middleware is applied globally for state-changing requests; the SPA must send `X-CSRF-Token`. See `sensor_hub/api/middleware/csrf_middleware.go`.
- Session cookie handling tightened: cookies are set with HttpOnly and SameSite; Secure is enabled automatically when TLS or production mode is detected. Logout clears the cookie with matching attributes. See `sensor_hub/api/authApi.go` and `sensor_hub/service/sessionRepository.go`.
- Login backoff (anti-brute-force): the service no longer sleeps in request handlers. When the failed-attempt threshold is reached the API responds with HTTP 429 and a `Retry-After` header and JSON diagnostics. This prevents handler goroutine blocking and allows upstream rate limiting. See `sensor_hub/service/authService.go` and `sensor_hub/api/authApi.go`.
- In-process login limiter: to make `Retry-After` deterministic for concurrent requests, an in-memory blocker records block windows and provides a small "allow-once" post-block allowance so a correct credential attempt immediately after expiry can succeed. This blocker is process-local; for multi-instance deployments prefer a Redis/DB-backed blocker. See `sensor_hub/service/loginLimiter.go`.
- Allow-once & re-blocking behaviour: after a block expires a small number of immediate attempts are permitted (configurable); failed attempts during that window may re-trigger blocking. Successful logins clear failed-attempt records and any pending allow-once entries.
- Middleware and performance: permission checks cache permissions in the request context to avoid repeated DB queries during a single request. Must-change-password enforcement is implemented at middleware level and matches on method+path to avoid accidental bypass.

Files of interest (auth & security)
- `sensor_hub/service/auth_service.go` — login, session creation, backoff logic
- `sensor_hub/service/login_limiter.go` — in-process login blocker and allow-once behavior
- `sensor_hub/api/middleware/csrf_middleware.go` — CSRF enforcement
- `sensor_hub/api/middleware/auth_middleware.go` — session validation and must-change-password enforcement
- `sensor_hub/api/middleware/permission_middleware.go` — permission enforcement and per-request caching
- `sensor_hub/api/auth_api.go` — login/logout/me endpoints and cookie handling

Frontend conventions and key files
- The SPA stores auth state in a React context provider `src/providers/AuthContext.tsx` that exposes `{ user, refresh }` and helper `useAuth()` for pages to check roles and access.
- UI pages follow a pattern where `PageContainer` wraps top-level page structure and `LayoutCard` provides consistent paper/card appearance.
- Data editing forms and tables are MUI-driven. For admin pages (users, roles, sessions) the DataGrid is used and context menus/dialogs implement CRUD behaviour.
- The client API (`src/api/Client.ts`) sets default headers and attaches the `X-CSRF-Token` header when available; state-changing endpoints expect that header and cookie-based session.

Testing & debugging tips
- For UI debugging, run the SPA with Vite (`npm run dev`) and watch network requests in browser devtools. Look specifically for `GET /auth/me`, missing CSRF tokens, or CORS issues for cross-origin setups.
- Server logs show API route registrations via Gin at startup; use these to confirm which handlers are active.
- When changing config via the UI, confirm `sensor_hub/configuration/*.properties` are updated on disk; the service will reload in-memory config when the maps are reloaded.

Operational & deployment notes (concise)
- Recommended: deploy via the Docker compose provided under `sensor_hub/docker/` or `sensor_hub/docker_tests/` for local testing. Flyway container will apply migrations when configured in compose.
- Provision initial admin via env var (e.g., `SENSOR_HUB_INITIAL_ADMIN=username:password`) during first run; the service will create the admin account if no users exist.
- Keep `configuration/` files secure (contain DB credentials and SMTP secrets).
