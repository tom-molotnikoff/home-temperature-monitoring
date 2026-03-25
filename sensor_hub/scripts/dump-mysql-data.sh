#!/bin/bash
set -euo pipefail

# dump-mysql-data.sh
# Exports sensor hub data from the MySQL database (running in Docker) to a
# SQLite-compatible SQL file. Run this BEFORE tearing down the old deployment.
#
# Usage:
#   ./dump-mysql-data.sh <mysql-container-name> [output-file] [db-user] [db-password]
#
# Examples:
#   ./dump-mysql-data.sh docker-sensor-hub-mysql-1
#   ./dump-mysql-data.sh docker-sensor-hub-mysql-1 my_backup.sql myuser mypassword
#
# The MySQL container must be running. Find the name with:  docker ps

CONTAINER="${1:?Usage: $0 <mysql-container-name> [output-file] [db-user] [db-password]}"
OUTPUT="${2:-mysql_export.sql}"
DB_USER="${3:-root}"
DB_PASS="${4:-password}"

DB_NAME="sensor_database"

# Tables to export, in foreign-key dependency order (parents before children).
# Excluded tables (ephemeral data, recreated automatically):
#   - sessions / session_audit  (users simply log in again)
#   - failed_login_attempts     (brute-force tracking, ephemeral)
#   - failed_login_summary      (ephemeral)
#   - flyway_schema_history     (Flyway metadata, no longer needed)
TABLES=(
    roles
    permissions
    sensors
    users
    user_roles
    role_permissions
    temperature_readings
    hourly_avg_temperature
    sensor_health_history
    sensor_alert_rules
    alert_sent_history
    notifications
    user_notifications
    notification_channel_defaults
    notification_channel_preferences
)

echo "=== MySQL → SQLite Data Export ==="
echo "Container: $CONTAINER"
echo "Database:  $DB_NAME"
echo "Output:    $OUTPUT"
echo ""

# Verify the container is actually running
if ! docker inspect --format='{{.State.Running}}' "$CONTAINER" 2>/dev/null | grep -q true; then
    echo "ERROR: Container '$CONTAINER' is not running."
    echo "Hint:  Run 'docker ps' to find the MySQL container name."
    exit 1
fi

RAW_FILE=$(mktemp)
trap 'rm -f "$RAW_FILE"' EXIT

echo "Dumping tables..."

# --compatible=ansi     → double-quoted identifiers, standard '' string escaping
# --complete-insert     → INSERT INTO "t" ("c1","c2") VALUES (...) form
# --skip-extended-insert→ one INSERT per row (slower but safe for import)
# --no-create-info      → data only, no DDL
docker exec "$CONTAINER" mysqldump \
    -u "$DB_USER" \
    -p"$DB_PASS" \
    --no-create-info \
    --skip-triggers \
    --complete-insert \
    --compatible=ansi \
    --skip-extended-insert \
    --skip-lock-tables \
    --skip-add-locks \
    --skip-comments \
    --skip-set-charset \
    "$DB_NAME" "${TABLES[@]}" 2>/dev/null > "$RAW_FILE"

# Build the SQLite-compatible output
{
    echo "-- MySQL → SQLite data migration export"
    echo "-- Generated: $(date -u +'%Y-%m-%dT%H:%M:%SZ')"
    echo "-- Source container: $CONTAINER"
    echo "-- Source database:  $DB_NAME"
    echo ""
    echo "PRAGMA foreign_keys = OFF;"
    echo "BEGIN TRANSACTION;"
    echo ""

    # Keep only INSERT statements — strip SET, LOCK, comments, blank lines.
    # Using awk instead of grep to preserve multi-line INSERT statements
    # (values may contain literal newlines, e.g. notification messages).
    awk '/^INSERT INTO /{p=1} p{print} /;$/{p=0}' "$RAW_FILE" || true

    echo ""
    echo "COMMIT;"
    echo "PRAGMA foreign_keys = ON;"
} > "$OUTPUT"

INSERT_COUNT=$(grep -c '^INSERT INTO ' "$OUTPUT" 2>/dev/null || echo "0")

echo ""
echo "Export complete!"
echo "  INSERT statements: $INSERT_COUNT"
echo "  Output file:       $OUTPUT"
echo "  File size:         $(du -h "$OUTPUT" | cut -f1)"
echo ""
echo "Per-table row counts:"

for TABLE in "${TABLES[@]}"; do
    COUNT=$(grep -c "^INSERT INTO \"$TABLE\"" "$OUTPUT" 2>/dev/null || echo "0")
    if [ "$COUNT" -gt 0 ]; then
        printf "  %-42s %s rows\n" "$TABLE" "$COUNT"
    else
        printf "  %-42s (empty)\n" "$TABLE"
    fi
done

echo ""
echo "Done. Next steps:"
echo "  1. Deploy the new sensor-hub version (SQLite-based)"
echo "  2. Start it once so the SQLite schema initialises, then stop it"
echo "  3. Run:  ./import-to-sqlite.sh <path-to-sensor_hub.db> $OUTPUT"
