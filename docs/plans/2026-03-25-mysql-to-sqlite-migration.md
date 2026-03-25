# MySQL to SQLite Migration — Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Replace MySQL with embedded SQLite to eliminate external database dependencies, enabling a single-binary deployment.

**Architecture:** Replace the MySQL driver with `modernc.org/sqlite` (pure Go), replace Flyway with `golang-migrate/migrate` using `//go:embed` for baked-in migrations, rewrite all 18 Flyway changesets as a single consolidated SQLite schema + seed migration, translate all MySQL-specific SQL in 8 repository files to SQLite-compatible SQL, replace the MySQL EVENT-driven hourly averaging with a Go scheduled task, simplify Docker Compose to remove MySQL/Flyway containers, and replace `database.properties` config with a single `database.path` property.

**Tech Stack:** Go 1.25, `modernc.org/sqlite`, `golang-migrate/migrate/v4`, `database/sql`, `//go:embed`

---

### Task 1: Install Go dependencies

**Files:**
- Modify: `sensor_hub/go.mod`

**Step 1: Add SQLite driver and migration library**

Run from `sensor_hub/`:
```bash
go get modernc.org/sqlite
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/sqlite
go get github.com/golang-migrate/migrate/v4/source/iofs
```

**Step 2: Remove MySQL driver**

```bash
go get -u  # tidy first
```

Do NOT remove `github.com/go-sql-driver/mysql` from go.mod yet — we'll do that after all code changes are complete so the build doesn't break mid-migration.

**Step 3: Verify go.mod**

Run: `go mod tidy`
Expected: No errors. Both `modernc.org/sqlite` and `github.com/golang-migrate/migrate/v4` appear in go.mod.

---

### Task 2: Create consolidated SQLite migration

**Files:**
- Create: `sensor_hub/db/migrations/000001_init_schema.up.sql`
- Create: `sensor_hub/db/migrations/000001_init_schema.down.sql`

**Step 1: Create migrations directory**

```bash
mkdir -p sensor_hub/db/migrations
```

**Step 2: Write the consolidated up migration**

Create `sensor_hub/db/migrations/000001_init_schema.up.sql` containing all 18 Flyway changesets merged into a single SQLite-compatible schema. Key translations from MySQL:

| MySQL | SQLite |
|-------|--------|
| `INT AUTO_INCREMENT PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `BIGINT AUTO_INCREMENT PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `SERIAL PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `FLOAT(4)` | `REAL` |
| `VARCHAR(N)` | `TEXT` |
| `DATETIME` | `TEXT` (ISO 8601 strings) |
| `TIMESTAMP DEFAULT CURRENT_TIMESTAMP` | `TEXT DEFAULT (datetime('now'))` |
| `TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | `TEXT DEFAULT (datetime('now'))` (updates handled in Go) |
| `BOOLEAN` | `INTEGER` (0/1) |
| `JSON` | `TEXT` |
| `UNIQUE KEY name (cols)` | `UNIQUE (cols)` |
| `INDEX idx_name (col)` | `CREATE INDEX idx_name ON table(col)` |
| `FOREIGN KEY ... ON DELETE CASCADE` | Same (with `PRAGMA foreign_keys = ON`) |
| MySQL EVENT (hourly avg) | Omitted — moved to Go Task 12 |

Tables (final state after all 18 migrations merged):
1. `sensors` — id, name, type, url, health_status, health_reason, enabled
2. `temperature_readings` — id, sensor_id (FK), time, temperature
3. `hourly_avg_temperature` — id, sensor_id (FK), time, average_temperature; UNIQUE(sensor_id, time)
4. `sensor_health_history` — id, sensor_id (FK), health_status, recorded_at
5. `users` — id, username, email, password_hash, must_change_password, disabled, created_at, updated_at
6. `roles` — id, name, description, created_at; seed: admin, user, viewer
7. `user_roles` — user_id, role_id; composite PK; FKs with CASCADE
8. `sessions` — id, user_id (FK CASCADE), token_hash, csrf_token, created_at, expires_at, last_accessed_at, ip_address, user_agent
9. `failed_login_attempts` — id, username, user_id (FK SET NULL), ip_address, attempt_time, reason
10. `failed_login_summary` — identifier PK, attempts, last_attempt_at
11. `permissions` — id, name, description; seed all 16 permissions
12. `role_permissions` — role_id, permission_id; composite PK; FKs with CASCADE; seed admin→all
13. `session_audit` — id, session_id (FK CASCADE), revoked_by_user_id, event_type, reason, created_at
14. `sensor_alert_rules` — id, sensor_id (FK CASCADE), alert_type, high/low_threshold, trigger_status, enabled, rate_limit_hours, created_at, updated_at
15. `alert_sent_history` — id, alert_rule_id (FK CASCADE), sensor_id (FK CASCADE), sent_at, alert_reason, reading_value, reading_status
16. `notifications` — id, category, severity, title, message, metadata (TEXT/JSON), created_at
17. `user_notifications` — id, user_id (FK CASCADE), notification_id (FK CASCADE), is_read, is_dismissed, read_at, dismissed_at; UNIQUE(user_id, notification_id)
18. `notification_channel_defaults` — id, category (UNIQUE), email_enabled, inapp_enabled; seed 3 rows
19. `notification_channel_preferences` — id, user_id (FK CASCADE), category, email_enabled, inapp_enabled; UNIQUE(user_id, category)

