-- Migration 000006: Generic Sensor Model
-- Evolves schema from temperature-only to arbitrary sensor types via SensorDriver architecture.

-- ============================================================
-- 1. Measurement types reference table
-- ============================================================
CREATE TABLE IF NOT EXISTS measurement_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('numeric', 'binary')),
    default_unit TEXT NOT NULL DEFAULT ''
);

INSERT OR IGNORE INTO measurement_types (name, display_name, category, default_unit) VALUES
    ('temperature', 'Temperature', 'numeric', '°C'),
    ('humidity', 'Humidity', 'numeric', '%'),
    ('pressure', 'Pressure', 'numeric', 'hPa'),
    ('power', 'Power', 'numeric', 'W'),
    ('battery', 'Battery', 'numeric', '%'),
    ('voltage', 'Voltage', 'numeric', 'V'),
    ('luminance', 'Luminance', 'numeric', 'lx'),
    ('motion', 'Motion', 'binary', ''),
    ('contact', 'Contact', 'binary', ''),
    ('doorbell', 'Doorbell', 'binary', '');

-- ============================================================
-- 2. Add sensor_driver column to sensors, migrate type data, drop type
-- ============================================================
ALTER TABLE sensors ADD COLUMN sensor_driver TEXT DEFAULT NULL;

-- Migrate: existing sensors with type='Temperature' get the HTTP temperature driver
UPDATE sensors SET sensor_driver = 'sensor-hub-http-temperature' WHERE LOWER(type) = 'temperature';
-- Any remaining sensors get their type value as-is (lowercase)
UPDATE sensors SET sensor_driver = LOWER(type) WHERE sensor_driver IS NULL AND type IS NOT NULL;

-- Drop the old index on type BEFORE dropping the column (SQLite requires this)
DROP INDEX IF EXISTS idx_sensor_type;

-- Drop the old type column (SQLite >= 3.35.0)
ALTER TABLE sensors DROP COLUMN type;

CREATE INDEX IF NOT EXISTS idx_sensor_driver ON sensors (sensor_driver);

-- ============================================================
-- 3. Sensor ↔ Measurement Type join table
-- ============================================================
CREATE TABLE IF NOT EXISTS sensor_measurement_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    measurement_type_id INTEGER NOT NULL,
    unit TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    FOREIGN KEY (measurement_type_id) REFERENCES measurement_types(id) ON DELETE CASCADE,
    UNIQUE (sensor_id, measurement_type_id)
);

-- Back-fill: all existing sensors get the temperature measurement type
INSERT OR IGNORE INTO sensor_measurement_types (sensor_id, measurement_type_id, unit)
SELECT s.id, mt.id, mt.default_unit
FROM sensors s, measurement_types mt
WHERE mt.name = 'temperature';

-- ============================================================
-- 4. Unified readings table + migrate from temperature_readings
-- ============================================================
CREATE TABLE IF NOT EXISTS readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    measurement_type_id INTEGER NOT NULL,
    numeric_value REAL,
    text_state TEXT,
    time TEXT NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    FOREIGN KEY (measurement_type_id) REFERENCES measurement_types(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_readings_time ON readings (time DESC);
CREATE INDEX IF NOT EXISTS idx_readings_sensor_id ON readings (sensor_id);
CREATE INDEX IF NOT EXISTS idx_readings_sensor_type_time ON readings (sensor_id, measurement_type_id, time DESC);

-- Migrate existing temperature readings
INSERT INTO readings (sensor_id, measurement_type_id, numeric_value, time)
SELECT tr.sensor_id, mt.id, tr.temperature, tr.time
FROM temperature_readings tr, measurement_types mt
WHERE mt.name = 'temperature';

-- ============================================================
-- 5. Unified hourly averages table + migrate from hourly_avg_temperature
-- ============================================================
CREATE TABLE IF NOT EXISTS hourly_averages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    measurement_type_id INTEGER NOT NULL,
    time TEXT NOT NULL,
    average_value REAL NOT NULL,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    FOREIGN KEY (measurement_type_id) REFERENCES measurement_types(id) ON DELETE CASCADE,
    UNIQUE (sensor_id, measurement_type_id, time)
);

