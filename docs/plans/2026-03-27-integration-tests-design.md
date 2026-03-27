# Integration Test Framework Design

## Problem

The project has ~150 unit tests using mocks (sqlmock, testify/mock, httptest) but zero integration tests against a real database and real HTTP sensor endpoints. This allowed several production bugs to ship undetected:

- Case-sensitive SQL queries silently returning empty results
- `sql.NullTime` scan errors on SQLite TEXT columns
- Goroutines dying silently with no panic recovery
- Config writes failing under systemd's `ProtectSystem=strict`

Unit tests with mocks can't catch these classes of bugs because the mocks bypass the real database driver, real SQL execution, and real HTTP collection pipeline.

## Solution

A comprehensive integration test framework that exercises the full stack: HTTP API → Gin router → middleware → service layer → repository → real SQLite database, with real HTTP calls to mock sensor containers.

## Architecture

### Three Layers

```
┌─────────────────────────────────────────────────┐
│  Integration Test Files (//go:build integration) │
│  sensor_hub/integration/*_test.go                │
├─────────────────────────────────────────────────┤
│  Test Harness (sensor_hub/testharness/)           │
│  - Server startup, auth, HTTP client helpers      │
├─────────────────────────────────────────────────┤
│  Testcontainers (mock sensor Docker containers)   │
│  - Built from docker_tests/mock-sensor.dockerfile │
└─────────────────────────────────────────────────┘
```

### Components

**1. Testcontainers (mock sensors)**

Uses `testcontainers-go` to launch mock sensor Docker containers from the existing `docker_tests/mock-sensor.dockerfile`. Each container runs the Flask mock sensor API returning random temperatures (18-22°C) with UTC timestamps. Containers are shared per test suite (stateless).

**2. Test Harness (`sensor_hub/testharness/`)**

A Go package providing reusable setup and client helpers:

- **`harness.go`** — Core setup/teardown:
  ```go
  type Env struct {
      ServerURL    string     // http://localhost:<random-port>
      AdminToken   string     // session cookie for admin user
      SensorURLs   []string   // mock sensor container URLs
      DBPath       string     // temp SQLite file path
      Cleanup      func()     // teardown function
  }
  ```
  `Setup(t)` creates a temp directory, starts mock sensor containers, writes minimal config files, initialises the SQLite database (with auto-migrations), creates an admin user, starts the Gin server on a random port, logs in as admin, and returns `Env`.

- **`client.go`** — HTTP test client wrapping `http.Client` with methods matching the API surface:
  ```go
  client.AddSensor(sensor) → (response, statusCode)
  client.CollectAll() → (response, statusCode)
  client.GetReadingsBetween(from, to, sensor) → ([]reading, statusCode)
  client.Login(username, password) → (token, statusCode)
  client.CreateAlertRule(rule) → (response, statusCode)
  // ... etc for all API endpoints
  ```

- **`containers.go`** — Testcontainers lifecycle management. Builds the mock sensor image, starts N containers, waits for health (HTTP GET `/temperature` returns 200), returns host-mapped URLs.

**3. Integration test files (`sensor_hub/integration/`)**

All integration tests in a single package with `//go:build integration` tag:

```
sensor_hub/integration/
├── main_test.go          # TestMain — starts harness once
├── auth_test.go          # Login, logout, sessions, API keys, rate limiting
├── sensor_crud_test.go   # Add, update, delete, list, enable/disable
├── collection_test.go    # Collect-all, collect-by-name, case-insensitive
├── readings_test.go      # Between dates, hourly averages, sensor filter
├── alerts_test.go        # Rules, triggering, rate limiting, history
├── notifications_test.go # CRUD, bulk ops, preferences
├── users_test.go         # CRUD, roles, permissions
├── properties_test.go    # Get/set configuration
└── health_test.go        # Health endpoint, sensor health history
```

### Test Lifecycle

```
TestMain
  ├── testharness.Setup(m)
  │     ├── Start 2 mock sensor containers (testcontainers)
  │     ├── Create temp directory + config files
  │     ├── InitialiseDatabase (real SQLite + migrations)
  │     ├── Create repositories + services + API handlers
  │     ├── Start Gin server on :0 (random port)
  │     ├── Create admin user + login → get auth token
  │     └── Return Env{ServerURL, AdminToken, SensorURLs, ...}
  ├── m.Run() — all Test* functions execute
  └── env.Cleanup()
        ├── Stop Gin server
        ├── Close database
        ├── Remove temp directory
        └── Terminate mock sensor containers
```