**Important:** Start the file with `PRAGMA foreign_keys = ON;` and `PRAGMA journal_mode = WAL;` (Write-Ahead Logging for concurrent read performance).

**Step 3: Write the down migration**

Create `sensor_hub/db/migrations/000001_init_schema.down.sql` with `DROP TABLE IF EXISTS` for all tables in reverse dependency order.

**Step 4: Validate SQL syntax**

Run: `sqlite3 :memory: < sensor_hub/db/migrations/000001_init_schema.up.sql`
Expected: No errors.

---

### Task 3: Rewrite `db/db.go` — database initialization with embedded migrations

**Files:**
- Modify: `sensor_hub/db/db.go`
- Create: `sensor_hub/db/embed.go` (embed directive for migrations)

**Step 1: Create embed.go with migration embedding**

Create `sensor_hub/db/embed.go`:
```go
package database

import "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS
```

**Step 2: Rewrite db.go**

Replace the entire file. New implementation:
- Import `modernc.org/sqlite` driver (instead of `github.com/go-sql-driver/mysql`)
- Read `database.path` from `appProps.AppConfig.DatabasePath` (new field, replaces hostname/port/user/pass)
- Default path: `data/sensor_hub.db`
- Open with: `sql.Open("sqlite", dsn)` where dsn is the file path with `?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)`
- Run golang-migrate after opening:
  ```go
  import (
      "github.com/golang-migrate/migrate/v4"
      sqlite_migrate "github.com/golang-migrate/migrate/v4/database/sqlite"
      "github.com/golang-migrate/migrate/v4/source/iofs"
  )
  ```
  - Create iofs source from `migrationsFS` (the embedded FS)
  - Create sqlite driver instance
  - Run `m.Up()` — applies all pending migrations
  - Handle `migrate.ErrNoChange` gracefully (log "schema up to date")
- Ensure parent directory of DB path exists (create with `os.MkdirAll`)

**Step 3: Verify compilation**

