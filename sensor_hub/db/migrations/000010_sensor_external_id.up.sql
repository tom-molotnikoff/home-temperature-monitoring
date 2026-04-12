-- Add external_id column for stable MQTT device identification.
-- This decouples the user-facing name from the MQTT device identity,
-- allowing users to rename sensors without breaking MQTT matching.
ALTER TABLE sensors ADD COLUMN external_id TEXT;

-- Backfill: push-based sensors use their current name as external_id
UPDATE sensors SET external_id = name WHERE LOWER(sensor_driver) = 'zigbee2mqtt';

-- Unique index (SQLite allows multiple NULLs with partial index)
CREATE UNIQUE INDEX IF NOT EXISTS idx_sensors_external_id ON sensors(external_id) WHERE external_id IS NOT NULL;
