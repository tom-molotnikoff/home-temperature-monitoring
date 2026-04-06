-- Rollback migration 000006: Generic Sensor Model
-- Reverse all changes, restoring temperature-specific tables and column names.

-- ============================================================
-- 1. Remove new permissions
-- ============================================================
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('view_drivers', 'view_measurement_types')
);
DELETE FROM permissions WHERE name IN ('view_drivers', 'view_measurement_types');

-- ============================================================
-- 2. Recreate temperature_readings from readings data
-- ============================================================
CREATE TABLE IF NOT EXISTS temperature_readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    time TEXT NOT NULL,
    temperature REAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);

CREATE INDEX IF NOT EXISTS idx_time ON temperature_readings (time DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_id ON temperature_readings (sensor_id);

INSERT INTO temperature_readings (sensor_id, time, temperature)
SELECT r.sensor_id, r.time, r.numeric_value
FROM readings r
JOIN measurement_types mt ON r.measurement_type_id = mt.id
WHERE mt.name = 'temperature' AND r.numeric_value IS NOT NULL;

-- ============================================================
-- 3. Recreate hourly_avg_temperature from hourly_averages
-- ============================================================
CREATE TABLE IF NOT EXISTS hourly_avg_temperature (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    time TEXT NOT NULL,
    average_temperature REAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id),
    UNIQUE (sensor_id, time)
);

CREATE INDEX IF NOT EXISTS hourly_idx_time ON hourly_avg_temperature (time DESC);
CREATE INDEX IF NOT EXISTS hourly_idx_sensor_id ON hourly_avg_temperature (sensor_id);

INSERT INTO hourly_avg_temperature (sensor_id, time, average_temperature)
SELECT ha.sensor_id, ha.time, ha.average_value
FROM hourly_averages ha
JOIN measurement_types mt ON ha.measurement_type_id = mt.id
WHERE mt.name = 'temperature';

-- ============================================================
-- 4. Remove measurement_type_id from alert_sent_history
-- ============================================================
ALTER TABLE alert_sent_history DROP COLUMN measurement_type_id;

-- ============================================================
-- 5. Rebuild sensor_alert_rules without measurement_type_id
-- ============================================================
CREATE TABLE sensor_alert_rules_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    high_threshold REAL,
    low_threshold REAL,
    trigger_status TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    rate_limit_hours INTEGER NOT NULL DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE
);

INSERT INTO sensor_alert_rules_old (id, sensor_id, alert_type, high_threshold, low_threshold, trigger_status, enabled, rate_limit_hours, created_at, updated_at)
SELECT id, sensor_id, alert_type, high_threshold, low_threshold, trigger_status, enabled, rate_limit_hours, created_at, updated_at
FROM sensor_alert_rules;

DROP TABLE sensor_alert_rules;
ALTER TABLE sensor_alert_rules_old RENAME TO sensor_alert_rules;

CREATE INDEX IF NOT EXISTS idx_alert_sensor_id ON sensor_alert_rules (sensor_id);
CREATE INDEX IF NOT EXISTS idx_alert_enabled ON sensor_alert_rules (enabled);

-- ============================================================
-- 6. Drop new tables
-- ============================================================
DROP TABLE IF EXISTS hourly_events;
DROP TABLE IF EXISTS hourly_averages;
DROP TABLE IF EXISTS readings;
DROP TABLE IF EXISTS sensor_measurement_types;

-- ============================================================
-- 7. Restore type column on sensors, drop sensor_driver
-- ============================================================
ALTER TABLE sensors ADD COLUMN type TEXT DEFAULT NULL;

-- Best-effort restore: map known driver names back to type values
UPDATE sensors SET type = 'Temperature' WHERE sensor_driver = 'sensor-hub-http-temperature';
UPDATE sensors SET type = sensor_driver WHERE type IS NULL AND sensor_driver IS NOT NULL;

-- Drop index on sensor_driver BEFORE dropping the column
DROP INDEX IF EXISTS idx_sensor_driver;
ALTER TABLE sensors DROP COLUMN sensor_driver;

CREATE INDEX IF NOT EXISTS idx_sensor_type ON sensors (type);

-- ============================================================
-- 8. Drop measurement_types last (other tables reference it)
-- ============================================================
DROP TABLE IF EXISTS measurement_types;