### Isolation Strategy

- **Mock sensor containers**: Shared across all tests (stateless — each GET returns a fresh random reading)
- **SQLite database**: One per test suite (created in TestMain). Tests run sequentially within the suite.
- **Server**: One per test suite, listening on a random port
- **Test data**: Each test file creates its own test data. A `resetDB()` helper is available to truncate tables between test files if needed.

## CI Integration

### New Workflow: `.github/workflows/ci.yml`

Triggers on PRs, pushes to main, and release tags:

```yaml
name: CI
on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: cd sensor_hub && go test ./...

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: cd sensor_hub && go test -tags integration -timeout 10m ./...

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '25'
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: cd sensor_hub/ui/sensor_hub_ui && npm ci && npm run build
      - run: cd sensor_hub && go build -o sensor-hub .
```

### Running Locally

```bash
# Unit tests only (fast, no Docker)
cd sensor_hub && go test ./...

# Integration tests (requires Docker)
cd sensor_hub && go test -tags integration -timeout 10m ./...

# Both
cd sensor_hub && go test -tags integration -timeout 10m ./...
# (integration tag doesn't exclude non-tagged tests)
```

## Test Coverage Plan

### Auth (`auth_test.go`)
- Login with valid credentials → 200 + session cookie
- Login with wrong password → 401
- Access protected endpoint without auth → 401
- Access protected endpoint with expired session → 401
- API key auth: create key → use key → revoke key → verify revoked fails
- Rate limiting: multiple failed logins trigger backoff
- Permission enforcement: viewer can't manage sensors (403)

### Sensor CRUD (`sensor_crud_test.go`)
- Add sensor → 201, verify in list
- Add duplicate name → error
- Update sensor name/URL → 200, verify change
- Delete sensor → cascades (readings, health history cleaned)
- List sensors → returns all
- Get sensor by name → 200
- Get sensor by name (wrong case) → still works (LOWER fix)
- Enable/disable sensor → status changes
- Sensor exists (HEAD) → 200 vs 404

### Collection (`collection_test.go`)
- Register 2 sensors, collect-all → readings stored for both
- Collect-by-name → only that sensor's reading stored
- Collect-by-name with wrong case → still works
- Collect when sensor unreachable → health status updates, no crash
- Collect with disabled sensor → skipped
- Verify reading timestamps are reasonable (not 1 hour off, not zero)
- Verify reading temperatures are in expected range (18-22°C from mock)

### Readings (`readings_test.go`)
- Get readings between dates → filtered correctly
- Get readings with sensor filter → only that sensor
- Get readings with sensor filter (wrong case) → still works
- Hourly averages: insert readings, trigger compute, verify averages exist
- Latest readings → one per sensor, most recent
- No results → empty array (not error)

### Alerts (`alerts_test.go`)
- Create alert rule for sensor → 201
- Get alert rule by sensor ID → matches
- Get alert rule by sensor name → matches
- Trigger collection with reading above high threshold → alert fires
- Rate limiting → alert doesn't re-fire within window
- Alert history → records triggering value
- Update/delete alert rule

### Notifications (`notifications_test.go`)
- Create notification → assigned to users with permission
- List notifications for user → returns expected
- Mark as read → is_read flag changes
- Dismiss → is_dismissed flag changes
- Bulk mark-as-read, bulk dismiss
- Unread count accuracy
- Channel preferences: set email/in-app toggles, verify persistence

### Users & Roles (`users_test.go`)
- Create user → 201
- List users → includes new user
- Change password → can login with new password
- Set must-change flag → verified
- Assign roles → permissions change
- Delete user → sessions cleaned up
- Role-based access: admin vs viewer vs user

### Properties (`properties_test.go`)
- Get all properties → returns defaults
- Set property → persists
- Get updated property → reflects change

### Health (`health_test.go`)
- Health endpoint → 200
- Sensor health history after collection → recorded
- Sensor health after failed collection → unhealthy status
