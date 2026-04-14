# Database

This guide covers the database schema, migration system, and patterns used in
the data access layer.

## Overview

Sensor Hub uses SQLite with the following configuration:

- **Journal mode:** WAL (Write-Ahead Logging) for concurrent reads
- **Foreign keys:** enabled via pragma
- **Max connections:** 1 (SQLite single-writer constraint)
- **Driver:** `modernc.org/sqlite` (pure Go, no CGo)
- **Instrumentation:** OpenTelemetry metrics via `otelsql`

The database file location is configured via `database.path` in
`database.properties` (default: `data/sensor_hub.db`).

## Schema

### Entity Relationship Diagram

```
sensors ─────────────┬── readings
                     ├── sensor_health_history
                     └── sensor_alert_rules ── alert_sent_history

measurement_types ── sensor_measurement_types ── sensors
                 └── measurement_type_aggregations

users ───────────────┬── user_roles ── roles ── role_permissions ── permissions
                     ├── sessions ── session_audit
                     ├── failed_login_attempts
                     ├── user_notifications ── notifications
                     ├── notification_channel_preferences
                     └── api_keys

notification_channel_defaults (standalone lookup table)
failed_login_summary (standalone aggregate table)
```

### Tables

#### Sensor Data

| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `sensors` | Sensor configuration | name (unique), sensor_driver, url, health_status, enabled |
| `readings` | Raw sensor readings | sensor_id (FK), measurement_type_id (FK), numeric_value, text_state, time |
| `measurement_types` | Measurement type definitions | name (unique), unit |
| `sensor_measurement_types` | Sensor-to-measurement-type mapping | sensor_id (FK), measurement_type_id (FK). UNIQUE(sensor_id, measurement_type_id) |
| `measurement_type_aggregations` | Default aggregation function per measurement type | measurement_type_id (FK, unique), aggregation_function (e.g. `avg`, `last`) |
| `sensor_health_history` | Audit trail of health status changes | sensor_id (FK), health_status, recorded_at |

#### Alert System

| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `sensor_alert_rules` | Alert configuration per sensor | sensor_id (FK), alert_type (`numeric_range` or `status_based`), thresholds, rate_limit_hours, enabled |
| `alert_sent_history` | Record of triggered alerts (used for rate limiting) | alert_rule_id (FK), sensor_id (FK), sent_at, reading_value |

#### Users and Auth

| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `users` | User accounts | username (unique), password_hash, must_change_password, disabled |
| `roles` | Role definitions (admin, user, viewer) | name (unique) |
| `user_roles` | User-to-role mapping | (user_id, role_id) PK |
| `permissions` | Permission definitions (17 pre-seeded) | name (unique) |
| `role_permissions` | Role-to-permission mapping | (role_id, permission_id) PK |
| `sessions` | Active user sessions | token_hash (SHA256), csrf_token, expires_at |
| `session_audit` | Session event log (logout, revocation) | session_id (FK), event_type |
| `failed_login_attempts` | Detailed login failure records | username, ip_address, attempt_time, reason |
| `failed_login_summary` | Aggregated failure counts per identifier | identifier (PK), attempts |
| `api_keys` | API key authentication | key_prefix, key_hash (unique), user_id (FK), expires_at, revoked |

#### Notifications

| Table | Purpose | Key Columns |
|-------|---------|-------------|
| `notifications` | Notification messages | category, severity, title, message, metadata (JSON) |
| `user_notifications` | Per-user notification state | user_id (FK), notification_id (FK), is_read, is_dismissed. UNIQUE(user_id, notification_id) |
| `notification_channel_defaults` | Default channel preferences by category | category (unique), email_enabled, inapp_enabled |
| `notification_channel_preferences` | Per-user channel overrides | user_id (FK), category. UNIQUE(user_id, category) |

### Data Types

SQLite has limited types. The codebase uses:

| Go Type | SQLite Type | Notes |
|---------|-------------|-------|
| `int` | INTEGER | Primary keys use AUTOINCREMENT |
| `bool` | INTEGER | 0 = false, 1 = true |
| `float64` | REAL | Numeric sensor values |
| `string` | TEXT | Names, descriptions, hashes |
| `time.Time` | TEXT | ISO 8601 format. Use custom `SQLiteTime` / `NullSQLiteTime` types for scanning |
| JSON | TEXT | Notification metadata stored as JSON string |

