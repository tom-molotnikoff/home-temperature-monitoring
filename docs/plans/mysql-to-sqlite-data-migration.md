# Migrating Data from MySQL to SQLite

> **One-off migration guide.** This document covers moving your existing sensor
> hub data from the MySQL database to the new embedded SQLite database. You only
> need to do this once.

## Overview

The old deployment stores all data in a MySQL 8 container managed by Docker
Compose. The new version uses an embedded SQLite database — a single file on
disk with no external processes.

This guide walks through exporting your MySQL data, deploying the new version,
and importing the data into SQLite.

## What gets migrated

| Category | Tables | Migrated? |
|----------|--------|-----------|
| **Sensors & readings** | `sensors`, `temperature_readings`, `hourly_avg_temperature`, `sensor_health_history` | ✅ Yes |
| **Users & access control** | `users`, `roles`, `permissions`, `user_roles`, `role_permissions` | ✅ Yes |
| **Alerts** | `sensor_alert_rules`, `alert_sent_history` | ✅ Yes |
| **Notifications** | `notifications`, `user_notifications`, `notification_channel_defaults`, `notification_channel_preferences` | ✅ Yes |
| **Sessions** | `sessions`, `session_audit` | ❌ Skipped — users log in again |
| **Security** | `failed_login_attempts`, `failed_login_summary` | ❌ Skipped — ephemeral rate-limit data |
| **Flyway** | `flyway_schema_history` | ❌ Skipped — no longer used |

Password hashes are preserved, so existing user credentials continue to work.

## Prerequisites

- `docker` CLI (to talk to the running MySQL container)
- `sqlite3` CLI (to run the import)
  - Install: `sudo apt install sqlite3` (Debian/Ubuntu) or `brew install sqlite` (macOS)
- The scripts in `sensor_hub/scripts/`:
  - `dump-mysql-data.sh`
  - `import-to-sqlite.sh`

## Migration steps

### 1. Stop the sensor-hub application

Stop the sensor-hub Go process (or container) so no new data is written during
the export. **Keep the MySQL container running** — the dump script needs it.

```bash
# If using Docker Compose (old deployment):
cd sensor_hub/docker            # or docker_tests/
docker compose stop sensor-hub  # stops app, leaves MySQL running
```

If running the binary directly, stop the process however you normally would.

### 2. Export data from MySQL

Find your MySQL container name:

```bash
docker ps --filter "ancestor=mysql:8" --format '{{.Names}}'
```

Run the dump script:

```bash
cd sensor_hub/scripts
./dump-mysql-data.sh <mysql-container-name> [output-file] [db-user] [db-password]
```

The `db-user` and `db-password` arguments default to `root` / `password` if
omitted. This produces `mysql_export.sql` in the current directory. The script
prints per-table row counts — verify these look reasonable.

> **Example with custom credentials and output path:**
> `./dump-mysql-data.sh my-mysql mysql_backup.sql dbadmin s3cret`

### 3. Tear down the old deployment

Now that the data is safely exported, shut down the old stack:

```bash
cd sensor_hub/docker
docker compose down
```

> **Don't delete the MySQL data volume yet.** Keep it as a safety net until
> you've verified the import. You can remove it later with
> `docker volume rm <volume-name>`.

### 4. Deploy the new version

Deploy the new version of sensor-hub however you intend to run it — Docker
Compose, a `.deb`/`.rpm` package, or as a standalone binary.

The key detail: **start the application once and let it initialise.** On first
start, it will:

1. Create the SQLite database file (default: `data/sensor_hub.db`)
2. Run the embedded schema migrations to set up all tables and seed data

You'll see log output like:

```
sensor-hub: Connected to database
sensor-hub: Applied migration 000001_init_schema
```

### 5. Stop the application

Stop sensor-hub so we can import data into the SQLite file without contention:

```bash
# However you stop it — Ctrl+C, systemctl stop, docker compose stop, etc.
```

### 6. Import data into SQLite

```bash
cd sensor_hub/scripts
./import-to-sqlite.sh <path-to-sensor_hub.db> mysql_export.sql
```

The `<path-to-sensor_hub.db>` depends on how you deployed:

| Deployment | Typical path |
|------------|-------------|
| Running from repo | `../data/sensor_hub.db` |
| Docker Compose | The volume-mounted data directory |
| System package | `/var/lib/sensor-hub/sensor_hub.db` or per config |

The script will:

1. **Back up** the empty database (just in case)
2. **Clear** seed data to avoid primary-key conflicts with the import
3. **Import** all rows from the dump
4. **Verify** by printing row counts for every table

Example output:

```
=== SQLite Data Import ===
Database: ../data/sensor_hub.db
SQL file: mysql_export.sql (2.4M)

Backup created: ../data/sensor_hub.db.pre-import-20260325112000
Clearing existing data (seed rows will be replaced by the import)...
Done.

Importing data...
Import succeeded.

Verification — row counts:
  sensors                                    4
  temperature_readings                       312847
  hourly_avg_temperature                     26071
  ...
```

### 7. Start the application and verify

Start sensor-hub again. It will detect the existing database and skip migration
(schema is already up to date).

Verify:

- Log in with your existing credentials
- Check the Sensors Overview page — all sensors should appear
- Check temperature graphs — historical data should render
- Check alert rules — should match your old configuration

### 8. Clean up

Once you're satisfied everything works:

```bash
# Remove the old MySQL data volume
docker volume ls                        # find the volume name
docker volume rm <mysql-data-volume>

# Delete the pre-import backup
rm <path-to-sensor_hub.db>.pre-import-*

# Delete the SQL dump
rm mysql_export.sql

# (Optional) Remove these migration scripts — they're no longer needed
rm sensor_hub/scripts/dump-mysql-data.sh
rm sensor_hub/scripts/import-to-sqlite.sh
```

## Troubleshooting

### Import fails with a SQL error

The dump script transforms `mysqldump` output into SQLite-compatible SQL. If a
value contains unusual characters that were escaped differently, you may see an
error. Open `mysql_export.sql`, find the offending line (the error message
includes a line number), and fix the escaping manually. Then re-run the import
— the script automatically restores from the backup on failure.

### Row counts don't match

Compare the dump script's per-table counts with the import script's
verification output. If any table has fewer rows after import, check for
primary-key conflicts or constraint violations in the SQLite error output.

### "Table not found" error during import

The schema wasn't initialised. Start the new sensor-hub application once (step
4) and let it create the database, then stop it and retry the import.

### Want to start over

Restore the backup:

```bash
cp <path-to-sensor_hub.db>.pre-import-* <path-to-sensor_hub.db>
```

Or delete the database file entirely and restart the app to reinitialise with
a fresh schema, then re-run the import.

### Rollback to MySQL entirely

If something goes seriously wrong, you still have the MySQL data volume. Just
redeploy the old version with `docker compose up` and it will reconnect to the
existing data.
