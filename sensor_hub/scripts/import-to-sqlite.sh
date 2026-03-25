#!/bin/bash
set -euo pipefail

# import-to-sqlite.sh
# Imports a MySQL data dump (produced by dump-mysql-data.sh) into the new
# SQLite database. The schema must already be initialised (start the new
# sensor-hub application once, then stop it before running this script).
#
# Usage:
#   ./import-to-sqlite.sh <sqlite-db-path> <sql-dump-file>
#
# Examples:
#   ./import-to-sqlite.sh data/sensor_hub.db mysql_export.sql
#   ./import-to-sqlite.sh /opt/sensor-hub/data/sensor_hub.db mysql_export.sql

DB_PATH="${1:?Usage: $0 <sqlite-db-path> <sql-dump-file>}"
SQL_FILE="${2:?Usage: $0 <sqlite-db-path> <sql-dump-file>}"

# --- Pre-flight checks ---

if [ ! -f "$DB_PATH" ]; then
    echo "ERROR: SQLite database not found at '$DB_PATH'"
    echo ""
    echo "The database file is created when the new sensor-hub starts for the"
    echo "first time. Start the application once, let the schema initialise,"
    echo "then stop it and re-run this script."
    exit 1
fi

if [ ! -f "$SQL_FILE" ]; then
    echo "ERROR: SQL dump file not found at '$SQL_FILE'"
    exit 1
fi

if ! command -v sqlite3 &>/dev/null; then
    echo "ERROR: sqlite3 CLI is required but not found."
    echo "Install with your package manager, e.g.:"
    echo "  Fedora/RHEL: sudo dnf install sqlite"
    echo "  Debian/Ubuntu: sudo apt install sqlite3"
    exit 1
fi

# Quick schema sanity check — make sure the expected tables exist
EXPECTED_TABLE="sensors"
if ! sqlite3 "$DB_PATH" "SELECT 1 FROM $EXPECTED_TABLE LIMIT 0;" 2>/dev/null; then
    echo "ERROR: Table '$EXPECTED_TABLE' not found in '$DB_PATH'."
    echo "The schema does not appear to be initialised. Start the new"
    echo "sensor-hub application once to create the schema, then stop it."
    exit 1
fi

echo "=== SQLite Data Import ==="
echo "Database: $DB_PATH"
echo "SQL file: $SQL_FILE ($(du -h "$SQL_FILE" | cut -f1))"
echo ""

# --- Backup ---

BACKUP="${DB_PATH}.pre-import-$(date +%Y%m%d%H%M%S)"
cp "$DB_PATH" "$BACKUP"
echo "Backup created: $BACKUP"
echo ""

# --- Clear seed data ---
#
# The schema migration inserts seed data (roles, permissions, defaults).
# The dump file contains the authoritative version of this data from MySQL,
# so we clear everything first to avoid primary-key conflicts.

echo "Clearing existing data (seed rows will be replaced by the import)..."

sqlite3 "$DB_PATH" <<'CLEAR_SQL'
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- Delete in reverse dependency order to avoid FK violations
DELETE FROM notification_channel_preferences;
DELETE FROM notification_channel_defaults;
DELETE FROM user_notifications;
DELETE FROM notifications;
DELETE FROM alert_sent_history;
DELETE FROM sensor_alert_rules;
DELETE FROM sensor_health_history;
DELETE FROM hourly_avg_temperature;
DELETE FROM temperature_readings;
DELETE FROM role_permissions;
DELETE FROM user_roles;
DELETE FROM users;
DELETE FROM permissions;
DELETE FROM roles;
DELETE FROM sensors;

COMMIT;
PRAGMA foreign_keys = ON;
CLEAR_SQL

echo "Done."
echo ""

# --- Import ---

echo "Importing data..."

if sqlite3 "$DB_PATH" < "$SQL_FILE"; then
    echo "Import succeeded."
else
    echo ""
    echo "ERROR: Import failed! Restoring database from backup..."
    cp "$BACKUP" "$DB_PATH"
    echo "Database restored to pre-import state."
    echo "The dump file may contain MySQL-specific SQL that was not"
    echo "correctly translated. Check the error above for details."
    exit 1
fi

# --- Verify ---

echo ""
echo "Verification — row counts:"

TABLES=(
    sensors
    temperature_readings
    hourly_avg_temperature
    sensor_health_history
    roles
    permissions
    users
    user_roles
    role_permissions
    sensor_alert_rules
    alert_sent_history
    notifications
    user_notifications
    notification_channel_defaults
    notification_channel_preferences
)

TOTAL=0
for TABLE in "${TABLES[@]}"; do
    COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM $TABLE;" 2>/dev/null || echo "ERR")
    printf "  %-42s %s\n" "$TABLE" "$COUNT"
    if [[ "$COUNT" =~ ^[0-9]+$ ]]; then
        TOTAL=$((TOTAL + COUNT))
    fi
done

echo ""
echo "Total rows imported: $TOTAL"
echo ""
echo "Import complete. You can now start the sensor-hub application."
echo ""
echo "If anything looks wrong, restore from the backup:"
echo "  cp '$BACKUP' '$DB_PATH'"
echo ""
echo "Once satisfied, delete the backup:"
echo "  rm '$BACKUP'"