Run: `go build ./db/...`
Expected: Compiles (other packages may fail until updated — that's fine).

### Task 4: Rewrite `db/sensor_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/sensor_repository.go`

**MySQL→SQLite translations needed in this file:**

| Line(s) | Change |
|---------|--------|
| All `fmt.Sprintf` table name injections | Keep as-is (safe — uses constants) |
| Health history INSERT in async goroutine | Keep as-is (standard INSERT, no MySQL-specific syntax) |

This file uses standard SQL throughout. The only concern is ensuring `BOOLEAN` columns are handled as `INTEGER` 0/1, which Go's `database/sql` already does transparently. **No SQL query changes needed** — just verify all queries work.

**Step 1: Review and verify** — Read through every query, confirm no MySQL-specific syntax.

**Step 2: Run existing tests** (after Task 15 updates the test expectations)

---

### Task 5: Rewrite `db/temperature_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/temperature_repository.go`

**Changes needed:** None — all queries use standard SQL (`INSERT INTO`, `SELECT...JOIN...WHERE...BETWEEN...ORDER BY`, `DELETE FROM...WHERE`, transactions). The `?` placeholder works in SQLite.

**Step 1: Review and verify** — Confirm all queries. The `BETWEEN` clause with datetime strings works in SQLite when stored as ISO 8601 text.

---

### Task 6: Rewrite `db/user_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/user_repository.go`

**MySQL→SQLite translations:**

| Location | MySQL | SQLite |
|----------|-------|--------|
| Line 138 | `INSERT IGNORE INTO user_roles` | `INSERT OR IGNORE INTO user_roles` |
| Line 249 | `INSERT IGNORE INTO user_roles` (in transaction) | `INSERT OR IGNORE INTO user_roles` |

2 changes total. All other queries (SELECT, UPDATE, DELETE, transactions) are standard SQL.

**Step 1: Replace `INSERT IGNORE` → `INSERT OR IGNORE`** (2 occurrences)

**Step 2: Verify `sql.NullTime` scanning** — SQLite stores timestamps as TEXT. The `modernc.org/sqlite` driver handles `time.Time` scanning from TEXT columns. Verify `updated_at` NULL handling still works.

---

### Task 7: Rewrite `db/session_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/session_repository.go`

**Changes needed:** None. All queries use standard SQL with `?` placeholders. Timestamps are passed as Go `time.Time` values. The SHA256 hashing and CSRF generation are pure Go. No MySQL-specific syntax.

**Step 1: Review and verify** — Confirm all queries are SQLite-compatible.

---

### Task 8: Rewrite `db/alert_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/alert_repository.go`

**Changes needed:** None. The complex JOINs, LEFT JOINs, subqueries with `MAX()`, `GROUP BY`, and `HAVING` are all standard SQL supported by SQLite. No `INSERT IGNORE`, no `NOW()`, no MySQL-specific features.

**Step 1: Review and verify** — Confirm the subquery pattern `SELECT alert_rule_id, MAX(sent_at) as sent_at FROM alert_sent_history GROUP BY alert_rule_id` works in SQLite (it does).

---

### Task 9: Rewrite `db/notification_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/notification_repository.go`

**MySQL→SQLite translations:**

| Line | MySQL | SQLite |
|------|-------|--------|
| 70 | `INSERT IGNORE INTO user_notifications` | `INSERT OR IGNORE INTO user_notifications` |
| 78 | `INSERT IGNORE INTO user_notifications` (INSERT...SELECT) | `INSERT OR IGNORE INTO user_notifications` |
| 197 | `read_at = NOW()` | `read_at = datetime('now')` |
| 205 | `dismissed_at = NOW()` | `dismissed_at = datetime('now')` |
| 213 | `read_at = NOW()` (bulk) | `read_at = datetime('now')` |
| 221 | `dismissed_at = NOW()` (bulk) | `dismissed_at = datetime('now')` |
| 283-286 | `INSERT...ON DUPLICATE KEY UPDATE email_enabled = VALUES(email_enabled), inapp_enabled = VALUES(inapp_enabled)` | `INSERT INTO...ON CONFLICT(user_id, category) DO UPDATE SET email_enabled = excluded.email_enabled, inapp_enabled = excluded.inapp_enabled` |

7 changes total. This is the file with the most MySQL-specific syntax.

**Step 1: Replace `INSERT IGNORE` → `INSERT OR IGNORE`** (2 occurrences)

**Step 2: Replace `NOW()` → `datetime('now')`** (4 occurrences)

**Step 3: Rewrite the upsert query** — Replace `ON DUPLICATE KEY UPDATE...VALUES()` with SQLite's `ON CONFLICT...DO UPDATE SET...excluded.` syntax.

---

### Task 10: Rewrite `db/role_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/role_repository.go`

**MySQL→SQLite translations:**

| Line | MySQL | SQLite |
|------|-------|--------|
| 112 | `INSERT IGNORE INTO role_permissions` | `INSERT OR IGNORE INTO role_permissions` |

1 change total.

**Step 1: Replace `INSERT IGNORE` → `INSERT OR IGNORE`**

---

### Task 11: Rewrite `db/failed_login_repository.go` — SQLite-compatible SQL

**Files:**
- Modify: `sensor_hub/db/failed_login_repository.go`

**Changes needed:** None. All queries use standard SQL with `?` placeholders and Go-side `time.Now()` for timestamps. No MySQL-specific syntax.

**Step 1: Review and verify.**

### Task 12: Implement hourly average computation in Go

**Files:**
- Modify: `sensor_hub/db/temperature_repository.go` (add `ComputeHourlyAverages` method)
- Modify: `sensor_hub/db/readings_repository_interface.go` (add method to interface)
- Modify: `sensor_hub/service/cleanup_service.go` (add hourly average scheduling)

**Context:** MySQL used a database EVENT (`hourly_average_temperature_event`) that ran every hour to compute `AVG(temperature)` grouped by sensor and hour, inserting results into `hourly_avg_temperature`. This must now be a Go-side scheduled task.

**Step 1: Add `ComputeHourlyAverages` to temperature repository**

Add a new method to `TemperatureRepository`:
```go
func (t *TemperatureRepository) ComputeHourlyAverages() error {
    query := `
        INSERT OR IGNORE INTO hourly_avg_temperature (sensor_id, time, average_temperature)
        SELECT
            tr.sensor_id,
            strftime('%Y-%m-%d %H:00:00', tr.time) AS hour,
            ROUND(AVG(tr.temperature), 2) AS avg_temp
        FROM temperature_readings tr
        WHERE tr.time >= strftime('%Y-%m-%d %H:00:00', datetime('now', '-1 hour'))
          AND tr.time < strftime('%Y-%m-%d %H:00:00', datetime('now'))
        GROUP BY tr.sensor_id, hour
    `
    _, err := t.db.Exec(query)
    return err
}
```

This is equivalent to the MySQL EVENT but uses `INSERT OR IGNORE` (relying on the UNIQUE constraint on `(sensor_id, time)`) instead of `HAVING NOT EXISTS`.

**Step 2: Add `ComputeHourlyAverages()` to the `ReadingsRepository` interface**

**Step 3: Add hourly average ticker to cleanup_service.go**

In `StartPeriodicCleanup`, add a second goroutine with a 1-hour ticker that calls `temperatureRepo.ComputeHourlyAverages()`. This replaces the MySQL EVENT. Optionally also run once at startup to catch any missed hours.

**Step 4: Write test for ComputeHourlyAverages**

Add test in `sensor_hub/db/temperature_repository_test.go` using sqlmock.

---

### Task 13: Update application properties — replace DB config with file path

**Files:**
- Modify: `sensor_hub/application_properties/application_configuration.go`
- Modify: `sensor_hub/application_properties/application_properties_files.go`
- Modify: `sensor_hub/application_properties/application_properties_defaults.go`
- Modify: `sensor_hub/application_properties/sensitive_properties.go`
- Modify: `sensor_hub/service/properties_service.go`
- Modify: `sensor_hub/configuration/database.properties`

**Step 1: Replace 4 DB config fields with 1 in `ApplicationConfiguration`**

Remove:
```go
DatabaseUsername string
DatabasePassword string
DatabaseHostname string
DatabasePort     string
```

Replace with:
```go
DatabasePath string
```

**Step 2: Update defaults**

In `application_properties_defaults.go`, replace:
```go
var DatabasePropertiesDefaults = map[string]string{
    "database.path": "data/sensor_hub.db",
}
```

**Step 3: Update `LoadConfigurationFromMaps` and `ConvertConfigurationToMaps`**

Replace the 4 database property reads/writes with `database.path`.

**Step 4: Update `dbValidateDatabaseProperties`**

Replace the 4-field validation with a single check that `database.path` is not empty.

**Step 5: Remove all Set/Get methods for old DB fields**

Remove `SetDatabaseUsername`, `SetDatabasePassword`, `SetDatabaseHostname`, `SetDatabasePort`. Add `SetDatabasePath`.

**Step 6: Remove `database.password` from `SensitivePropertiesKeys`**

The file path is not sensitive.

**Step 7: Update `SaveConfigurationToFiles`**

Replace the 4 database property write lines with a single `database.path` line.

**Step 8: Update `configuration/database.properties`**

Replace contents with:
```
database.path = data/sensor_hub.db
```

**Step 9: Remove environment variable overrides for old DB fields**

Search for `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS` in `application_properties_files.go` or wherever env overrides are applied — replace with `DB_PATH` if env override is desired.

**Step 10: Update properties_service.go if needed** — The existing code dynamically handles whatever keys exist in the maps, so it should work without changes after the maps are updated.

---

### Task 14: Update Docker Compose files — remove MySQL & Flyway

**Files:**
- Modify: `sensor_hub/docker/docker-compose.yml`
- Modify: `sensor_hub/docker_tests/docker-compose.yml`
- Modify: `sensor_hub/docker/sensor-hub.dockerfile`
- Modify: `sensor_hub/docker_tests/sensor-hub.dockerfile`
- Delete or archive: `sensor_hub/docker/wait-for-mysql.sh`
- Delete or archive: `sensor_hub/docker/write-db-properties-and-wait.sh`
- Delete or archive: `sensor_hub/docker_tests/wait-for-mysql.sh`
- Delete or archive: `sensor_hub/docker_tests/recover-hourly-averages.sh`

**Step 1: Rewrite `docker/docker-compose.yml`**

Remove: `mysql` service, `flyway` service.
Remove from `sensor-hub`: `depends_on` for mysql/flyway, `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASS` env vars.
Add: Volume mount for SQLite data directory (e.g., `./data:/app/data`).
Add: `DB_PATH` environment variable if env override is desired.

**Step 2: Rewrite `docker_tests/docker-compose.yml`**

Same changes as production compose. Remove mysql, flyway, related env vars. Add data volume.

**Step 3: Update Dockerfiles**

In both `sensor-hub.dockerfile` and `docker_tests/sensor-hub.dockerfile`:
- Remove `RUN chmod +x ./docker/wait-for-mysql.sh` (or docker_tests equivalent)
- Remove `ENTRYPOINT ["/app/docker/wait-for-mysql.sh"]`
- Change to direct entrypoint: `ENTRYPOINT ["sensor-hub"]` (production) or keep air for dev

**Step 4: Delete MySQL helper scripts**

Remove `wait-for-mysql.sh` from both docker/ and docker_tests/. Remove `write-db-properties-and-wait.sh`. Remove `recover-hourly-averages.sh` (functionality now built into Go app via Task 12).

---

### Task 15: Update all repository tests for SQLite SQL patterns

**Files:**
- Modify: `sensor_hub/db/user_repository_test.go`
- Modify: `sensor_hub/db/notification_repository_test.go`
- Modify: `sensor_hub/db/role_repository_test.go`
- Modify: `sensor_hub/db/temperature_repository_test.go` (new test for hourly averages)

**Step 1: Update sqlmock expectations for `INSERT IGNORE` → `INSERT OR IGNORE`**

In `user_repository_test.go`: Update all `mock.ExpectExec("INSERT IGNORE INTO user_roles")` → `mock.ExpectExec("INSERT OR IGNORE INTO user_roles")`.

In `notification_repository_test.go`: Update `INSERT IGNORE INTO user_notifications` expectations.

In `role_repository_test.go`: Update `INSERT IGNORE INTO role_permissions` expectations.

**Step 2: Update sqlmock expectations for `NOW()` → `datetime('now')`**

In `notification_repository_test.go`: Update any mock expectations that match on `NOW()`.

**Step 3: Update upsert test expectations**

In `notification_repository_test.go`: Update the `ON DUPLICATE KEY UPDATE` mock expectation to match `ON CONFLICT...DO UPDATE SET`.

**Step 4: Add test for `ComputeHourlyAverages`**

**Step 5: Run all tests**

Run: `cd sensor_hub && go test ./db/...`
Expected: All tests pass.

---

### Task 16: Build and run full test suite

**Files:** None (validation only)

**Step 1: Remove MySQL driver from go.mod**

Now that all code has been migrated:
```bash
cd sensor_hub
go mod edit -droprequire github.com/go-sql-driver/mysql
go mod tidy
```

**Step 2: Build the entire project**

Run: `cd sensor_hub && go build ./...`
Expected: Clean build with no errors.

**Step 3: Run all tests**

Run: `cd sensor_hub && go test ./...`
Expected: All tests pass.

**Step 4: Verify the binary runs**

Run: `cd sensor_hub && go build -o sensor-hub . && ls -la sensor-hub`
Expected: Single binary file exists.

---

### Task 17: Update documentation

**Files:**
- Modify: `docs/docs/installation.md`
- Modify: `docs/docs/configuration.md`
- Modify: `docs/docs/upgrading.md`
- Modify: `docs/docs/prerequisites.md`
- Modify: `docs/docs/overview.md`
- Modify: `sensor_hub/README.md`

**Step 1: Update overview.md** — Replace "MySQL 8 Database" references with "SQLite (embedded)". Update the architecture description to mention single-binary deployment.

**Step 2: Update prerequisites.md** — Remove MySQL/Docker Compose requirements for database. Simplify to just Docker for the Go app + UI containers (or native binary).

**Step 3: Update installation.md** — Remove `database.properties` 4-field setup. Replace with `database.path` single field. Remove Flyway migration step from startup sequence. Simplify Docker Compose instructions.

**Step 4: Update configuration.md** — Replace database properties section (4 fields → 1 field `database.path`). Remove password masking mention. Update environment variables section.

**Step 5: Update upgrading.md** — Replace Flyway migration description with golang-migrate embedded migration. Note that the binary handles schema upgrades automatically at startup. Remove "backup MySQL data volume" step — replace with "backup SQLite database file".

**Step 6: Update sensor_hub/README.md** — Reflect simplified setup.
