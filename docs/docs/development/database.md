# Database

This guide covers the database system used by Sensor Hub.

## Overview

Sensor Hub uses SQLite.

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

**Important:** Use the custom `SQLiteTime` and `NullSQLiteTime` types from
`db/sqlite_time.go` when scanning datetime columns — the SQLite driver returns
TEXT, not `time.Time`, so `sql.NullTime` will not work.

## Migration System

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate) with
SQL files embedded into the binary.

### File Location

Migration files live in `sensor_hub/db/migrations/` following the naming pattern
`NNNNNN_description.{up|down}.sql`. The `up` file applies the migration, the
`down` file reverses it.

### How Migrations Run

Migrations run automatically on startup and are idempotent — already-applied
versions are skipped. Applied versions are tracked in a `schema_migrations`
table. If a migration fails mid-way, the database is marked "dirty" and the
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