CREATE INDEX IF NOT EXISTS idx_hourly_averages_time ON hourly_averages (time DESC);
CREATE INDEX IF NOT EXISTS idx_hourly_averages_sensor_id ON hourly_averages (sensor_id);
CREATE INDEX IF NOT EXISTS idx_hourly_averages_sensor_type_time ON hourly_averages (sensor_id, measurement_type_id, time DESC);

-- Migrate existing hourly averages
INSERT INTO hourly_averages (sensor_id, measurement_type_id, time, average_value)
SELECT hat.sensor_id, mt.id, hat.time, hat.average_temperature
FROM hourly_avg_temperature hat, measurement_types mt
WHERE mt.name = 'temperature';

-- ============================================================
-- 6. Hourly events table (for binary measurement types)
-- ============================================================
CREATE TABLE IF NOT EXISTS hourly_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    measurement_type_id INTEGER NOT NULL,
    time TEXT NOT NULL,
    event_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    FOREIGN KEY (measurement_type_id) REFERENCES measurement_types(id) ON DELETE CASCADE,
    UNIQUE (sensor_id, measurement_type_id, time)
);

CREATE INDEX IF NOT EXISTS idx_hourly_events_time ON hourly_events (time DESC);
CREATE INDEX IF NOT EXISTS idx_hourly_events_sensor_type_time ON hourly_events (sensor_id, measurement_type_id, time DESC);

-- ============================================================
-- 7. Rebuild sensor_alert_rules with measurement_type_id
--    SQLite cannot ADD columns with UNIQUE constraints, so we rebuild.
-- ============================================================
CREATE TABLE sensor_alert_rules_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    measurement_type_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    high_threshold REAL,
    low_threshold REAL,
    trigger_status TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    rate_limit_hours INTEGER NOT NULL DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    FOREIGN KEY (measurement_type_id) REFERENCES measurement_types(id) ON DELETE CASCADE,
    UNIQUE (sensor_id, measurement_type_id, alert_type)
);

-- Migrate existing rules, assigning temperature as the measurement type
INSERT INTO sensor_alert_rules_new (id, sensor_id, measurement_type_id, alert_type, high_threshold, low_threshold, trigger_status, enabled, rate_limit_hours, created_at, updated_at)
SELECT sar.id, sar.sensor_id, mt.id, sar.alert_type, sar.high_threshold, sar.low_threshold, sar.trigger_status, sar.enabled, sar.rate_limit_hours, sar.created_at, sar.updated_at
FROM sensor_alert_rules sar, measurement_types mt
WHERE mt.name = 'temperature';

DROP TABLE sensor_alert_rules;
ALTER TABLE sensor_alert_rules_new RENAME TO sensor_alert_rules;

CREATE INDEX IF NOT EXISTS idx_alert_sensor_id ON sensor_alert_rules (sensor_id);
CREATE INDEX IF NOT EXISTS idx_alert_enabled ON sensor_alert_rules (enabled);
CREATE INDEX IF NOT EXISTS idx_alert_sensor_mt ON sensor_alert_rules (sensor_id, measurement_type_id);

-- ============================================================
-- 8. Add measurement_type_id to alert_sent_history
-- ============================================================
ALTER TABLE alert_sent_history ADD COLUMN measurement_type_id INTEGER REFERENCES measurement_types(id);

-- Back-fill existing history with temperature type
UPDATE alert_sent_history SET measurement_type_id = (SELECT id FROM measurement_types WHERE name = 'temperature');

-- ============================================================
-- 9. Drop old temperature-specific tables
-- ============================================================
DROP TABLE IF EXISTS temperature_readings;
DROP TABLE IF EXISTS hourly_avg_temperature;

-- ============================================================
-- 10. New permissions for drivers and measurement types
-- ============================================================
INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('view_drivers', 'View available sensor drivers'),
    ('view_measurement_types', 'View measurement types');

-- Grant to all roles
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('view_drivers', 'view_measurement_types');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('view_drivers', 'view_measurement_types');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name IN ('view_drivers', 'view_measurement_types');
