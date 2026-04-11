DROP INDEX IF EXISTS idx_sensors_external_id;

-- SQLite does not support DROP COLUMN before 3.35.0, so recreate the table
CREATE TABLE sensors_backup AS SELECT id, name, sensor_driver, config, health_status, health_reason, enabled, status FROM sensors;
DROP TABLE sensors;
ALTER TABLE sensors_backup RENAME TO sensors;
