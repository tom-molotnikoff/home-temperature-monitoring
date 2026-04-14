-- Restore hourly tables (empty — pre-computed data cannot be recovered)
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

DROP TABLE IF EXISTS measurement_type_aggregations;
DROP INDEX IF EXISTS idx_readings_time_asc;
