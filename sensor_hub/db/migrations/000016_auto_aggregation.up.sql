-- Migration 000016: Smart auto-aggregation
-- Adds per-measurement-type aggregation functions, drops pre-computed hourly tables.

-- 1. Aggregation functions per measurement type
CREATE TABLE IF NOT EXISTS measurement_type_aggregations (
    measurement_type_id INTEGER NOT NULL,
    function TEXT NOT NULL CHECK (function IN ('avg', 'count', 'last', 'min', 'max', 'sum')),
    is_default INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (measurement_type_id, function),
    FOREIGN KEY (measurement_type_id) REFERENCES measurement_types(id) ON DELETE CASCADE
);

-- Seed numeric types: default avg
INSERT INTO measurement_type_aggregations (measurement_type_id, function, is_default)
SELECT id, 'avg', 1 FROM measurement_types
WHERE name IN ('temperature','humidity','pressure','power','battery','voltage',
               'luminance','link_quality','illuminance','energy','current','co2',
               'voc','formaldehyde','pm25','soil_moisture','energy_today',
               'energy_month','energy_yesterday');

-- Seed binary types: default count, also last
INSERT INTO measurement_type_aggregations (measurement_type_id, function, is_default)
SELECT id, 'count', 1 FROM measurement_types
WHERE name IN ('motion','contact','doorbell','occupancy','water_leak','smoke',
               'carbon_monoxide','tamper','vibration','state','battery_low');

INSERT INTO measurement_type_aggregations (measurement_type_id, function, is_default)
SELECT id, 'last', 0 FROM measurement_types
WHERE name IN ('motion','contact','doorbell','occupancy','water_leak','smoke',
               'carbon_monoxide','tamper','vibration','state','battery_low');

-- 2. Drop pre-computed hourly tables (no longer needed)
DROP TABLE IF EXISTS hourly_averages;
DROP TABLE IF EXISTS hourly_events;

-- 3. Index for efficient query-time aggregation
CREATE INDEX IF NOT EXISTS idx_readings_time_asc ON readings (time ASC);