**Important:** The `modernc.org/sqlite` driver returns TEXT columns as Go
`string`, not `time.Time`. Standard `sql.NullTime` cannot scan these values.
Use the custom `SQLiteTime` and `NullSQLiteTime` types from `db/sqlite_time.go`
instead.

### Cascade Behaviour

Deleting a user cascades to: user_roles, sessions (→ session_audit),
user_notifications, notification_channel_preferences. Deleting a sensor
cascades to: sensor_alert_rules (→ alert_sent_history). Readings
and health history have foreign keys but no cascade — they are cleaned up by
the periodic data cleanup task.

## Migration System

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate) with
SQL files embedded into the binary.

### File Location

```
sensor_hub/db/migrations/
├── 000001_init_schema.up.sql      # Full schema (19 tables)
├── 000001_init_schema.down.sql    # Drop all tables
├── 000002_api_keys.up.sql         # API keys table
├── 000002_api_keys.down.sql
├── 000003_api_key_permission.up.sql  # manage_api_keys permission
└── 000003_api_key_permission.down.sql
```

Files are embedded via `//go:embed` in `db/embed.go` and loaded at startup.

### Numbering Scheme

Files follow the pattern `NNNNNN_description.{up|down}.sql` with a sequential
6-digit prefix. The `up` file applies the migration, the `down` file reverses
it.

### How Migrations Run

1. `InitialiseDatabase()` in `db/db.go` opens the SQLite connection
2. `runMigrations()` creates an iofs source from the embedded filesystem
3. `migrate.NewWithInstance()` creates the migrator
4. `m.Up()` applies all pending migrations (idempotent — skips already-applied)
5. golang-migrate tracks applied versions in a `schema_migrations` table

If a migration fails mid-way, the database is marked "dirty" and the
application will refuse to start until the issue is resolved manually.

### Writing a New Migration

1. Create two files following the next number in sequence:
   ```
   000004_your_description.up.sql
   000004_your_description.down.sql
   ```
2. Write the SQL. The `up` file should be idempotent where possible
   (use `IF NOT EXISTS`, `INSERT OR IGNORE`, etc.)
3. The `down` file should reverse the `up` file completely
4. Test by running the application — migrations execute automatically on startup

## Repository Layer

### Interface Pattern

Each repository defines a Go interface and a concrete implementation:

```go
// Interface (db/sensor_repository_interface.go)
type SensorRepositoryInterface[T any] interface {
    AddSensor(ctx context.Context, sensor T) error
    GetSensorByName(ctx context.Context, name string) (*T, error)
    GetSensorsByDriver(ctx context.Context, sensorDriver string) ([]T, error)
    // ...
}

// Implementation (db/sensor_repository.go)
type SensorRepository struct {
    db     *sql.DB
    logger *slog.Logger
}

func NewSensorRepository(db *sql.DB, logger *slog.Logger) *SensorRepository {
    return &SensorRepository{
        db:     db,
        logger: logger.With("component", "sensor_repository"),
    }
}
```

### Repositories

| Repository | Interface | Tables |
|-----------|-----------|--------|
| `SensorRepository` | `SensorRepositoryInterface[T]` | sensors, sensor_health_history |
| `ReadingsRepository` | `ReadingsRepository` | readings |
| `AlertRepositoryImpl` | `AlertRepository` | sensor_alert_rules, alert_sent_history |
| `SqlUserRepository` | `UserRepository` | users, user_roles |
| `SqlSessionRepository` | `SessionRepository` | sessions, session_audit |
| `SqlFailedLoginRepository` | `FailedLoginRepository` | failed_login_attempts, failed_login_summary |
| `SqlRoleRepository` | `RoleRepository` | roles, permissions, role_permissions |
| `SqlApiKeyRepository` | `ApiKeyRepository` | api_keys |
| `SqlNotificationRepository` | `NotificationRepository` | notifications, user_notifications, notification_channel_defaults, notification_channel_preferences |

### Query Conventions

- All methods take `ctx context.Context` as the first parameter
- All string-equality WHERE clauses use `LOWER(col) = LOWER(?)` for
  case-insensitive matching
- Errors are wrapped with `fmt.Errorf("context: %w", err)` for traceability
- Logging is at DEBUG level for successful operations, ERROR for failures
- Some repositories depend on other repositories (e.g. `ReadingsRepository`
  uses `SensorRepository` to look up sensor IDs by name)
